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
	archiveSuffix = ".tar.gz"
)

var _ io.WriteCloser = (*Writer)(nil)

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
	created  time.Time // file creation date
	creates  []byte    // file creation date for compare
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
	now := time.Now().In(w.loc)
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
// Guarded so concurrent day-crossing writes and SetCompress calls cannot
// run two sweeps at once.
func (w *Writer) sweep() {
	if !atomic.CompareAndSwapInt32(&w.sweeping, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&w.sweeping, 0)

	w.mu.Lock()
	doCompress := w.compress
	maxage := w.maxage
	w.mu.Unlock()
	if doCompress {
		w.archive()
	}
	w.delete(maxage)
}

// compressDays returns how many recent days stay uncompressed:
// min(7, max(1, maxage/2)). Today and yesterday are always kept as plain
// files regardless of the threshold.
func (w *Writer) compressDays() int {
	d := w.maxage / 2
	if d < 1 {
		d = 1
	}
	if d > 7 {
		d = 7
	}
	return d
}

// archive groups rotated backups by day and archives every full day older
// than the compress threshold into <fname>.<YYYY-MM-DD>.tar.gz. Entries keep
// the original backup filenames so archives extract to greppable .log files.
// Archiving is atomic (temp file + rename) and idempotent (existing archives
// are skipped); sources are removed only after a successful archive.
func (w *Writer) archive() {
	keep := w.compressDays()
	now := time.Now().In(w.loc)
	limit := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, -keep)

	dirs, err := os.ReadDir(w.fdir)
	if err != nil {
		return
	}
	days := make(map[string][]string)
	for _, e := range dirs {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		t, err := w.name2time(name)
		if err != nil {
			continue
		}
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, w.loc)
		if !day.Before(limit) {
			continue // within the uncompressed window
		}
		key := t.Format("2006-01-02")
		days[key] = append(days[key], name)
	}
	if len(days) == 0 {
		return
	}
	keys := make([]string, 0, len(days))
	for k := range days {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		w.archiveDay(k, days[k])
	}
}

// archiveDay writes one tar.gz containing the given backup files and removes
// the sources on success.
func (w *Writer) archiveDay(day string, names []string) {
	apath := filepath.Join(w.fdir, w.fname+"."+day+archiveSuffix)
	if _, err := os.Stat(apath); err == nil {
		return // already archived
	}
	tmp := apath + ".tmp"
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
			writeErr = copyStream(tw, src)
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

func copyStream(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

// delete removes log backups and archives older than maxage days.
func (w *Writer) delete(maxage int) {
	if maxage <= 0 {
		return
	}
	dir := w.fdir
	fakeNow := time.Now().In(w.loc).AddDate(0, 0, -maxage)
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, path := range dirs {
		name := path.Name()
		if path.IsDir() {
			continue
		}
		t, err := w.name2time(name)
		if err != nil {
			t, err = w.archive2time(name)
		}
		// only delete files matching the date pattern
		if err == nil && t.Before(fakeNow) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func (w *Writer) name2time(name string) (time.Time, error) {
	// Backup filename: <fname>.<date><fsuffix>, e.g. app.2024-01-02-030405.log
	// Zero-allocation: slicing shares the backing array.
	if !strings.HasPrefix(name, w.fname+".") || !strings.HasSuffix(name, w.fsuffix) ||
		len(name) != len(w.fname)+len(".2006-01-02-150405")+len(w.fsuffix) {
		return time.Time{}, errors.New("bad format")
	}
	return time.ParseInLocation(".2006-01-02-150405", name[len(w.fname):len(name)-len(w.fsuffix)], w.loc)
}

// archive2time parses archive names: <fname>.<YYYY-MM-DD>.tar.gz
func (w *Writer) archive2time(name string) (time.Time, error) {
	if !strings.HasPrefix(name, w.fname+".") || !strings.HasSuffix(name, archiveSuffix) ||
		len(name) != len(w.fname)+1+len("2006-01-02")+len(archiveSuffix) {
		return time.Time{}, errors.New("bad format")
	}
	mid := name[len(w.fname)+1 : len(name)-len(archiveSuffix)]
	return time.ParseInLocation("2006-01-02", mid, w.loc)
}

func (w *Writer) time2name(t time.Time) string {
	return t.Format(".2006-01-02-150405")
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
