package logs

import (
	"context"
	"io"
	"os"
)

// l is the package-level default instance.
var l = New(os.Stderr)

// Set* functions below modify the package-level default instance.
// They are provided for one-time initialization before logging starts.
// Runtime modification after logging has begun is NOT recommended —
// config writes are unsynchronized and may race with concurrent log output.
// Prefer New() with functional options for immutable, concurrency-safe loggers.

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
// Must be called after SetFile.
func SetMaxAge(ma int) {
	l.cfg.setMaxAge(ma)
}

// SetMaxSize sets the maximum size of a single log file in MiB.
// Must be called after SetFile.
func SetMaxSize(ms int64) {
	l.cfg.setMaxSize(ms)
}

// SetCons sets whether to also output to the console.
//
// Deprecated: Use SetConsole instead. This alias is kept for backward
// compatibility and will be removed in a future major version.
func SetCons(b bool) {
	l.cfg.setConsole(b)
}

// SetConsole sets whether to also output to stderr when writing to a file.
// Only takes effect when a file writer is active (SetFile).
func SetConsole(b bool) {
	l.cfg.setConsole(b)
}

// SetTrace sets the trace.
func SetTrace(trace string) {
	l.trace = trace
}

// The following functions use method-valued variables instead of wrapper
// functions to keep the caller skip depth identical: logs.Debug and l.Debug
// produce the same caller:file:line. Wrapping with `func Debug(...) { l.Debug(...) }`
// would add one extra frame, pushing the caller one level further.

// Debug logs at debug level.
var Debug = l.Debug

// Debugf logs a formatted message at debug level.
var Debugf = l.Debugf

// Info logs at info level.
var Info = l.Info

// Infof logs a formatted message at info level.
var Infof = l.Infof

// Warn logs at warn level.
var Warn = l.Warn

// Warnf logs a formatted message at warn level.
var Warnf = l.Warnf

// Error logs at error level.
var Error = l.Error

// Errorf logs a formatted message at error level.
var Errorf = l.Errorf

// Print logs at info level (stdlib-compatible).
var Print = l.Print

// Printf logs a formatted message at info level (stdlib-compatible).
var Printf = l.Printf

// Println logs at info level (stdlib-compatible).
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
