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
	defMaxage  = 64 //天
	defMaxsize = 64 //MiB
)

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	maxage  int       // 最大保留天数
	maxsize int64     // 单个日志最大容量 默认 64MB
	size    int64     // 累计大小
	cons    bool      // 标准输出  默认false
	fpath   string    // 文件目录 完整路径 fpath=fdir+fname+fsuffix
	fdir    string    //
	fname   string    // 文件名
	fsuffix string    // 文件后缀名 默认 .log
	created time.Time // 文件创建日期
	creates []byte    // 文件创建日期 for compare
	file    *os.File
	bw      *bufio.Writer
	tk      *time.Ticker
	mu      sync.Mutex
	done    chan struct{}
	closed  int32 // 0 = open, 1 = closed
}

func New(path string, cons ...bool) *Writer {
	consv := false
	if len(cons) > 0 {
		consv = cons[0]
	}
	w := &Writer{
		fpath: path, //dir1/dir2/app.log
		mu:    sync.Mutex{},
		cons:  consv,
		done:  make(chan struct{}),
	}
	w.fdir = filepath.Dir(w.fpath)                                  //dir1/dir2
	w.fsuffix = filepath.Ext(w.fpath)                               //.log
	w.fname = strings.TrimSuffix(filepath.Base(w.fpath), w.fsuffix) //app
	if w.fsuffix == "" {
		w.fsuffix = ".log"
	}
	w.maxsize = sizeMiB * defMaxsize
	w.maxage = defMaxage
	os.MkdirAll(filepath.Dir(w.fpath), 0755)
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

// SetMaxAge 最大保留天数
func (w *Writer) SetMaxAge(ma int) {
	w.mu.Lock()
	w.maxage = ma
	w.mu.Unlock()
}

// SetMaxSize 单个日志最大容量MiB
func (w *Writer) SetMaxSize(ms int64) {
	if ms < 1 {
		return
	}
	w.mu.Lock()
	w.maxsize = ms * sizeMiB
	w.mu.Unlock()
}

// SetCons 同时输出控制台
func (w *Writer) SetCons(b bool) {
	w.mu.Lock()
	w.cons = b
	w.mu.Unlock()
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
	if w.cons {
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
	// 按天切割
	if !w.equaldate(w.creates, p) { //2023-04-05
		go w.delete() // 每天检测一次旧文件
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// 按大小切割
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

// rotate 切割文件
func (w *Writer) rotate() error {
	now := time.Now()
	if w.file != nil {
		w.bw.Flush()
		w.file.Sync()
		w.file.Close()
		// 保存
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
	fout, err := os.OpenFile(w.fpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	w.file = fout
	w.bw = bufio.NewWriter(w.file)
	return nil
}

// 删除旧日志
func (w *Writer) delete() {
	w.mu.Lock()
	maxage := w.maxage
	w.mu.Unlock()
	if maxage <= 0 {
		return
	}
	dir := filepath.Dir(w.fpath)
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
		// 只删除满足格式的文件
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
