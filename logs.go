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
// Only takes effect when a file writer is active (SetFile).
func SetMaxAge(ma int) {
	l.cfg.setMaxAge(ma)
}

// SetMaxSize sets the maximum size of a single log file in MiB.
// Only takes effect when a file writer is active (SetFile).
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

// Debug logs at debug level.
func Debug(args ...any) {
	if LevelDebug >= l.cfg.level {
		l.cfg.print(l.trace, LevelDebug, l.cfg.caller, l.preb(), args...)
	}
}

// Debugf logs a formatted message at debug level.
func Debugf(format string, args ...any) {
	if LevelDebug >= l.cfg.level {
		l.cfg.printf(l.trace, LevelDebug, l.cfg.caller, l.preb(), format, args...)
	}
}

// Info logs at info level.
func Info(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

// Infof logs a formatted message at info level.
func Infof(format string, args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.printf(l.trace, LevelInfo, l.cfg.caller, l.preb(), format, args...)
	}
}

// Warn logs at warn level.
func Warn(args ...any) {
	if LevelWarn >= l.cfg.level {
		l.cfg.print(l.trace, LevelWarn, l.cfg.caller, l.preb(), args...)
	}
}

// Warnf logs a formatted message at warn level.
func Warnf(format string, args ...any) {
	if LevelWarn >= l.cfg.level {
		l.cfg.printf(l.trace, LevelWarn, l.cfg.caller, l.preb(), format, args...)
	}
}

// Error logs at error level.
func Error(args ...any) {
	if LevelError >= l.cfg.level {
		l.cfg.print(l.trace, LevelError, l.cfg.caller, l.preb(), args...)
	}
}

// Errorf logs a formatted message at error level.
func Errorf(format string, args ...any) {
	if LevelError >= l.cfg.level {
		l.cfg.printf(l.trace, LevelError, l.cfg.caller, l.preb(), format, args...)
	}
}

// Print logs at info level (stdlib-compatible).
func Print(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

// Printf logs a formatted message at info level (stdlib-compatible).
func Printf(format string, args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.printf(l.trace, LevelInfo, l.cfg.caller, l.preb(), format, args...)
	}
}

// Println logs at info level (stdlib-compatible).
func Println(args ...any) {
	if LevelInfo >= l.cfg.level {
		l.cfg.print(l.trace, LevelInfo, l.cfg.caller, l.preb(), args...)
	}
}

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
