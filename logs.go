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

// SetTrace sets the trace.
func SetTrace(trace string) {
	l.cfg.mu.Lock()
	defer l.cfg.mu.Unlock()
	l.trace = trace
}

// Debug logs at debug level using the default instance.
var Debug = l.Debug

// Debugf logs a formatted message at debug level using the default instance.
var Debugf = l.Debugf

// Info logs at info level using the default instance.
var Info = l.Info

// Infof logs a formatted message at info level using the default instance.
var Infof = l.Infof

// Warn logs at warn level using the default instance.
var Warn = l.Warn

// Warnf logs a formatted message at warn level using the default instance.
var Warnf = l.Warnf

// Error logs at error level using the default instance.
var Error = l.Error

// Errorf logs a formatted message at error level using the default instance.
var Errorf = l.Errorf

// Print logs at info level (stdlib-compatible) using the default instance.
var Print = l.Print

// Printf logs a formatted message at info level (stdlib-compatible) using the default instance.
var Printf = l.Printf

// Println logs at info level (stdlib-compatible) using the default instance.
var Println = l.Println

// With is the field logging entry.
func With(trace ...string) *fielder {
	return l.With(trace...)
}

// Ctx is the context logging entry.
func Ctx(ctx context.Context) *fielder {
	return l.Ctx(ctx)
}

// Trace replaces the namespace of the default instance and returns a child Logger.
func Trace(trace string) *Logger {
	return l.Trace(trace)
}

// Clone derives a child Logger from the default instance; an optional trace is appended to the namespace.
func Clone(trace ...string) *Logger {
	return l.Clone(trace...)
}

// Close closes the log file.
func Close() error {
	return l.cfg.close()
}
