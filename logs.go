package logs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
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
	maxAge        = 180               // 180 天
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
	maxAge   int           // 最大保留天数
	maxSize  int64         // 单个日志最大容量 默认 256MB
	size     int64         // 累计大小 无后缀
	lpath    string        // 文件目录 完整路径 lpath=lname+lsuffix
	lname    string        // 文件名
	lsuffix  string        // 文件后缀名 默认 .log
	created  string        // 文件创建日期
	level    logLevel      // 输出的日志等级
	pool     sync.Pool     // Pool
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
	os.MkdirAll(filepath.Dir(lpath), 0755)
	fl.level = DEBUG
	fl.maxAge = maxAge
	fl.maxSize = maxSize
	fl.pool = sync.Pool{
		New: func() interface{} {
			return new(buffer)
		},
	}
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

// 设置最大保存天数
// 小于0不删除
func SetMaxAge(ma int) {
	fish.SetMaxAge(ma)
}

// 设置最大保存天数
// 小于0不删除
func (fl *FishLogger) SetMaxAge(ma int) {
	fl.lock.Lock()
	fl.maxAge = ma
	fl.lock.Unlock()
}

// 写入文件
func Flush() {
	fish.Flush()
}

// 写入文件
func (fl *FishLogger) Flush() {
	fl.lock.Lock()
	fl.flushSync()
	fl.lock.Unlock()
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

// 生成日志头信息
func (fl *FishLogger) header(lv logLevel, depth int) *buffer {
	now := time.Now()
	buf := fl.pool.Get().(*buffer)
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
	buf.temp[19] = ' '
	copy(buf.temp[20:23], lv.Str())
	buf.temp[23] = ' '
	buf.Write(buf.temp[:24])
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
	fl.Write(buf.Bytes())
	buf.Reset()
	fl.pool.Put(buf)
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
	fl.Write(buf.Bytes())
	buf.Reset()
	fl.pool.Put(buf)
}

// 写入数据
func (fl *FishLogger) Write(buf []byte) (n int, err error) {
	fl.lock.Lock()
	defer fl.lock.Unlock()
	if fl.cons {
		os.Stderr.Write(buf)
	}
	if fl.file == nil {
		if err := fl.rotate(); err != nil {
			os.Stderr.Write(buf)
			fl.exit(err)
		}
	}
	// 按天切割
	if fl.created != string(buf[0:10]) {
		go fl.delete() // 每天检测一次旧文件
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}
	// 按大小切割
	if fl.size+int64(len(buf)) >= fl.maxSize {
		if err := fl.rotate(); err != nil {
			fl.exit(err)
		}
	}
	n, err = fl.writer.Write(buf)
	fl.size += int64(n)
	if err != nil {
		fl.exit(err)
	}
	return
}

// 删除旧日志
func (fl *FishLogger) delete() {
	if fl.maxAge < 0 {
		return
	}
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
		fl.Flush()
	}
}

// 不能锁
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

// rotate 切割文件
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
	fl.created = now.Format("2006/01/02")
	if err == nil {
		fl.size = finfo.Size()
		fl.created = finfo.ModTime().Format("2006/01/02")
	}
	fout, err := os.OpenFile(fl.lpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	fl.file = fout
	fl.writer = bufio.NewWriterSize(fl.file, bufferSize)
	return nil
}

type buffer struct {
	temp [64]byte
	bytes.Buffer
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

// -------- 实例 fish

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
func Writer() io.Writer {
	return fish
}

// -------- 实例 自定义

func (fl *FishLogger) Debug(args ...interface{}) {
	fl.println(DEBUG, args...)
}

func (fl *FishLogger) Debugf(format string, args ...interface{}) {
	fl.printf(DEBUG, format, args...)
}
func (fl *FishLogger) Info(args ...interface{}) {
	fl.println(INFO, args...)
}

func (fl *FishLogger) Infof(format string, args ...interface{}) {
	fl.printf(INFO, format, args...)
}

func (fl *FishLogger) Warn(args ...interface{}) {
	fl.println(WARN, args...)
}

func (fl *FishLogger) Warnf(format string, args ...interface{}) {
	fl.printf(WARN, format, args...)
}

func (fl *FishLogger) Error(args ...interface{}) {
	fl.println(ERROR, args...)
}

func (fl *FishLogger) Errorf(format string, args ...interface{}) {
	fl.printf(ERROR, format, args...)
}

func (fl *FishLogger) Fatal(args ...interface{}) {
	fl.println(FATAL, args...)
	os.Exit(0)
}

func (fl *FishLogger) Fatalf(format string, args ...interface{}) {
	fl.printf(FATAL, format, args...)
	os.Exit(0)
}

func (fl *FishLogger) Writer() io.Writer {
	return fl
}
