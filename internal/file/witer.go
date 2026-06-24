package file

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	sizeMiB    = 1024 * 1024
	defMaxage  = 64 // days
	defMaxsize = 64 // MiB
)

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	maxage  int       // max retention days
	maxsize int64     // max size per file, default 64 MiB
	console bool      // mirror to stderr
	size    int64     // accumulated size
	fpath   string    // full path fpath=fdir+fname+fsuffix
	fdir    string    // directory
	fname   string    // filename
	fsuffix string    // suffix, default .log
	created time.Time // file creation date
	creates []byte    // file creation date for compare
	file    *os.File
	bw      *bufio.Writer
	tk      *time.Ticker
	mu      sync.Mutex
	done    chan struct{}
	closed  int32 // 0 = open, 1 = closed
}

func New(path string, cons bool) *Writer {
	w := &Writer{
		fpath:   path, //dir1/dir2/app.log
		mu:      sync.Mutex{},
		console: cons,
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

func (w *Writer) equaldate(file []byte, msg []byte) bool {
	// Only supports zxysilent/logs
	if len(file) < 10 || len(msg) < 15 {
		return true
	}
	return bytes.Equal(file[:10], msg[5:15])
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.console {
		os.Stderr.Write(p)
	}
	if atomic.LoadInt32(&w.closed) != 0 {
		return 0, os.ErrClosed
	}
	if w.file == nil {
		if err := w.rotate(); err != nil {
			os.Stderr.Write(p)
			return 0, err
		}
	}
	// rotate by day
	if !w.equaldate(w.creates, p) { //2023-04-05
		go w.delete(w.maxage) // daily cleanup
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// rotate by size
	if w.size+int64(len(p)) >= w.maxsize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// n, err = w.file.Write(p)
	n, err = w.bw.Write(p)
	w.size += int64(n)
	if err != nil {
		return n, err
	}
	return
}

// rotate closes the current file and opens a new one.
func (w *Writer) rotate() error {
	now := time.Now()
	if w.file != nil {
		w.bw.Flush()
		w.file.Sync()
		w.file.Close()
		// save backup
		fbak := w.fname + w.time2name(w.created) + w.fsuffix
		os.Rename(w.fpath, filepath.Join(w.fdir, fbak))
		w.size = 0
	}
	finfo, err := os.Stat(w.fpath)
	w.created = now
	if err == nil {
		w.size = finfo.Size()
		w.created = finfo.ModTime()
	}
	w.creates = w.created.AppendFormat(nil, time.RFC3339)
	os.MkdirAll(w.fdir, 0755)
	fout, err := os.OpenFile(w.fpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = fout
	w.bw = bufio.NewWriter(w.file)
	return nil
}

// delete removes log files older than maxage days.
func (w *Writer) delete(maxage int) {
	if maxage <= 0 {
		return
	}
	dir := w.fdir
	fakeNow := time.Now().AddDate(0, 0, -maxage)
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
		// only delete files matching the date pattern
		if err == nil && t.Before(fakeNow) {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

func (w *Writer) name2time(name string) (time.Time, error) {
	name = strings.TrimPrefix(name, filepath.Base(w.fname))
	name = strings.TrimSuffix(name, w.fsuffix)
	return time.Parse(".2006-01-02-150405", name)
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
