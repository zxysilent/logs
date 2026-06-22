package logs

import (
	"context"
	"io"
	"os"
)

// l is the package-level default instance.
var l = New(os.Stderr)

// SetLevel sets the log level of the default instance.
func SetLevel(lv Level) {
	l.cfg.setLevel(lv)
}

// SetSep sets the caller path separators.
func SetSep(sep ...string) {
	l.cfg.setSep(sep...)
}

// SetCaller sets whether the default instance outputs caller information.
func SetCaller(b bool) {
	l.cfg.setCaller(b)
}

// SetSkip sets the number of caller frames to skip.
func SetSkip(skip int) {
	l.cfg.setSkip(skip)
}

// SetOutput sets the output writer.
func SetOutput(out io.Writer) {
	l.cfg.setOutput(out)
}

// SetFile sets the log file path.
func SetFile(path string) {
	l.cfg.setFile(path)
}

// SetMaxAge sets the maximum number of days to retain log files.
func SetMaxAge(ma int) {
	l.cfg.setMaxAge(ma)
}

// SetMaxSize sets the maximum size of a single log file in MiB.
func SetMaxSize(ms int64) {
	l.cfg.setMaxSize(ms)
}

// SetCons sets whether to also output to the console.
// Deprecated: Use SetConsole instead.
func SetCons(b bool) {
	l.cfg.setConsole(b)
}

// SetConsole sets whether to also output to the console.
func SetConsole(b bool) {
	l.cfg.setConsole(b)
}

// Debug logs at debug level using the default instance.
func Debug(args ...any) { l.Debug(args...) }

// Debugf logs a formatted message at debug level using the default instance.
func Debugf(format string, args ...any) { l.Debugf(format, args...) }

// Info logs at info level using the default instance.
func Info(args ...any) { l.Info(args...) }

// Infof logs a formatted message at info level using the default instance.
func Infof(format string, args ...any) { l.Infof(format, args...) }

// Warn logs at warn level using the default instance.
func Warn(args ...any) { l.Warn(args...) }

// Warnf logs a formatted message at warn level using the default instance.
func Warnf(format string, args ...any) { l.Warnf(format, args...) }

// Error logs at error level using the default instance.
func Error(args ...any) { l.Error(args...) }

// Errorf logs a formatted message at error level using the default instance.
func Errorf(format string, args ...any) { l.Errorf(format, args...) }

// Print logs at info level (stdlib-compatible) using the default instance.
func Print(args ...any) { l.Print(args...) }

// Println logs at info level (stdlib-compatible) using the default instance.
func Println(args ...any) { l.Println(args...) }

// Printf logs a formatted message at info level (stdlib-compatible) using the default instance.
func Printf(format string, args ...any) { l.Printf(format, args...) }

// With is the field logging entry.
func With(trace ...string) *fielder {
	return l.With(trace...)
}

// Ctx is the context logging entry.
func Ctx(ctx context.Context) *fielder {
	return l.Ctx(ctx)
}

// Trace is the trace logging entry.
func Trace(trace string) *Logger {
	return l.Trace(trace)
}

// Clone creates a new logger instance with the same configuration as the default instance.
func Clone() *Logger {
	return l.Clone()
}

// Close closes the log file.
func Close() error {
	return l.cfg.close()
}
