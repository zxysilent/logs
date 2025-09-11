package logs

import (
	"context"
	"io"
	"sync"

	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/file"
	"github.com/zxysilent/logs/internal/textenc"
)

// 日志等级
type logLevel int

const (
	LDEBUG logLevel = iota
	LINFO
	LWARN
	LERROR
	LNONE
	logShort = "DBGINFWRNERR" //DBG INF WRN ERR
)
const (
	timeFieldName   = "time"
	traceFieldName  = "trace"
	levelFieldName  = "level"
	msgFieldName    = "msg"
	errorFieldName  = "error"
	callerFieldName = "caller"
)

// 字符串等级
func (lv logLevel) String() string {
	return logShort[lv*3 : lv*3+3]
}

type Logger struct {
	out    io.Writer        // 输出
	sep    string           // 路径分隔
	caller bool             // 调用信息
	level  logLevel         // 日志等级
	skip   int              //
	enc    *textenc.Encoder //
	mu     sync.Mutex       // logger🔒
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
		enc:    textenc.NewEncoder(),
	}
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
	if lv < LDEBUG || lv > LNONE {
		panic("illegal log level")
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
	f.trace, _ = ctx.Value(traceKey).(string)
	f.logger = l
	f.caller = l.caller
	f.attr = buffer.Get()
	return f
}

func (l *Logger) Debug(args ...any) {
	if LDEBUG >= l.level {
		print("", LDEBUG, l.caller, l, nil, args...)
	}
}
func (l *Logger) Debugf(foramt string, args ...any) {
	if LDEBUG >= l.level {
		printf("", LDEBUG, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Info(args ...any) {
	if LINFO >= l.level {
		print("", LINFO, l.caller, l, nil, args...)
	}
}

func (l *Logger) Infof(foramt string, args ...any) {
	if LINFO >= l.level {
		printf("", LINFO, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Warn(args ...any) {
	if LWARN >= l.level {
		print("", LWARN, l.caller, l, nil, args...)
	}
}

func (l *Logger) Warnf(foramt string, args ...any) {
	if LWARN >= l.level {
		printf("", LWARN, l.caller, l, nil, foramt, args...)
	}
}
func (l *Logger) Error(args ...any) {
	if LERROR >= l.level {
		print("", LERROR, l.caller, l, nil, args...)
	}
}

func (l *Logger) Errorf(foramt string, args ...any) {
	if LERROR >= l.level {
		printf("", LERROR, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) Writer() io.Writer {
	return l
}
