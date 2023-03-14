package logs

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/encoder"
)

var (
	enc = encoder.Encoder{}
)

// 日志等级
type logLevel int

const (
	LDEBUG logLevel = iota
	LINFO
	LWARN
	LERROR
	LNone
	logShort = "DBGINFWRNERR" //TRC DBG INF WRN ERR FTL PNC
)
const (
	timeFieldName     = "time"
	traceFieldName    = "trace"
	levelFieldName    = "level"
	msgFieldName      = "msg"
	errorFieldName    = "error"
	callerFieldName   = "caller"
	timeFieldFormat   = "2006/01/02 15:04:05.000"
	durationFieldUnit = time.Millisecond
)

// 字符串等级
func (lv logLevel) String() string {
	if lv >= LDEBUG && lv <= LNone {
		return logShort[lv*3 : lv*3+3]
	}
	return "NIL"
}

type Logger struct {
	out    io.Writer  // 输出
	sep    string     // 路径分隔
	caller bool       // 调用信息
	level  logLevel   // 日志等级
	skip   int        //
	mu     sync.Mutex // logger🔒
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
	return n
}

// 设置实例等级
func SetLevel(lv logLevel) {
	log.SetLevel(lv)
}

// 设置输出等级
func (fl *Logger) SetLevel(lv logLevel) {
	if lv < LDEBUG || lv > LERROR {
		panic("非法的日志等级")
	}
	fl.mu.Lock()
	fl.level = lv
	fl.mu.Unlock()
}

func SetCaller(b bool) {
	log.SetCaller(b)
}

// 设置调用信息
func (fl *Logger) SetCaller(b bool) {
	fl.mu.Lock()
	fl.caller = b
	fl.mu.Unlock()
}

func SetSep(sep string) {
	log.SetSep(sep)
}

func (fl *Logger) SetSep(sep string) {
	fl.mu.Lock()
	fl.sep = sep
	fl.mu.Unlock()
}

func SetSkip(skip string) {
	log.SetSkip(skip)
}

func (fl *Logger) SetSkip(skip string) {
	fl.mu.Lock()
	fl.sep = skip
	fl.mu.Unlock()
}

func (l *Logger) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.out.Write(p)
}

func (l *Logger) With() *FieldLogger {
	f := &FieldLogger{}
	f.logger = l
	f.attr = buffer.Get()
	return f
}

// tracking
func (l *Logger) Ctx(ctx context.Context) *FieldLogger {
	f := &FieldLogger{}
	f.logger = l
	f.ctx = ctx
	f.attr = buffer.Get()
	return f
}

const trackKey = "x-track-id"

func TrackCtx(ctx context.Context, trackid ...string) context.Context {
	val := ctx.Value(trackKey)
	if val == nil {
		var id = ""
		if len(trackid) == 0 {
			id = uuid()
		} else {
			id = trackid[0]
		}
		ctx = context.WithValue(ctx, trackKey, id)
	}
	return ctx
}

type FieldLogger struct {
	logger *Logger
	ctx    context.Context
	attr   *buffer.Buffer //调用输出后清空
	buf    *buffer.Buffer //每次输出的时候重置
}

func header(ctx context.Context, caller bool, skip int, sep string, buf *buffer.Buffer, lv logLevel) {
	// f.buf.Reset()
	*buf = enc.PutBeginMarker(*buf)
	*buf = enc.PutTimeFast(enc.PutKey(*buf, timeFieldName), time.Now())
	*buf = enc.PutString(enc.PutKey(*buf, levelFieldName), lv.String())
	if ctx != nil {
		val := ctx.Value(trackKey)
		if val != nil {
			if traceId, ok := val.(string); ok {
				*buf = enc.PutString(enc.PutKey(*buf, traceFieldName), traceId)
			}
		}
	}
	if caller {
		_, file, line, ok := runtime.Caller(skip + 3)
		if !ok {
			file = "###"
			line = 1
		} else {
			slash := strings.LastIndex(file, sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = enc.PutString(enc.PutKey(*buf, callerFieldName), fmt.Sprintf("%s:%d", file, line))
	}
}

func (fl *Logger) print(lv logLevel, args ...interface{}) {
	if lv < fl.level {
		return
	}
	buf := buffer.Get()
	header(nil, fl.caller, fl.skip, fl.sep, buf, lv)
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprint(args...))
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	fl.Write(*buf)
	buffer.Put(buf)
}

func (fl *Logger) printf(lv logLevel, format string, args ...interface{}) {
	if lv < fl.level {
		return
	}
	buf := buffer.Get()
	header(nil, fl.caller, fl.skip, fl.sep, buf, lv)
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), format)
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	fl.Write(*buf)
	buffer.Put(buf)
}

func (fl *FieldLogger) print(lv logLevel, args ...interface{}) {
	if lv < fl.logger.level {
		return
	}
	fl.buf = buffer.Get()
	header(fl.ctx, fl.logger.caller, fl.logger.skip, fl.logger.sep, fl.buf, lv)
	if fl.attr != nil && len(*fl.attr) >= 1 {
		*fl.buf = append(*fl.buf, ',')
		*fl.buf = append(*fl.buf, *fl.attr...)
	}
	if len(args) >= 1 {
		*fl.buf = enc.PutString(enc.PutKey(*fl.buf, msgFieldName), fmt.Sprint(args...))
	}
	*fl.buf = enc.PutEndMarker(*fl.buf)
	*fl.buf = enc.PutLineBreak(*fl.buf)
	fl.logger.Write(*fl.buf)
	buffer.Put(fl.buf)
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) printf(lv logLevel, format string, args ...interface{}) {
	if lv < fl.logger.level {
		return
	}
	fl.buf = buffer.Get()
	header(fl.ctx, fl.logger.caller, fl.logger.skip, fl.logger.sep, fl.buf, lv)
	if fl.attr != nil && len(*fl.attr) >= 1 {
		*fl.buf = append(*fl.buf, ',')
		*fl.buf = append(*fl.buf, *fl.attr...)
	}
	if format != "" {
		if len(args) >= 1 {
			*fl.buf = enc.PutString(enc.PutKey(*fl.buf, msgFieldName), fmt.Sprintf(format, args...))
		} else {
			*fl.buf = enc.PutString(enc.PutKey(*fl.buf, msgFieldName), format)
		}
	}

	*fl.buf = enc.PutEndMarker(*fl.buf)
	*fl.buf = enc.PutLineBreak(*fl.buf)
	fl.logger.Write(*fl.buf)
	buffer.Put(fl.buf)
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Debug(args ...interface{}) {
	fl.print(LDEBUG, args...)
}

func (fl *FieldLogger) Debugf(foramt string, args ...interface{}) {
	fl.printf(LDEBUG, foramt, args...)
}

func (fl *FieldLogger) Info(args ...interface{}) {
	fl.print(LINFO, args...)
}

func (fl *FieldLogger) Infof(foramt string, args ...interface{}) {
	fl.printf(LINFO, foramt, args...)
}

func (fl *FieldLogger) Warn(args ...interface{}) {
	fl.print(LWARN, args...)
}

func (fl *FieldLogger) Warnf(foramt string, args ...interface{}) {
	fl.printf(LWARN, foramt, args...)
}
func (fl *FieldLogger) Error(args ...interface{}) {
	fl.print(LERROR, args...)
}

func (fl *FieldLogger) Errorf(foramt string, args ...interface{}) {
	fl.printf(LERROR, foramt, args...)
}
func (fl *Logger) Debug(args ...interface{}) {
	fl.print(LDEBUG, args...)
}

//----------------------------------------------------------------

func (l *Logger) Debugf(foramt string, args ...interface{}) {
	l.printf(LDEBUG, foramt, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.print(LINFO, args...)
}

func (l *Logger) Infof(foramt string, args ...interface{}) {
	l.printf(LINFO, foramt, args...)
}

func (fl *Logger) Warn(args ...interface{}) {
	fl.print(LWARN, args...)
}

func (l *Logger) Warnf(foramt string, args ...interface{}) {
	l.printf(LWARN, foramt, args...)
}
func (l *Logger) Error(args ...interface{}) {
	l.print(LERROR, args...)
}

func (l *Logger) Errorf(foramt string, args ...interface{}) {
	l.printf(LERROR, foramt, args...)
}

func (l *Logger) Writer() io.Writer {
	return l
}

var log = New(os.Stdout)

func Debug(args ...interface{}) {
	log.print(LDEBUG, args...)
}
func Debugf(foramt string, args ...interface{}) {
	log.printf(LDEBUG, foramt, args...)
}

func Info(args ...interface{}) {
	log.print(LINFO, args...)
}

func Infof(foramt string, args ...interface{}) {
	log.printf(LINFO, foramt, args...)
}

func Warn(args ...interface{}) {
	log.print(LWARN, args...)
}

func Warnf(foramt string, args ...interface{}) {
	log.printf(LWARN, foramt, args...)
}

func Error(args ...interface{}) {
	log.print(LERROR, args...)
}

func Errorf(foramt string, args ...interface{}) {
	log.printf(LERROR, foramt, args...)
}

func With() *FieldLogger {
	return log.With()
}

func Ctx(ctx context.Context) *FieldLogger {
	return log.Ctx(ctx)
}
