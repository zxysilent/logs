# logs — Simple & Fast Structured Logging for Go

[中文文档](README.md)

## Features

- **Four log levels**: `DEBUG` `INFO` `WARN` `ERROR`
- **Global default or custom instances**: `logs.New(w)` or package-level functions
- **Structured field chains**: `With().Str("k","v").Int("n",1).Info()`
- **Namespace (Ns)**: `Ns("api").Info()` → `trace=api`
- **Distributed tracing**: `TraceCtx` / `TraceId` / `Ctx`
- **Auto-hijack stdlib** `log`: `New()` converts stdlog → logfmt automatically
- **Stdlib-compatible signatures**: `Print/Printf/Println`
- **File output**: daily rotation, configurable max age/size, optional console mirroring
- **High performance**: zero-allocation fast path, `sync.Pool` buffer reuse

---

## Quick Start

```go
package main

import (
    "context"
    stdlog "log"

    "github.com/zxysilent/logs"
)

func main() {
    // === Default global instance ===
    logs.SetLevel(logs.LDEBUG) // LDEBUG for dev, LINFO for production
    logs.SetCaller(true)
    logs.Info("hello world")

    // === Namespace (Ns) ===
    apiLog := logs.Ns("api")
    apiLog.Info("server started")                  // trace=api
    ctx := logs.TraceCtx(context.Background(), "req-1")
    apiLog.Ctx(ctx).Info("handle")                 // trace=api.req-1

    // === Structured fields ===
    logs.With().
        Str("user", "alice").
        Int("age", 30).
        Info("user login")

    // === Distributed tracing ===
    ctx = logs.TraceCtx(context.Background())
    logs.Ctx(ctx).Str("op", "query").Debug("trace")

    // === Stdlib compatibility ===
    logs.Print("stdlib", "message")                // msg=stdlibmessage
    logs.Printf("stdlib %s", "format")             // msg=stdlib format
    stdlog.Println("auto hijacked to logfmt")      // hijacked by New()

    // === Custom instance (file output) ===
    applog := logs.New(nil)
    applog.SetFile("./logs/app.log")
    applog.SetCons(true)   // also print to stderr
    applog.SetMaxAge(7)    // keep 7 days
    applog.SetMaxSize(64)  // max 64MB per file
    defer applog.Close()
    applog.Info("app started")
}
```

---

## API Reference

### Log Levels

| Constant | Value | Description |
|----------|-------|-------------|
| `logs.LDEBUG` | 0 | Debug |
| `logs.LINFO` | 1 | Info |
| `logs.LWARN` | 2 | Warning |
| `logs.LERROR` | 3 | Error |
| `logs.LNONE` | 4 | Disabled |

### Package-level Functions (operate on default instance)

```go
logs.SetLevel(lv logLevel)                          // set log level
logs.SetCaller(b bool)                              // enable/disable caller line
logs.SetSep(sep ...string)                          // path separators, default "/" (right-most match wins)
logs.SetSkip(skip int)                              // extra caller skip frames
logs.SetOutput(out io.Writer)                       // set output writer
logs.SetFile(path string)                           // set file output
logs.SetMaxAge(ma int)                              // max retention days, default 31
logs.SetMaxSize(ms int64)                           // max file size (MB), default 64
logs.SetCons(b bool)                                // also print to stderr
logs.Close() error                                  // close

// Output
logs.Debug(args ...any)
logs.Debugf(format string, args ...any)
logs.Info(args ...any)
logs.Infof(format string, args ...any)
logs.Warn(args ...any)
logs.Warnf(format string, args ...any)
logs.Error(args ...any)
logs.Errorf(format string, args ...any)

// Stdlib compatibility
logs.Print(args ...any)
logs.Println(args ...any)
logs.Printf(format string, args ...any)

// Field chain / tracing
logs.With(trace ...string) *fielder
logs.Ctx(ctx context.Context) *fielder

// Namespace
logs.Ns(ns string) *ScopeLogger
```

### Logger (custom instance)

```go
l := logs.New(w io.Writer) // nil means Discard
// All package-level functions have corresponding methods
l.SetLevel(lv)  l.SetCaller(b)  l.SetSep(s)  l.SetSkip(n)
l.SetOutput(w)  l.SetFile(p)    l.SetMaxAge(d) l.SetMaxSize(s)
l.SetCons(b)    l.Close()
l.Debug(...)  l.Debugf(...)  l.Info(...)  l.Infof(...)
l.Warn(...)   l.Warnf(...)   l.Error(...) l.Errorf(...)
l.Print(...)  l.Println(...) l.Printf(...)
l.With(trace ...string) *fielder   l.Ctx(ctx) *fielder
l.Ns(ns string) *ScopeLogger
l.Writer() io.Writer     // returns an io.Writer, writes logfmt
```

### ScopeLogger (reusable namespace / field scope)

```go
api := logs.Ns("api")            // *ScopeLogger, trace=api
api.Debug(...)  api.Debugf(...)  api.Info(...)  api.Infof(...)
api.Warn(...)   api.Warnf(...)   api.Error(...) api.Errorf(...)
api.Print(...)  api.Println(...) api.Printf(...)
api.With() *fielder              // derive a one-shot fielder, inherits attr+trace
api.Ctx(ctx context.Context) *fielder
// trace = ns (no ctx) or ns.trace (with ctx)

// Freeze a field chain into a persistent, reusable, concurrency-safe ScopeLogger:
base := logs.With().Str("svc", "api").Int("pid", 1).Scope() // *ScopeLogger
base.Info("started")             // svc=api pid=1, not released, reusable
base.With().Int("uid", 9).Info("login")
```

### fielder (structured fields + output control)

```go
// Fields
fl.Str(key, val string)          fl.Stringer(key string, val fmt.Stringer)
fl.Bytes(key string, val []byte) fl.Err(err error)    fl.IfErr(err error)
fl.Bool(key string, b bool)
fl.Int(key string, i int)        fl.Int8(key, i int8)   fl.Int16(key, i int16)
fl.Int32(key, i int32)           fl.Int64(key, i int64)
fl.Uint(key, i uint)             fl.Uint8(key, i uint8) fl.Uint16(key, i uint16)
fl.Uint32(key, i uint32)         fl.Uint64(key, i uint64)
fl.Float32(key string, f float32) fl.Float64(key string, f float64)
fl.Time(key string, t time.Time) fl.Dur(key string, d time.Duration)
fl.Any(key string, i any)        fl.Raw(key string, b []byte)

// Control
fl.If(b bool)                    // conditional output
fl.Caller(b bool)                // per-entry caller control

// Freeze into a reusable ScopeLogger
fl.Scope() *ScopeLogger               // persist field chain (no manual release)

// Terminal methods (fielder is recycled after call)
fl.Debug(args ...any)   fl.Debugf(format string, args ...any)
fl.Info(args ...any)    fl.Infof(format string, args ...any)
fl.Warn(args ...any)    fl.Warnf(format string, args ...any)
fl.Error(args ...any)   fl.Errorf(format string, args ...any)
```

### Distributed Tracing

```go
ctx := logs.TraceCtx(context.Background())          // generate new trace
ctx := logs.TraceCtx(context.Background(), "myid")  // use specified id
ctx = logs.TraceCtx(ctx, "child")                   // append → myid.child
ctx = logs.TraceCtx(ctx)                            // reuse existing trace
traceId := logs.TraceOf(ctx)                        // read trace
id := logs.TraceId()                                // generate standalone id
```

### Stdlib Integration

```go
// Auto-hijack — New() calls hijackstd(), converting stdlog → logfmt
// prefix is captured as log namespace
stdlog.SetPrefix("myprefix")
_ = logs.New(nil)             // hijack reads prefix → trace=myprefix
stdlog.Println("hello")       // output: trace=myprefix level=INF msg=hello

// Print compat — package-level Print/Printf/Println → logfmt
logs.Print("a", "b")          // msg=ab
logs.Printf("%s:%d", "k", 1)  // msg=k:1

// Writer — returns io.Writer for bridging third-party libraries
w := l.Writer()
w.Write([]byte("raw message"))
```

---

## Output Format (logfmt)

```
time=2026-01-01T12:00:00.000 level=INF msg="hello world"
time=2026-01-01T12:00:00.000 level=INF trace=api.req-1 caller=/main.go:42 user=alice msg=login
time=2026-01-01T12:00:00.000 level=ERR trace=api error="something failed" msg="request failed"
```

- `time` / `level` always present
- `trace` — present when tracing/namespace is used
- `caller` — present when `SetCaller(true)` is set (`file:line`)
- `error` — present when `Err/IfErr` is called

---

## xorm Integration

```go
db.AddHook(&repoHook{showSql: true})

type repoHook struct { showSql bool }

func (rh *repoHook) BeforeProcess(ctx *contexts.ContextHook) (context.Context, error) {
    return ctx.Ctx, nil
}

func (rh *repoHook) AfterProcess(ctx *contexts.ContextHook) error {
    if ctx.Err != nil {
        logs.Ctx(ctx.Ctx).Err(ctx.Err).Str("SQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Error()
    } else if ctx.ExecuteTime > 200*time.Millisecond {
        logs.Ctx(ctx.Ctx).Str("SlowSQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Warn()
    } else if rh.showSql {
        logs.Ctx(ctx.Ctx).Str("SQL", ctx.SQL).
            Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Debug()
    }
    return ctx.Err
}
```

---

## Performance

```
pkg: github.com/zxysilent/logs
cpu: 12th Gen Intel(R) Core(TM) i5-12500H

BenchmarkSimple           95 ns/op,   0 B/op, 0 allocs
BenchmarkInfof           158 ns/op,  16 B/op, 1 allocs
BenchmarkWith5Fields     296 ns/op,   0 B/op, 0 allocs
BenchmarkWith10Fields    539 ns/op,   0 B/op, 0 allocs
BenchmarkError           192 ns/op,   0 B/op, 0 allocs
BenchmarkDisabled        0.5 ns/op,   0 B/op, 0 allocs
BenchmarkParallelSimple   14 ns/op,   0 B/op, 0 allocs
BenchmarkParallel         97 ns/op,   0 B/op, 0 allocs
BenchmarkParallelSpan     80 ns/op,   0 B/op, 0 allocs
```

### Optimizations

- Built-in keys (`time`/`level`/`trace`/`caller`/`msg`) skip quoting via `PutKeyRaw`
- Single-argument type dispatch bypasses `fmt.Sprint`
- `sync.Pool` buffer reuse, zero-allocation fast path

---

## Inspired by

[zerolog](https://github.com/rs/zerolog/)
