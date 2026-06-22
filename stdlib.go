package logs

import (
	"bytes"
	"io"
	stdlog "log"
)

// hijackstd 劫持标准库 log 输出，将其重定向到本库的日志系统。
// hijackstd hijacks standard library log output and redirects it to this library's logging system.
func (l *Logger) hijackstd() {
	stdlog.SetFlags(0)
	prefix := stdlog.Prefix()
	stdlog.SetPrefix("")
	stdlog.SetOutput(l.stdWriter(prefix))
}

// stdWriter is an io.Writer that redirects standard library log output into this library.
type stdWriter struct {
	cfg    *config
	prefix string
}

// Write parses a standard library log line and emits it via the shared config.
func (w *stdWriter) Write(p []byte) (int, error) {
	if w == nil || w.cfg == nil {
		return len(p), nil
	}
	if LevelInfo < w.cfg.level {
		return len(p), nil
	}
	msg := bytes.TrimRight(p, "\n")
	nsb := []byte(w.prefix)
	if w.prefix != "" {
		msg = bytes.TrimPrefix(msg, nsb)
	}
	w.cfg.printb(w.prefix, LevelInfo, w.cfg.caller, nil, msg)
	return len(p), nil
}

// Print logs at info level (stdlib-compatible).
func (l *Logger) Print(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

// Println logs at info level (stdlib-compatible).
func (l *Logger) Println(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

// Printf logs a formatted message at info level (stdlib-compatible).
func (l *Logger) Printf(format string, args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.printf(l.trace, LevelInfo, l.cfg.caller, l.preb(), format, args...)
	}
}

// stdWriter builds an io.Writer that feeds standard library log output into this Logger.
func (l *Logger) stdWriter(prefix string) io.Writer {
	return &stdWriter{cfg: l.cfg, prefix: prefix}
}
