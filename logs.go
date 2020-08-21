package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 日志等级
type logLevel int

const (
	DEBUG logLevel = iota
	INFO
	WARN
	ERROR
	FATAL
	maxSize       = 1024 * 1024 * 256 // 256 MB
	bufferSize    = 1024 * 256        // 256 KB
	digits        = "0123456789"
	flushInterval = 5 * time.Second
	logShort      = "[D][I][W][E][F]"
)

// 字符串等级
func (lv logLevel) Str() string {
	if lv >= DEBUG && lv <= FATAL {
		return logShort[lv*3 : lv*3+3]
	}
	return "[N]"
}

// logger
type FishLogger struct {
	cons     bool          // 标准输出  默认 false
	callInfo bool          // 是否输出行号和文件名 默认 false
	maxAge   int           // 最大天数
	maxSize  int64         // 单个日志最大容量 默认 256MB
	size     int64         // 累计大小
	lpath    string        // 文件目录 完整路径 lpath=lname+lsuffix
	lname    string        // 文件名
	lsuffix  string        // 文件后缀名 默认 .log
	created  string        // 文件创建日期
	level    logLevel      // 输出的日志等级
	list     *buffer       // 缓存
	listLock sync.Mutex    // 链表🔒
	lock     sync.Mutex    // logger🔒
	writer   *bufio.Writer // 缓存io 缓存到文件
	file     *os.File      // 日志文件
}

// 默认实例
var fish = NewLogger("logs/app.log")

// NewLogger 实例化logger
// path 日志完整路径 eg:logs/app.log
func NewLogger(lpath string) *FishLogger {
	fl := new(FishLogger)
	fl.lpath = lpath                                 // logs/app.log
	fl.lsuffix = filepath.Ext(lpath)                 // .log
	fl.lname = strings.TrimSuffix(lpath, fl.lsuffix) // logs/app
	if fl.lsuffix == "" {
		fl.lsuffix = ".log"
	}
	os.MkdirAll(filepath.Dir(lpath), 0666)
	fl.level = DEBUG
	fl.maxSize = maxSize
	go fl.daemon()
	return fl
}

// 设置实例等级
func SetLevel(lv logLevel) {
	fish.SetLevel(lv)
}

// 设置输出等级
func (fl *FishLogger) SetLevel(lv logLevel) {
	if lv < DEBUG || lv > FATAL {
		panic("非法的日志等级")
	}
	fl.lock.Lock()
	fl.level = lv
	fl.lock.Unlock()
}

// 设置调用信息
func Flush() {
	fish.lockFlush()
}
func SetCallInfo(b bool) {
	fish.SetCallInfo(b)
}

// 设置调用信息
func (fl *FishLogger) SetCallInfo(b bool) {
	fl.lock.Lock()
	fl.callInfo = b
	fl.lock.Unlock()
}

// 设置控制台输出
func SetConsole(b bool) {
	fish.SetConsole(b)
}

// 设置控制台输出
func (fl *FishLogger) SetConsole(b bool) {
	fl.lock.Lock()
	fl.cons = b
	fl.lock.Unlock()
}

// 获取缓存
func (l *FishLogger) getBuffer() *buffer {
	l.listLock.Lock()
	b := l.list
	if b != nil {
		l.list = b.next
	}
	l.listLock.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// 放回缓存
func (fl *FishLogger) putBuffer(b *buffer) {
	// 大缓存等待gc
	if b.Len() >= 128 {
		return
	}
	fl.listLock.Lock()
	b.next = fl.list
	fl.list = b
	fl.listLock.Unlock()
}

// 生成日志头信息
func (fl *FishLogger) header(lv logLevel, depth int) *buffer {
	now := time.Now()
	buf := fl.getBuffer()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	// format yyyymmdd hh:mm:ss.uuuu [DIWEF] file:line] msg
	buf.write4(0, year)
	buf.temp[4] = '/'
	buf.write2(5, int(month))
	buf.temp[7] = '/'
	buf.write2(8, day)
	buf.temp[10] = ' '
	buf.write2(11, hour)
	buf.temp[13] = ':'
	buf.write2(14, minute)
	buf.temp[16] = ':'
	buf.write2(17, second)
	buf.temp[19] = '.'
	buf.write4(20, now.Nanosecond()/1e5)
	buf.temp[24] = ' '
	copy(buf.temp[25:28], lv.Str())
	buf.temp[28] = ' '
	buf.Write(buf.temp[:29])
	// 调用信息
	if fl.callInfo {
		_, file, line, ok := runtime.Caller(3 + depth)
		if !ok {
			file = "###"
			line = 1
		} else {
			slash := strings.LastIndex(file, "/")
			if slash >= 0 {
				file = file[slash+1:]
			}
		}
		buf.WriteString(file)
		buf.temp[0] = ':'
		n := buf.writeN(1, line)
		buf.temp[n+1] = ']'
		buf.temp[n+2] = ' '
		buf.Write(buf.temp[:n+3])
	}
	return buf
}

// 换行输出
func (fl *FishLogger) println(lv logLevel, args ...interface{}) {
	if lv < fl.level {
		return
	}
	buf := fl.header(lv, 0)
	fmt.Fprintln(buf, args...)
	fl.write(lv, buf)
}

// 格式输出
func (fl *FishLogger) printf(lv logLevel, format string, args ...interface{}) {
	if lv < fl.level {
		return
	}
	buf := fl.header(lv, 0)
	fmt.Fprintf(buf, format, args...)
	if buf.Bytes()[buf.Len()-1] != '\n' {
		buf.WriteByte('\n')
	}
	fl.write(lv, buf)
}

// 写入数据
func (fl *FishLogger) write(lv logLevel, buf *buffer) {
	fl.lock.Lock()
	defer fl.lock.Unlock()
	data := buf.Bytes()
	if fl.cons {
		os.Stderr.Write(data)
	}
	if fl.file == nil {
		if err := fl.rotate(); err != nil {
			os.Stderr.Write(data)
			fl.exit(err)
		}
	}
	// 按天切割
	if fl.created != string(data[0:10]) {
		go fl.delete() // 每天检测一次旧文件
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}
	// 按大小切割
	if fl.size+int64(len(data)) >= fl.maxSize {
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}
	n, err := fl.writer.Write(data)
	fl.size += int64(n)
	if err != nil {
		fl.exit(err)
	}
	fl.putBuffer(buf)
}

// 删除旧日志
func (fl *FishLogger) delete() {
	dir := filepath.Dir(fl.lpath)
	fakeNow := time.Now().AddDate(0, 0, -fl.maxAge)
	filepath.Walk(dir, func(fpath string, info os.FileInfo, err error) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "logs: unable to delete old file '%s', error: %v\n", fpath, r)
			}
		}()
		if info == nil {
			return nil
		}
		// 防止误删
		if !info.IsDir() && info.ModTime().Before(fakeNow) && strings.HasSuffix(info.Name(), fl.lsuffix) {
			os.Remove(fpath)
		}
		return nil
	})
}

// 定时写入文件
func (fl *FishLogger) daemon() {
	for range time.NewTicker(flushInterval).C {
		fl.lockFlush()
	}
}

func (fl *FishLogger) lockFlush() {
	fl.lock.Lock()
	fl.flushSync()
	fl.lock.Unlock()
}

func (fl *FishLogger) flushSync() {
	if fl.file != nil {
		fl.writer.Flush() // 写入底层数据
		fl.file.Sync()    // 同步到磁盘
	}
}

func (fl *FishLogger) exit(err error) {
	fmt.Fprintf(os.Stderr, "logs: exiting because of error: %s\n", err)
	fl.flushSync()
	os.Exit(0)
}

// rotate
func (fl *FishLogger) rotate() error {
	now := time.Now()
	if fl.file != nil {
		fl.writer.Flush()
		fl.file.Sync()
		fl.file.Close()
		// 保存
		fbak := filepath.Join(fl.lname + now.Format(".2006-01-02_150405") + fl.lsuffix)
		os.Rename(fl.lpath, fbak)
		fl.size = 0
	}
	finfo, err := os.Stat(fl.lpath)
	if err == nil {
		fl.size = finfo.Size()
	}
	fout, err := os.OpenFile(fl.lpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fl.file = fout
	fl.writer = bufio.NewWriterSize(fl.file, bufferSize)
	fl.created = now.Format("2006/01/02")
	return nil
}

type buffer struct {
	temp [64]byte
	bytes.Buffer
	next *buffer
}

func (buf *buffer) write2(i, d int) {
	buf.temp[i+1] = digits[d%10]
	d /= 10
	buf.temp[i] = digits[d%10]
}

func (buf *buffer) write4(i, d int) {
	buf.temp[i+3] = digits[d%10]
	d /= 10
	buf.temp[i+2] = digits[d%10]
	d /= 10
	buf.temp[i+1] = digits[d%10]
	d /= 10
	buf.temp[i] = digits[d%10]
}

func (buf *buffer) writeN(i, d int) int {
	j := len(buf.temp)
	for d > 0 {
		j--
		buf.temp[j] = digits[d%10]
		d /= 10
	}
	return copy(buf.temp[i:], buf.temp[j:])
}

func Debug(args ...interface{}) {
	fish.println(DEBUG, args...)
}

func Debugf(format string, args ...interface{}) {
	fish.printf(DEBUG, format, args...)
}
func Info(args ...interface{}) {
	fish.println(INFO, args...)
}

func Infof(format string, args ...interface{}) {
	fish.printf(INFO, format, args...)
}

func Warn(args ...interface{}) {
	fish.println(WARN, args...)
}

func Warnf(format string, args ...interface{}) {
	fish.printf(WARN, format, args...)
}

func Error(args ...interface{}) {
	fish.println(ERROR, args...)
}

func Errorf(format string, args ...interface{}) {
	fish.printf(ERROR, format, args...)
}

func Fatal(args ...interface{}) {
	fish.println(FATAL, args...)
	os.Exit(0)
}

func Fatalf(format string, args ...interface{}) {
	fish.printf(FATAL, format, args...)
	os.Exit(0)
}
