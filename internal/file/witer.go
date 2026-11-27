package file

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	sizeMiB       = 1024 * 1024
	defMaxage     = 64 // days
	defMaxsize    = 64 // MiB
	keepDays      = 7  // calendar days kept uncompressed (today included)
	archiveSuffix = ".tar.gz"
	nameLayout    = ".2006-01-02-150405"
)

var _ io.WriteCloser = (*Writer)(nil)

// errBadName reports a file that is not managed by this writer.
var errBadName = errors.New("bad format")

// nameKind distinguishes rotated backups from per-day archives. They share
// one filename layout and differ only by suffix.
type nameKind int

const (
	kindBackup  nameKind = iota // <fname>.<date><fsuffix>
	kindArchive                 // <fname>.<date>.tar.gz
)

type Writer struct {
	maxage   int    // max retention days
	maxsize  int64  // max size per file, default 64 MiB
	compress bool   // archive old backups into per-day tar.gz
	console  bool   // mirror to stderr
	size     int64  // accumulated size
	fpath    string // full path fpath=fdir+fname+fsuffix
	fdir     string // directory
	fname    string // filename
	fsuffix  string // suffix, default .log
	loc      *time.Location
	now      func() time.Time // clock, test hook; nil means time.Now
	created  time.Time        // file creation date
	creates  []byte           // file creation date for compare
	file     *os.File
	bw       *bufio.Writer
	tk       *time.Ticker
	mu       sync.Mutex
	done     chan struct{}
	closed   int32 // 0 = open, 1 = closed
	sweeping int32 // 0 = idle, 1 = sweep in progress
}

func New(path string) *Writer {
	w := &Writer{
		fpath:   path, //dir1/dir2/app.log
		mu:      sync.Mutex{},
		console: true,
		loc:     time.Local,
		now:     time.Now,
		done:    make(chan struct{}),
	}
	w.fdir = filepath.Dir(w.fpath)                                  //dir1/dir2
	w.fsuffix = filepath.Ext(w.fpath)                               //.log
	w.fname = strings.TrimSuffix(filepath.Base(w.fpath), w.fsuffix) //app
	if w.fsuffix == "" {
		w.fsuffix = ".log"
	}
	w.maxsize = sizeMiB * defMaxsize
	w.maxage = defMaxage
	os.MkdirAll(w.fdir, 0755)
	w.tk = time.NewTicker(time.Second * 5)
	go w.daemon()
	return w
}

func (w *Writer) daemon() {
	for {
		select {
		case <-w.tk.C:
			w.flush()
		case <-w.done:
			return
		}
	}
}

// SetMaxAge sets the max retention days.
func (w *Writer) SetMaxAge(ma int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.maxage = ma
}

// SetMaxSize sets the max size of a single log file in MiB.
func (w *Writer) SetMaxSize(ms int64) {
	if ms < 1 {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.maxsize = ms * sizeMiB
}

// SetConsole sets whether to also output to stderr.
func (w *Writer) SetConsole(b bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.console = b
}

// SetCompress enables or disables per-day tar.gz archiving of rotated
// backups. It only toggles the flag; the sweep runs on the next
// day-crossing write.
func (w *Writer) SetCompress(b bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.compress = b
}

func (w *Writer) Write(p []byte) (n int, err error) {
	// Write to stderr before acquiring the file lock.
	if w.console {
		os.Stderr.Write(p)
	}
	// Fast path: if already closed, skip locking entirely.
	if atomic.LoadInt32(&w.closed) != 0 {
		return 0, os.ErrClosed
	}
	w.mu.Lock()
	if w.file == nil {
		if err := w.rotate(); err != nil {
			w.mu.Unlock()
			os.Stderr.Write(p)
			return 0, err
		}
	}
	// rotate by day — inline date comparison (hot path: date changes once/day)
	if len(w.creates) >= 10 && len(p) >= 15 && !bytes.Equal(w.creates[:10], p[5:15]) {
		go w.sweep() // daily compress + cleanup
		if err := w.rotate(); err != nil {
			w.mu.Unlock()
			return 0, err
		}
	}
	// rotate by size
	if w.size+int64(len(p)) >= w.maxsize {
		if err := w.rotate(); err != nil {
			w.mu.Unlock()
			return 0, err
		}
	}
	n, err = w.bw.Write(p)
	w.size += int64(n)
	w.mu.Unlock()
	return
}

// rotate closes the current file, saves a backup, and opens a new one.
func (w *Writer) rotate() error {
	now := w.clock()
	if w.file != nil {
		w.bw.Flush()
		w.file.Sync()
		w.file.Close()
		// save backup
		fbak := w.fname + w.time2name(w.created) + w.fsuffix
		os.Rename(w.fpath, filepath.Join(w.fdir, fbak))
		w.size = 0
		w.created = now
	} else if finfo, err := os.Stat(w.fpath); err == nil {
		// First rotation: inherit size and modtime from existing file.
		w.size = finfo.Size()
		w.created = finfo.ModTime().In(w.loc)
	} else {
		w.created = now
	}
	w.creates = w.created.AppendFormat(nil, time.RFC3339)
	os.MkdirAll(w.fdir, 0755)
	fout, err := os.OpenFile(w.fpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = fout
	w.bw = bufio.NewWriterSize(w.file, 32*1024)
	return nil
}

// sweep compresses eligible backups and removes files beyond retention.
// Guarded so concurrent day-crossing writes cannot run two sweeps at once.
func (w *Writer) sweep() {
	if !atomic.CompareAndSwapInt32(&w.sweeping, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&w.sweeping, 0)

	w.mu.Lock()
	compress := w.compress
	maxage := w.maxage
	w.mu.Unlock()
	// Archiving only pays off when backups outlive the uncompressed window:
	// with maxage <= keepDays retention drops them as soon as they are packed.
	if compress && maxage > keepDays {
		w.archive()
	}
	w.delete(maxage)
}

// clock returns the current time in the writer's location. Tests replace now
// so day boundaries stay deterministic.
func (w *Writer) clock() time.Time {
	now := time.Now()
	if w.now != nil {
		now = w.now()
	}
	if w.loc != nil {
		now = now.In(w.loc)
	}
	return now
}

// dayof truncates t to the start of its calendar day.
func (w *Writer) dayof(t time.Time) time.Time {
	loc := w.loc
	if loc == nil {
		loc = t.Location()
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// archive groups rotated backups by day and archives every closed day older
// than keepDays into <fname>.<YYYY-MM-DD>-<hhmmss>.tar.gz, where the date is
// the archived day and the time of day is the moment the sweep runs. Entries
// keep the original backup filenames so archives extract to greppable .log
// files. Archiving is atomic (temp file + rename) and keeps one archive per
// day: an existing archive is never rewritten, so sources are removed only
// after a successful archive.
func (w *Writer) archive() {
	now := w.clock()
	limit := w.dayof(now).AddDate(0, 0, -(keepDays - 1))

	dirs, err := os.ReadDir(w.fdir)
	if err != nil {
		return
	}
	archived := make(map[time.Time]struct{})
	temps := make(map[time.Time]string)
	days := make(map[time.Time][]string)
	for _, e := range dirs {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".tmp") {
			if t, _, err := w.pasename(strings.TrimSuffix(name, ".tmp")); err == nil {
				temps[w.dayof(t)] = name
			}
			continue
		}
		t, kind, err := w.pasename(name)
		if err != nil {
			continue
		}
		day := w.dayof(t)
		if kind == kindArchive {
			archived[day] = struct{}{}
			continue
		}
		if !day.Before(limit) {
			continue // within the uncompressed window
		}
		days[day] = append(days[day], name)
	}
	// Drop temp files left behind for days that are already archived.
	for day, name := range temps {
		if _, ok := archived[day]; ok {
			os.Remove(filepath.Join(w.fdir, name))
		}
	}
	if len(days) == 0 {
		return
	}
	keys := make([]time.Time, 0, len(days))
	for k := range days {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].Before(keys[j]) })
	for _, day := range keys {
		if _, ok := archived[day]; ok {
			// One archive per day: a late backup for an archived day stays
			// plain until retention drops it.
			continue
		}
		w.archiveday(day, now, days[day])
	}
}

// archiveday writes one tar.gz containing the given backup files and removes
// the sources on success.
func (w *Writer) archiveday(day, now time.Time, names []string) {
	stamp := time.Date(day.Year(), day.Month(), day.Day(), now.Hour(), now.Minute(), now.Second(), 0, w.dayof(day).Location())
	apath := filepath.Join(w.fdir, w.fname+w.time2name(stamp)+archiveSuffix)
	if _, err := os.Stat(apath); err == nil {
		os.Remove(w.tmppath(day))
		return // already archived
	}
	tmp := w.tmppath(day)
	fout, err := os.Create(tmp)
	if err != nil {
		return
	}
	gz := gzip.NewWriter(fout)
	tw := tar.NewWriter(gz)

	ok := true
	for _, name := range names {
		src, err := os.Open(filepath.Join(w.fdir, name))
		if err != nil {
			ok = false
			break
		}
		finfo, err := src.Stat()
		if err != nil {
			src.Close()
			ok = false
			break
		}
		hdr, err := tar.FileInfoHeader(finfo, "")
		if err != nil {
			src.Close()
			ok = false
			break
		}
		hdr.Name = name // keep the full original name for grep after extract
		writeErr := tw.WriteHeader(hdr)
		if writeErr == nil {
			_, writeErr = io.Copy(tw, src)
		}
		closeErr := src.Close()
		if writeErr != nil || closeErr != nil {
			ok = false
			break
		}
	}
	if tw.Close() != nil {
		ok = false
	}
	if gz.Close() != nil {
		ok = false
	}
	if fout.Sync() != nil {
		ok = false
	}
	if fout.Close() != nil {
		ok = false
	}
	if !ok {
		os.Remove(tmp)
		return
	}
	if os.Rename(tmp, apath) != nil {
		os.Remove(tmp)
		return
	}
	for _, name := range names {
		os.Remove(filepath.Join(w.fdir, name))
	}
}

// tmppath returns a stable temp name for a day so retries reuse one file
// instead of accumulating leftovers. It keeps the managed layout so the sweep
// can map a leftover temp file back to its day.
func (w *Writer) tmppath(day time.Time) string {
	return filepath.Join(w.fdir, w.fname+w.time2name(w.dayof(day))+archiveSuffix+".tmp")
}

// delete removes managed files once their whole calendar day falls outside the
// retention window. Backups and archives share one rule: a file goes away when
// the day it belongs to has ended at or before now-maxage. Removal therefore
// trails the exact timestamp by up to one day and never runs early.
func (w *Writer) delete(maxage int) {
	if maxage <= 0 {
		return
	}
	dir := w.fdir
	fakeNow := w.clock().AddDate(0, 0, -maxage)
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, path := range dirs {
		name := path.Name()
		if path.IsDir() {
			continue
		}
		t, _, err := w.pasename(name)
		if err != nil {
			continue
		}
		if !w.dayof(t).AddDate(0, 0, 1).After(fakeNow) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// pasename parses a managed filename:
//
//	<fname>.<YYYY-MM-DD-hhmmss><fsuffix>   rotated backup
//	<fname>.<YYYY-MM-DD-hhmmss>.tar.gz     per-day archive
//
// The date must follow the prefix and the suffix decides the kind, so one
// parser covers both. Zero-allocation: slicing shares the backing array.
func (w *Writer) pasename(name string) (time.Time, nameKind, error) {
	if !strings.HasPrefix(name, w.fname+".") {
		return time.Time{}, kindBackup, errBadName
	}
	rest := name[len(w.fname):] // ".2026-01-02-030405<kind suffix>"
	kind := kindBackup
	switch {
	case strings.HasSuffix(rest, archiveSuffix):
		rest = strings.TrimSuffix(rest, archiveSuffix)
		kind = kindArchive
	case w.fsuffix != "" && strings.HasSuffix(rest, w.fsuffix):
		rest = strings.TrimSuffix(rest, w.fsuffix)
	default:
		return time.Time{}, kindBackup, errBadName
	}
	loc := w.loc
	if loc == nil {
		loc = time.Local
	}
	t, err := time.ParseInLocation(nameLayout, rest, loc)
	if err != nil {
		return time.Time{}, kindBackup, err
	}
	return t, kind, nil
}

func (w *Writer) time2name(t time.Time) string {
	return t.Format(nameLayout)
}

func (w *Writer) flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.bw == nil {
		return nil
	}
	return w.bw.Flush()
}

func (w *Writer) Close() error {
	if !atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		return nil
	}
	w.tk.Stop()
	close(w.done)
	w.flush()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	w.file.Sync()
	err := w.file.Close()
	w.file = nil
	w.bw = nil
	return err
}
