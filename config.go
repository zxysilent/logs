package logs

import (
	"io"
	"sync"

	"github.com/zxysilent/logs/internal/file"
)

// config is the root configuration shared by Logger instances.
// Convention: configuration should be finalized before logging starts, and should not be modified during concurrent use.
type config struct {
	mu     sync.Mutex
	out    io.Writer
	fw     *file.Writer
	sep    []string // 路径分隔（取匹配位置最靠后的一个）
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
		if lv < LevelDebug || lv > LevelNone {
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

// setLevel sets the log level under lock.
func (c *config) setLevel(lv Level) {
	if lv < LevelDebug || lv > LevelNone {
		panic("illegal logs level")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.level = lv
}

// setCaller toggles caller output under lock.
func (c *config) setCaller(b bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.caller = b
}

// setSep sets the caller path separators under lock.
func (c *config) setSep(sep ...string) {
	if len(sep) == 0 {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sep = sep
}

// setSkip sets the number of caller frames to skip under lock.
func (c *config) setSkip(skip int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.skip = skip
}

// setOutput sets the output writer under lock (nil becomes io.Discard).
func (c *config) setOutput(out io.Writer) {
	if out == nil {
		out = io.Discard
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.out = out
}

// setFile opens a file writer at path and routes output to it under lock.
func (c *config) setFile(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fw = file.New(path, true)
	c.out = c.fw
}

// setMaxAge sets the file writer's max retention days (no-op without a file writer).
func (c *config) setMaxAge(ma int) {
	if c.fw == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fw.SetMaxAge(ma)
}

// setMaxSize sets the file writer's max size in MiB (no-op without a file writer).
func (c *config) setMaxSize(ms int64) {
	if c.fw == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fw.SetMaxSize(ms)
}

// setConsole toggles console mirroring on the file writer (no-op without a file writer).
func (c *config) setConsole(b bool) {
	if c.fw == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fw.SetConsole(b)
}

// close closes the underlying file writer if any. file.Writer.Close is idempotent (CAS).
func (c *config) close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.fw != nil {
		return c.fw.Close()
	}
	return nil
}
