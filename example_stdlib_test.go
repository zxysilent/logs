// Package logs_test demonstrates standard library integration API.
package logs_test

import (
	"context"
	stdlog "log"
	"log/slog"
	"os"
	"time"

	"github.com/zxysilent/logs"
)

// The global logger is initialized with New(os.Stderr), which calls hijackstd().
// After that, all stdlib log output is converted to logfmt and respects the
// logs level/caller settings.

// Example: stdlib log is automatically hijacked into logfmt.
func Example_stdlibHijack() {
	stdlog.Println("hello from stdlib log")
	// Output:
}

// Example: stdlib log respects the logs level filter.
func Example_stdlibHijackLevel() {
	logs.SetLevel(logs.LevelWarn)
	stdlog.Println("this INFO message is suppressed")
	logs.SetLevel(logs.LevelInfo) // restore
	// Output:
}

// Example: stdlib log respects the caller setting.
func Example_stdlibHijackCaller() {
	logs.SetCaller(true)
	logs.SetSep("/")
	stdlog.Println("caller visible")
	logs.SetCaller(false) // restore
	// Output:
}

// Example: stdlib prefix set before New becomes the log namespace.
func Example_stdlibHijackPrefix() {
	// Set stdlib prefix before creating the logger
	stdlog.SetPrefix("myprefix")
	l := logs.New(os.Stderr, logs.WithCaller(false)) // hijackstd reads prefix as namespace
	_ = l
	defer stdlog.SetPrefix("") // restore

	stdlog.Println("message with ns")
	// Output:
}

// Example: Logger.Print / Println / Printf mirror stdlib signatures.
func Example_stdlibPrintCompat() {
	l := logs.New(os.Stderr)
	l.Print("a", "b")         // msg=ab
	l.Println("a", "b")       // msg=ab
	l.Printf("%s:%d", "k", 1) // msg=k:1
	// Output:
}

// Example: use the package-level logs config as slog's default handler.
func Example_slogDefault() {
	previous := slog.Default()
	defer slog.SetDefault(previous)

	slog.SetDefault(slog.New(logs.NewSlogHandler()))
	slog.Info("request handled", "method", "GET", "status", 200)
	// Output:
}

// Example: create a slog logger backed by a specific logs Logger.
func Example_slogLogger() {
	l := logs.New(os.Stderr,
		logs.WithCaller(true),
		logs.WithHijack(false),
	)
	logger := slog.New(l.NewSlogHandler())

	logger.Info("request handled",
		slog.String("method", "GET"),
		slog.Group("response", slog.Int("status", 200)),
	) // includes caller=example_stdlib_test.go:<line>
	// Output:
}

// Example: slog records respect the level configured on the logs Logger.
func Example_slogLevels() {
	l := logs.New(os.Stderr,
		logs.WithLevel(logs.LevelWarn),
		logs.WithHijack(false),
	)
	logger := slog.New(l.NewSlogHandler())

	logger.Debug("filtered debug message")
	logger.Warn("visible warning")
	// Output:
}

// Example: With attaches attributes reused by subsequent records.
func Example_slogWith() {
	l := logs.New(os.Stderr, logs.WithHijack(false))
	logger := slog.New(l.NewSlogHandler()).With(
		"service", "payments",
		"version", 2,
	)

	logger.Info("started")
	logger.Info("ready", "port", 8080)
	// Output:
}

// Example: slog groups are encoded as dotted logfmt keys.
func Example_slogGroup() {
	l := logs.New(os.Stderr, logs.WithHijack(false))
	logger := slog.New(l.NewSlogHandler())

	// Group attributes are encoded as request.method and request.path.
	logger.Info("request received",
		slog.Group("request",
			slog.String("method", "GET"),
			slog.String("path", "/health"),
		),
	)
	// Output:
}

// Example: LogAttrs records strongly typed attributes without key/value pairs.
func Example_slogLogAttrs() {
	l := logs.New(os.Stderr, logs.WithHijack(false))
	logger := slog.New(l.NewSlogHandler())

	logger.LogAttrs(context.Background(), slog.LevelInfo, "request completed",
		slog.String("method", "GET"),
		slog.Int("status", 200),
		slog.Bool("cached", true),
		slog.Duration("elapsed", 25*time.Millisecond),
	)
	// Output:
}

type exampleSlogUser struct {
	ID   int
	Name string
}

func (u exampleSlogUser) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Int("id", u.ID),
		slog.String("name", u.Name),
	)
}

// Example: LogValuer controls how a custom value is represented.
func Example_slogLogValuer() {
	l := logs.New(os.Stderr, logs.WithHijack(false))
	logger := slog.New(l.NewSlogHandler())

	// LogValue expands user into user.id and user.name.
	logger.Info("user authenticated", "user", exampleSlogUser{ID: 42, Name: "alice"})
	// Output:
}
