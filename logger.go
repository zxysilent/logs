package logs

import (
	"context"
	"io"

	"github.com/zxysilent/logs/internal/file"
)

const (
	timeFieldName   = "time"
	traceFieldName  = "trace"
	levelFieldName  = "level"
	mesgFieldName   = "msg"
	errorFieldName  = "error"
	callerFieldName = "caller"
)

// Log level (aligned with log/slog numeric values).
type Level int

const (
	LevelDebug Level = -4       // slog.LevelDebug
	LevelInfo  Level = 0        // slog.LevelInfo
	LevelWarn  Level = 4        // slog.LevelWarn
	LevelError Level = 8        // slog.LevelError
	LevelNone  Level = 20241020 // sentinel: disables all output
)

// Deprecated: Use LevelDebug / LevelInfo / LevelWarn / LevelError / LevelNone instead.
const (
	LDEBUG = LevelDebug
	LINFO  = LevelInfo
	LWARN  = LevelWarn
	LERROR = LevelError
	LNONE  = LevelNone
)

// String returns the short name of the level.
func (lv Level) String() string {
	switch {
	case lv < LevelInfo:
		return "DBG"
	case lv < LevelWarn:
		return "INF"
	case lv < LevelError:
		return "WRN"
	case lv < LevelNone:
		return "ERR"
	default:
		return "OFF"
	}
}

// Logger is a lightweight handle that shares the root Config.
// Loggers constructed via New are immutable after creation.
type Logger struct {
	cfg   *config // shared root config
	trace string  // namespace / trace
	attr  []byte  // frozen preset fields (nil for plain loggers)
}

// New creates a Logger. When out is nil, output is discarded by default.
func New(out io.Writer, opts ...Option) *Logger {
	if out == nil {
		out = io.Discard
	}
	cfg := &config{
		out:    out,
		sep:    []string{"/"},
		level:  LevelInfo,
		skip:   0,
		caller: false,
		hijack: true,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	l := &Logger{cfg: cfg}
	if l.cfg.hijack {
		l.hijackstd()
	}
	return l
}

// NewFile opens a log file writer, returning the Writer and its close handle.
func NewFile(path string, opts ...FileOption) (io.Writer, func() error) {
	fw := file.New(path, true)
	for _, opt := range opts {
		opt(fw)
	}
	return fw, fw.Close
}

// preb wraps the frozen preset fields into a temporary *buffer for reuse by print/printf.
func (l *Logger) preb() *buffer {
	if len(l.attr) == 0 {
		return nil
	}
	b := buffer(l.attr)
	return &b
}

// Trace 派生一个子 Logger，使用 trace 替换当前命名空间（不拼接），保留预设字段。
// Trace derives a child Logger, replacing the current namespace with trace (no joining), preserving preset fields.
func (l *Logger) Trace(trace string) *Logger {
	c := &Logger{cfg: l.cfg, trace: trace}
	if len(l.attr) > 0 {
		c.attr = make([]byte, len(l.attr))
		copy(c.attr, l.attr)
	}
	return c
}

// Clone 派生一个子 Logger，保留预设字段；可选 trace 以点号追加到当前命名空间。
// Clone derives a child Logger preserving preset fields; an optional trace is appended to the current namespace with a dot.
func (l *Logger) Clone(trace ...string) *Logger {
	nt := l.trace
	if len(trace) > 0 {
		nt = joinTrace(l.trace, trace[0])
	}
	c := &Logger{cfg: l.cfg, trace: nt}
	if len(l.attr) > 0 {
		c.attr = make([]byte, len(l.attr))
		copy(c.attr, l.attr)
	}
	return c
}

// With 创建一个一次性 fielder，用于攒字段后通过 Group 固化为持久 Logger。
// With creates a one-time fielder for accumulating fields and then solidifying them into a persistent Logger via Group.
func (l *Logger) With(trace ...string) *fielder {
	f := getfl()
	f.cfg = l.cfg
	f.caller = l.cfg.caller
	f.attr = getb()
	*f.attr = append(*f.attr, l.attr...)
	ntrace := ""
	if len(trace) > 0 {
		ntrace = trace[0]
	}
	f.trace = joinTrace(l.trace, ntrace)
	return f
}

// Ctx 从 context 取出 traceid，与命名空间拼接后派生一次性 fielder。
// Ctx extracts traceid from context, joins it with the namespace, and derives a one-time fielder.
func (l *Logger) Ctx(ctx context.Context) *fielder {
	f := getfl()
	f.cfg = l.cfg
	f.caller = l.cfg.caller
	f.attr = getb()
	*f.attr = append(*f.attr, l.attr...)
	tid, _ := ctx.Value(traceKey).(string)
	f.trace = joinTrace(l.trace, tid)
	return f
}

// joinTrace joins namespace and sub-trace.
func joinTrace(base, sub string) string {
	if base != "" && sub != "" {
		return base + "." + sub
	}
	if base != "" {
		return base
	}
	return sub
}

// Debug logs at debug level.
func (l *Logger) Debug(args ...any) {
	if LevelDebug >= l.cfg.level {
		l.cfg.print(l.trace, LevelDebug, l.cfg.caller, l.preb(), args...)
	}
}

// Debugf logs a formatted message at debug level.
func (l *Logger) Debugf(format string, args ...any) {
	if LevelDebug >= l.cfg.level {
		l.cfg.printf(l.trace, LevelDebug, l.cfg.caller, l.preb(), format, args...)
	}
}

// Info logs at info level.
func (l *Logger) Info(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(format string, args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.printf(l.trace, LevelInfo, l.cfg.caller, l.preb(), format, args...)
	}
}

// Warn logs at warn level.
func (l *Logger) Warn(args ...any) {
	if LevelWarn >= l.cfg.level {
		l.cfg.print(l.trace, LevelWarn, l.cfg.caller, l.preb(), args...)
	}
}

// Warnf logs a formatted message at warn level.
func (l *Logger) Warnf(format string, args ...any) {
	if LevelWarn >= l.cfg.level {
		l.cfg.printf(l.trace, LevelWarn, l.cfg.caller, l.preb(), format, args...)
	}
}

// Error logs at error level.
func (l *Logger) Error(args ...any) {
	if LevelError >= l.cfg.level {
		l.cfg.print(l.trace, LevelError, l.cfg.caller, l.preb(), args...)
	}
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(format string, args ...any) {
	if LevelError >= l.cfg.level {
		l.cfg.printf(l.trace, LevelError, l.cfg.caller, l.preb(), format, args...)
	}
}
