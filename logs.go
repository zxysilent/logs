package logs

import (
	"context"
	"io"
	"sync"

	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/encoder"
	"github.com/zxysilent/logs/internal/encoder/json"
	"github.com/zxysilent/logs/internal/encoder/text"
	"github.com/zxysilent/logs/internal/file"
)

// 日志等级
type logLevel int

const (
	LDEBUG logLevel = iota
	LINFO
	LWARN
	LERROR
	logShort = "DBGINFWRNERR" //DBG INF WRN ERR
)
const (
	timeFieldName   = "time"
	traceFieldName  = "trace"
	levelFieldName  = "level"
	msgFieldName    = "msg"
	errorFieldName  = "error"
	callerFieldName = "caller"
	timeFieldFormat = "2006/01/02 15:04:05.000"
)

// 字符串等级
func (lv logLevel) String() string {
	return logShort[lv*3 : lv*3+3]
}

type Logger struct {
	out    io.Writer       // 输出
	sep    string          // 路径分隔
	caller bool            // 调用信息
	level  logLevel        // 日志等级
	skip   int             //
	enc    encoder.Encoder //
	mu     sync.Mutex      // logger🔒
	fw     *file.Writer
}

func New(out io.Writer) *Logger {
	if out == nil {
		out = io.Discard
	}
	n := &Logger{
		out:    out,
		caller: false,
		level:  LINFO,
		skip:   0,
		sep:    "/",
	}
	n.enc = &text.Encoder{}
	return n
}

func (l *Logger) SetFile(path string) {
	l.fw = file.New(path)
	l.SetOutput(l.fw)
}

// SetMaxAge 最大保留天数
func (l *Logger) SetMaxAge(ma int) {
	if l.fw == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fw.SetMaxAge(ma)
}

// SetMaxSize 单个日志最大容量
func (l *Logger) SetMaxSize(ms int64) {
	if l.fw == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fw.SetMaxSize(ms)
}

// SetCons 同时输出控制台
func (l *Logger) SetCons(b bool) {
	if l.fw == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fw.SetCons(b)
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.fw != nil {
		return l.fw.Close()
	}
	return nil
}

// 设置输出等级
func (l *Logger) SetLevel(lv logLevel) {
	if lv < LDEBUG || lv > LERROR {
		panic("logs: illegal log level")
	}
	l.mu.Lock()
	l.level = lv
	l.mu.Unlock()
}

// 设置调用信息
func (l *Logger) SetCaller(b bool) {
	l.mu.Lock()
	l.caller = b
	l.mu.Unlock()
}

func (l *Logger) SetJSON() {
	l.enc = &json.Encoder{}
}

func (l *Logger) SetText() {
	l.enc = &text.Encoder{}
}

func (l *Logger) SetSep(sep string) {
	l.mu.Lock()
	l.sep = sep
	l.mu.Unlock()
}

func (l *Logger) SetSkip(skip int) {
	l.mu.Lock()
	l.skip = skip
	l.mu.Unlock()
}

func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	l.out = out
	l.mu.Unlock()
}
func (l *Logger) Write(p []byte) (int, error) {
	return l.out.Write(p)
}

func (l *Logger) With() *FieldLogger {
	f := &FieldLogger{}
	f.enc = l.enc
	f.logger = l
	f.caller = l.caller
	f.attr = buffer.Get()
	return f
}

// tracing
func (l *Logger) Ctx(ctx context.Context) *FieldLogger {
	f := &FieldLogger{}
	f.enc = l.enc
	f.ctx = ctx
	f.logger = l
	f.caller = l.caller
	f.attr = buffer.Get()
	return f
}

func (l *Logger) Debug(args ...interface{}) {
	if LDEBUG >= l.level {
		print(nil, LDEBUG, l.caller, l, nil, args...)
	}
}
func (l *Logger) Debugf(foramt string, args ...interface{}) {
	if LDEBUG >= l.level {
		printf(nil, LDEBUG, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Info(args ...interface{}) {
	if LINFO >= l.level {
		print(nil, LINFO, l.caller, l, nil, args...)
	}
}

func (l *Logger) Infof(foramt string, args ...interface{}) {
	if LINFO >= l.level {
		printf(nil, LINFO, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Warn(args ...interface{}) {
	if LWARN >= l.level {
		print(nil, LWARN, l.caller, l, nil, args...)
	}
}

func (l *Logger) Warnf(foramt string, args ...interface{}) {
	if LWARN >= l.level {
		printf(nil, LWARN, l.caller, l, nil, foramt, args...)
	}
}
func (l *Logger) Error(args ...interface{}) {
	if LERROR >= l.level {
		print(nil, LERROR, l.caller, l, nil, args...)
	}
}

func (l *Logger) Errorf(foramt string, args ...interface{}) {
	if LERROR >= l.level {
		printf(nil, LERROR, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Writer() io.Writer {
	return l
}
