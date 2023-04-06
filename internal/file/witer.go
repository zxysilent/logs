package file

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	sizeMiB    = 1024 * 1024
	defMaxAge  = 31
	defMaxSize = 64 //MiB
)

var _ io.WriteCloser = (*Writer)(nil)

type Writer struct {
	maxAge  int       // 最大保留天数
	maxSize int64     // 单个日志最大容量 默认 64MB
	size    int64     // 累计大小
	fpath   string    // 文件目录 完整路径 fpath=fdir+fname+fsuffix
	fdir    string    //
	fname   string    // 文件名
	fsuffix string    // 文件后缀名 默认 .log
	created time.Time // 文件创建日期
	creates []byte    // 文件创建日期
	cons    bool      // 标准输出  默认 false
	file    *os.File
	mu      sync.Mutex
}

func New(path string) *Writer {
	w := &Writer{
		fpath: path, //dir1/dir2/app.log
		mu:    sync.Mutex{},
	}
	w.fdir = filepath.Dir(w.fpath)                                  //dir1/dir2
	w.fsuffix = filepath.Ext(w.fpath)                               //.log
	w.fname = strings.TrimSuffix(filepath.Base(w.fpath), w.fsuffix) //app
	if w.fsuffix == "" {
		w.fsuffix = ".log"
	}
	w.maxSize = sizeMiB * defMaxSize
	w.maxAge = defMaxAge
	os.MkdirAll(filepath.Dir(w.fpath), 0755)
	return w
}

// SetMaxAge 最大保留天数
func (w *Writer) SetMaxAge(ma int) {
	w.mu.Lock()
	defer w.mu.Lock()
	w.maxAge = ma
}

// SetMaxSize 单个日志最大容量
func (w *Writer) SetMaxSize(ms int64) {
	if ms < 1 {
		return
	}
	w.mu.Lock()
	defer w.mu.Lock()
	w.maxSize = ms
}

// SetCons 同时输出控制台
func (w *Writer) SetCons(b bool) {
	w.mu.Lock()
	defer w.mu.Lock()
	w.cons = b
}

func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.cons {
		os.Stderr.Write(p)
	}
	if w.file == nil {
		if err := w.rotate(); err != nil {
			os.Stderr.Write(p)
			return 0, err
		}
	}
	// 按天切割
	if !bytes.Equal(w.creates[:10], p[9:19]) { //2023-04-05
		go w.delete() // 每天检测一次旧文件
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	// 按大小切割
	if w.size+int64(len(p)) >= w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err = w.file.Write(p)
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
	return nil
}

// 删除旧日志
func (w *Writer) delete() {
	if w.maxAge <= 0 {
		return
	}
	dir := filepath.Dir(w.fpath)
	fakeNow := time.Now().AddDate(0, 0, -w.maxAge)
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

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.close()
}

// close closes the file if it is open.
func (w *Writer) close() error {
	if w.file == nil {
		return nil
	}
	w.file.Sync()
	err := w.file.Close()
	w.file = nil
	return err
}
