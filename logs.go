package logs

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
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
	logShort = "DBGINFWRNERR" //DBG INF WRN ERR
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
	return logShort[lv*3 : lv*3+3]
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

func SetSkip(skip int) {
	log.SetSkip(skip)
}

func (fl *Logger) SetSkip(skip int) {
	fl.mu.Lock()
	fl.skip = skip
	fl.mu.Unlock()
}
func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

func (fl *Logger) SetOutput(out io.Writer) {
	fl.mu.Lock()
	fl.out = out
	fl.mu.Unlock()
}
func (l *Logger) Write(p []byte) (int, error) {
	// l.mu.Lock()
	// defer l.mu.Unlock()
	return l.out.Write(p)
}

func (l *Logger) With() *FieldLogger {
	f := &FieldLogger{}
	f.logger = l
	f.caller = l.caller
	f.attr = buffer.Get()
	return f
}

// tracking
func (l *Logger) Ctx(ctx context.Context) *FieldLogger {
	f := &FieldLogger{}
	f.ctx = ctx
	f.logger = l
	f.caller = l.caller
	f.attr = buffer.Get()
	return f
}

const trackKey = "logs-track-id"

func TrackCtx(ctx context.Context, trackid ...string) context.Context {
	val := ctx.Value(trackKey)
	if val == nil {
		var id = ""
		if len(trackid) == 0 {
			id = trace()
		} else {
			id = trackid[0]
		}
		ctx = context.WithValue(ctx, trackKey, id)
	}
	return ctx
}

type FieldLogger struct {
	ctx    context.Context
	attr   *buffer.Buffer //调用输出后清空
	buf    *buffer.Buffer //每次输出的时候重置
	logger *Logger
	caller bool
}

func (s *FieldLogger) Caller(b bool) *FieldLogger {
	s.caller = b
	return s
}
func header(ctx context.Context, caller bool, skip int, sep string, buf *buffer.Buffer, lv logLevel) {
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
		*buf = enc.PutString(enc.PutKey(*buf, callerFieldName), file+":"+strconv.Itoa(line))
	}
}

func print(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log.skip, log.sep, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ',')
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprint(args...))
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)
}

func printf(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, format string, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log.skip, log.sep, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ',')
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), format)
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)

}
func (fl *FieldLogger) Debug(args ...interface{}) {
	if LDEBUG >= fl.logger.level {
		print(fl.ctx, LDEBUG, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Debugf(foramt string, args ...interface{}) {
	if LDEBUG >= fl.logger.level {
		printf(fl.ctx, LDEBUG, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Info(args ...interface{}) {
	if LINFO >= fl.logger.level {
		print(fl.ctx, LINFO, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Infof(foramt string, args ...interface{}) {
	if LINFO >= fl.logger.level {
		printf(fl.ctx, LINFO, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Warn(args ...interface{}) {
	if LWARN >= fl.logger.level {
		print(fl.ctx, LWARN, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Warnf(foramt string, args ...interface{}) {
	if LWARN >= fl.logger.level {
		printf(fl.ctx, LWARN, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}
func (fl *FieldLogger) Error(args ...interface{}) {
	if LERROR >= fl.logger.level {
		print(fl.ctx, LERROR, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Errorf(foramt string, args ...interface{}) {
	if LERROR >= fl.logger.level {
		printf(fl.ctx, LERROR, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

// ----------------------------------------------------------------
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

var log = New(os.Stdout)

func Debug(args ...interface{}) {
	if LDEBUG >= log.level {
		print(nil, LDEBUG, log.caller, log, nil, args...)
	}
}

func Debugf(foramt string, args ...interface{}) {
	if LDEBUG >= log.level {
		printf(nil, LDEBUG, log.caller, log, nil, foramt, args...)
	}
}

func Info(args ...interface{}) {
	if LINFO >= log.level {
		print(nil, LINFO, log.caller, log, nil, args...)
	}
}

func Infof(foramt string, args ...interface{}) {
	if LINFO >= log.level {
		printf(nil, LINFO, log.caller, log, nil, foramt, args...)
	}
}

func Warn(args ...interface{}) {
	if LWARN >= log.level {
		print(nil, LWARN, log.caller, log, nil, args...)
	}
}

func Warnf(foramt string, args ...interface{}) {
	if LWARN >= log.level {
		printf(nil, LWARN, log.caller, log, nil, foramt, args...)
	}
}

func Error(args ...interface{}) {
	if LERROR >= log.level {
		print(nil, LERROR, log.caller, log, nil, args...)
	}
}

func Errorf(foramt string, args ...interface{}) {
	if LERROR >= log.level {
		printf(nil, LERROR, log.caller, log, nil, foramt, args...)
	}
}

func With() *FieldLogger {
	return log.With()
}

func Ctx(ctx context.Context) *FieldLogger {
	return log.Ctx(ctx)
}
