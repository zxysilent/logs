package logs

import (
	"io"

	"github.com/zxysilent/logs/internal/file"
)

// config is the root configuration shared by Logger instances.
// All configuration MUST be finalized before any logging starts;
// concurrent Set* calls during logging are undefined behavior.
// Only the package-level default instance exposes runtime Set* methods —
// these are provided as a convenience for simple applications that configure
// before logging, not for dynamic reconfiguration at runtime.
//
// Set* methods write config fields without synchronization; they must not
// be called concurrently with logging output.
type config struct {
	out    io.Writer
	fw     *file.Writer
	sep    []string // Path separator (take the one furthest to the right in the matching position)
	level  Level
	skip   int
	caller bool
	hijack bool
}

// Option is a functional configuration item for New.
type Option func(*config)

// WithLevel sets the log level.
func WithLevel(lv Level) Option {
	return func(c *config) {
		if lv < LevelDebug || lv > LevelMute {
			panic("illegal logs level")
		}
		c.level = lv
	}
}

// WithCaller sets whether to output caller information.
func WithCaller(b bool) Option {
	return func(c *config) { c.caller = b }
}

// WithHijack sets whether to hijack standard library log output.
func WithHijack(b bool) Option {
	return func(c *config) { c.hijack = b }
}

// WithSep sets the caller path separators, multiple values are allowed.
func WithSep(sep ...string) Option {
	return func(c *config) {
		if len(sep) > 0 {
			c.sep = sep
		}
	}
}

// WithSkip sets the number of caller frames to skip.
func WithSkip(skip int) Option {
	return func(c *config) { c.skip = skip }
}

// FileOption is a file-related configuration item for NewFile.
type FileOption func(*file.Writer)

// WithMaxAge sets the maximum number of days to retain log files.
func WithMaxAge(ma int) FileOption { return func(fw *file.Writer) { fw.SetMaxAge(ma) } }

// WithMaxSize sets the maximum size of a single log file in MiB.
func WithMaxSize(ms int64) FileOption { return func(fw *file.Writer) { fw.SetMaxSize(ms) } }

// WithConsole sets whether to also output to the console.
func WithConsole(b bool) FileOption { return func(fw *file.Writer) { fw.SetConsole(b) } }

// ----- Runtime modification entry (for package-level default instance only) -----

// setLevel sets the log level.
func (c *config) setLevel(lv Level) {
	if lv < LevelDebug || lv > LevelMute {
		panic("illegal logs level")
	}
	c.level = lv
}

// setCaller toggles caller output.
func (c *config) setCaller(b bool) {
	c.caller = b
}

// setSep sets the caller path separators.
func (c *config) setSep(sep ...string) {
	if len(sep) == 0 {
		return
	}
	c.sep = sep
}

// setSkip sets the number of caller frames to skip.
func (c *config) setSkip(skip int) {
	c.skip = skip
}

// setOutput sets the output writer (nil becomes io.Discard).
func (c *config) setOutput(out io.Writer) {
	if out == nil {
		out = io.Discard
	}
	if c.fw != nil {
		c.fw.Close()
		c.fw = nil
	}
	c.out = out
}

// setFile opens a file writer at path and routes output to it.
func (c *config) setFile(path string) {
	if c.fw != nil {
		c.fw.Close()
	}
	c.fw = file.New(path, true)
	c.out = c.fw
}

// setMaxAge sets the file writer's max retention days (no-op without a file writer).
func (c *config) setMaxAge(ma int) {
	if c.fw == nil {
		return
	}
	c.fw.SetMaxAge(ma)
}

// setMaxSize sets the file writer's max size in MiB (no-op without a file writer).
func (c *config) setMaxSize(ms int64) {
	if c.fw == nil {
		return
	}
	c.fw.SetMaxSize(ms)
}

// setConsole toggles console mirroring on the file writer (no-op without a file writer).
func (c *config) setConsole(b bool) {
	if c.fw == nil {
		return
	}
	c.fw.SetConsole(b)
}

// close closes the underlying file writer if any. file.Writer.Close is idempotent (CAS).
func (c *config) close() error {
	if c.fw != nil {
		return c.fw.Close()
	}
	return nil
}
