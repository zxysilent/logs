---
description: Use this skill when working with github.com/zxysilent/logs — a high-performance structured logging library for Go. Covers Logger creation, structured fields (fielder), namespaces (Trace/Clone), distributed tracing, file output, stdlib hijacking, and configuration options.
globs:
  - "**/*.go"
alwaysApply: false
---

# github.com/zxysilent/logs — Structured Logging for Go

High-performance structured logging library. Four log levels aligned with `log/slog` (-4/0/4/8), zero-allocation field chains, `sync.Pool` buffer reuse, file rotation, and automatic stdlib log hijacking.

## Quick Reference

### Import
```go
import "github.com/zxysilent/logs"
```

### Log Levels
```go
LevelDebug  Level = -4       // Debug
LevelInfo   Level = 0        // Info (default)
LevelWarn   Level = 4        // Warning  
LevelError  Level = 8        // Error
LevelMute   Level = 20241020 // Disables all output
```

### Create a Logger
**Prefer the package-level default instance.** It requires no initialization and the caller skip depth is already correct.

```go
// ✅ Recommended: use the default instance
logs.SetLevel(logs.LevelDebug)
logs.SetCaller(true)
logs.Info("hello")
```

**Custom instances** are immutable after construction. When using a custom instance, caller skip may need adjustment via `WithSkip` if the logger is wrapped inside another helper.

```go
// Custom immutable instance (out=nil defaults to io.Discard)
l := logs.New(w, logs.WithLevel(logs.LevelDebug), logs.WithCaller(true))

// If your custom logger is wrapped in a helper function, add WithSkip(1):
l := logs.New(w, logs.WithCaller(true), logs.WithSkip(1))

// File output: NewFile returns Writer + close handle
w, closeFn := logs.NewFile("./app.log", logs.WithMaxAge(7), logs.WithMaxSize(64), logs.WithConsole(false))
defer closeFn()
l := logs.New(w)
```

### Write Logs
```go
l.Debug(args...)    l.Debugf(format, args...)
l.Info(args...)     l.Infof(format, args...)  
l.Warn(args...)     l.Warnf(format, args...)
l.Error(args...)    l.Errorf(format, args...)
```

### Structured Fields (fielder chain)
```go
// One-shot
l.With().Str("user", "alice").Int("age", 30).Info("login")

// Persistent (reusable)
base := l.With().Str("svc", "api").Int("pid", 1).Group()
base.Info("started")
base.With().Int("step", 1).Info("processing")
```

### Field Types
```go
Str(key, val)  Bool(key, val)  Err(err)  IfErr(err)  
Int(key, val)  Int8/16/32/64  Uint/Uint8/16/32/64  
Float32/64    Time(key, t)    Dur(key, d)
Stringer(key, val)  Any(key, val)  Bytes(key, val)  Raw(key, val)
Caller(bool)  If(bool)
```

### Namespace (Trace/Clone)
```go
api := l.Trace("api")    // Replace namespace: trace=api
pay := api.Clone("pay")  // Append namespace: trace=api.pay
cpy := api.Clone()       // Pure copy: trace=api
```

### Distributed Tracing
```go
ctx := logs.TraceCtx(context.Background())           // Generate random id
ctx := logs.TraceCtx(context.Background(), "myid")   // Use specific id
ctx = logs.TraceCtx(ctx, "child")                     // Append: myid.child
id := logs.TraceOf(ctx)                                // Read trace id
l.Ctx(ctx).Info("traced")                              // Log with trace
```

### Stdlib Compatibility
```go
logs.Print("a", "b")          // msg=ab (joins args)
logs.Printf("%s:%d", "k", 1)  // msg=k:1
// New() automatically hijacks stdlog → logfmt
```

### Parse Level from String
```go
lv := logs.ParseLevel("debug")  // LevelDebug  
lv := logs.ParseLevel("WARN")   // LevelWarn
```

## Configuration Options

### Logger Construction (immutable)
```go
New(w, WithLevel(LevelDebug), WithCaller(true), WithSep("/internal", "/"), WithSkip(0), WithHijack(true))
```

### Package-level (mutable, for simple apps)
```go
SetLevel(LevelDebug)  SetCaller(true)  SetSep("/")  SetSkip(0)
SetOutput(w)  SetFile("./app.log")  SetMaxAge(7)  SetMaxSize(64)
SetConsole(true)  SetTrace("trace-id")  Close()
```

### File Options
```go
NewFile(path, WithMaxAge(days), WithMaxSize(MiB), WithConsole(bool))
```

## Defaults to Know
- **Prefer the package-level default instance** — caller skip is already correct
- Moving from `logs.Info()` to `l.Info()` keeps the same caller depth (method-value binding)
- Level defaults to `LevelInfo` — don't pass `WithLevel(LevelInfo)` unnecessarily
- Caller defaults to `false` in `New()` — explicit `WithCaller(true)` required
- `WithHijack` defaults to `true` — stdlib log is automatically hijacked by `New()`
- `LevelMute` stringifies to `"OFF"`
- **Custom instances**: if you wrap the logger in a helper function, add `WithSkip(1)` so caller points to the actual call site

## Output Format (logfmt)
```
time=2026-01-01T12:00:00.000 level=INF msg="hello world"
time=2026-01-01T12:00:00.000 level=INF trace=api.req-1 caller=/main.go:42 user=alice msg=login
```
