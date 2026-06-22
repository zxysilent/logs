# logs — 简单高效的 golang 结构化日志库

[English](README_EN.md)

## 特性

- **四级日志**：`DEBUG` `INFO` `WARN` `ERROR`（数值对齐 `log/slog`：-4/0/4/8）
- **自建实例或全局默认**：`logs.New(w)` 或直接用包级函数
- **结构化字段链**：`With().Str("k","v").Int("n",1).Info()`
- **命名空间 (Trace)**：`Trace("api").Info()` → `trace=api`
- **链路追踪**：`TraceCtx` / `TraceId` / `Ctx`
- **自动劫持标准库** `log`：`New()` 自动转换 stdlog → logfmt（可用 `WithHijack(false)` 关闭）
- **兼容标准库签名**：`Print/Printf/Println`
- **写入文件**：按天切分，可设最大天数/单文件大小，默认同时输出控制台，也可关闭
- **高性能**：关键路径零分配，`sync.Pool` 复用 buffer

---

## 快速开始

```go
package main

import (
    "context"
    stdlog "log"

    "github.com/zxysilent/logs"
)

func main() {
    // === 默认全局实例 ===
    logs.SetLevel(logs.LevelDebug) // 开发环境 LevelDebug，线上 LevelInfo
    logs.SetCaller(true)
    logs.Info("hello world")

    // === 命名空间 (Trace) ===
    apiLog := logs.Trace("api")
    apiLog.Info("server started")                  // trace=api
    ctx := logs.TraceCtx(context.Background(), "req-1")
    apiLog.Ctx(ctx).Info("handle")                 // trace=api.req-1

    // === 结构化字段 ===
    logs.With().
        Str("user", "alice").
        Int("age", 30).
        Info("user login")

    // === 链路追踪 ===
    ctx = logs.TraceCtx(context.Background())
    logs.Ctx(ctx).Str("op", "query").Debug("trace")

    // === 标准库兼容 ===
    logs.Print("stdlib", "message")                // msg=stdlibmessage
    logs.Printf("stdlib %s", "format")             // msg=stdlib format
    stdlog.Println("auto hijacked to logfmt")      // New() 自动劫持

    // === 自定义实例 (文件输出) ===
    w, closeFn := logs.NewFile("./logs/app.log", logs.MaxAge(7), logs.MaxSize(64), logs.Cons(true))
    defer closeFn()
    applog := logs.New(w, logs.WithLevel(logs.LevelInfo))
    applog.Info("app started")
}
```

---

## API 参考

### 日志等级

数值对齐 `log/slog`（越大越严重）。

| 常量 | 值 | 说明 |
|------|-----|------|
| `logs.LevelDebug` | -4 | 调试 |
| `logs.LevelInfo` | 0 | 常规 |
| `logs.LevelWarn` | 4 | 警告 |
| `logs.LevelError` | 8 | 错误 |
| `logs.LevelMute` | 20241020 | 关闭全部输出（哨兵） |

> 旧名 `LDEBUG` / `LINFO` / `LWARN` / `LERROR` / `LNONE` 已废弃，后续主版本移除。

### 全局函数（操作默认实例）

```go
logs.SetLevel(lv Level)                             // 设置等级
logs.SetCaller(b bool)                              // 开启/关闭调用行号
logs.SetSep(sep ...string)                          // 路径分隔符，默认 "/"（取最靠后的匹配）
logs.SetSkip(skip int)                              // 额外跳帧
logs.SetOutput(out io.Writer)                       // 设置输出
logs.SetFile(path string)                           // 设置文件输出
logs.SetMaxAge(ma int)                              // 最大保留天数，默认 64
logs.SetMaxSize(ms int64)                           // 单文件最大容量(MiB)，默认 64
logs.SetConsole(b bool)                             // 同时输出控制台 (推荐)
logs.Close() error                                  // 关闭

// 输出
logs.Debug(args ...any)
logs.Debugf(format string, args ...any)
logs.Info(args ...any)
logs.Infof(format string, args ...any)
logs.Warn(args ...any)
logs.Warnf(format string, args ...any)
logs.Error(args ...any)
logs.Errorf(format string, args ...any)

// 标准库兼容
logs.Print(args ...any)
logs.Println(args ...any)
logs.Printf(format string, args ...any)

// 字段链 / 追踪
logs.With(trace ...string) *fielder
logs.Ctx(ctx context.Context) *fielder

// 命名空间 / 子 Logger
logs.Trace(trace string) *Logger     // 替换命名空间的子 Logger
logs.Clone(trace ...string) *Logger  // 纯复制（无参）或追加 trace（有参）
```

### Logger（自建实例）

`New` 创建的 Logger 通过函数式选项一次性配置，之后**不可变**（没有 `Set*` 方法）。
如需运行期修改配置，请使用包级默认实例。

```go
// 用选项构造（out 为 nil 时 Discard）
l := logs.New(w,
    logs.WithLevel(logs.LevelDebug),
    logs.WithCaller(true),
    logs.WithSep("/internal", "/"),
    logs.WithSkip(0),
)

// 文件输出：NewFile 返回 Writer + 关闭句柄，可选 MaxAge/MaxSize/Cons
w, closeFn := logs.NewFile("app.log", logs.MaxAge(7), logs.MaxSize(64), logs.Cons(true))
defer closeFn()
fl := logs.New(w)
l.Debug(...)  l.Debugf(...)  l.Info(...)  l.Infof(...)
l.Warn(...)   l.Warnf(...)   l.Error(...) l.Errorf(...)
l.Print(...)  l.Println(...) l.Printf(...)
l.With(trace ...string) *fielder   l.Ctx(ctx) *fielder
l.Trace(trace string) *Logger      // 命名空间子 Logger（共享根配置）
l.Clone(trace ...string) *Logger // 纯复制（无参）或追加 trace（有参），共享根配置
```

### 命名空间 / 子 Logger

`Trace`/`Clone` 派生共享父级根 `Config` 的子 `Logger`。

```go
api := logs.Trace("api")         // *Logger，trace=api（替换）
api.Clone("pay")                 // *Logger，trace=api.pay（追加）
api.Debug(...)  api.Info(...)  api.Warn(...)  api.Error(...)  api.Print(...)
api.With() *fielder              // 派生一次性 fielder，继承 attr+trace
api.Ctx(ctx context.Context) *fielder
// trace = ns（无 ctx）或 ns.trace（有 ctx）

// 将字段链固化为持久、可复用、并发安全的 *Logger：
base := logs.With().Str("svc", "api").Int("pid", 1).Group() // *Logger
base.Info("started")             // svc=api pid=1，不释放，可复用
base.With().Int("uid", 9).Info("login")
```

### fielder（结构化字段 + 输出控制）

```go
// 字段
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

// 控制
fl.If(b bool)                    // 条件输出
fl.Caller(b bool)                // 每条单独控制 caller

// 固化为可复用的 *Logger
fl.Group() *Logger                    // 持久化字段链（无需手动释放）

// 终端方法（调用后 fielder 被回收）
fl.Debug(args ...any)   fl.Debugf(format string, args ...any)
fl.Info(args ...any)    fl.Infof(format string, args ...any)
fl.Warn(args ...any)    fl.Warnf(format string, args ...any)
fl.Error(args ...any)   fl.Errorf(format string, args ...any)
```

### 链路追踪

```go
ctx := logs.TraceCtx(context.Background())          // 生成新 trace
ctx := logs.TraceCtx(context.Background(), "myid")  // 使用指定 id
ctx = logs.TraceCtx(ctx, "child")                   // 追加 → myid.child
ctx = logs.TraceCtx(ctx)                            // 复用已有 trace
traceId := logs.TraceOf(ctx)                        // 读取 trace
id := logs.TraceId()                                // 独立生成 trace id
```

### 标准库集成

```go
// 自动劫持 — New() 调用 hijackstd()，stdlog → logfmt
// prefix 自动作为日志 namespace
stdlog.SetPrefix("myprefix")
_ = logs.New(nil)             // hijack 读取 prefix → trace=myprefix
stdlog.Println("hello")       // output: trace=myprefix level=INF msg=hello

// Print 兼容 — 包级 Print/Printf/Println → logfmt
logs.Print("a", "b")          // msg=ab
logs.Printf("%s:%d", "k", 1)  // msg=k:1
```

---

## 输出格式（logfmt）

```
time=2026-01-01T12:00:00.000 level=INF msg="hello world"
time=2026-01-01T12:00:00.000 level=INF trace=api.req-1 caller=/main.go:42 user=alice msg=login
time=2026-01-01T12:00:00.000 level=ERR trace=api error="something failed" msg="request failed"
```

- `time` / `level` 始终存在
- `trace` — 有链路/命名空间时存在
- `caller` — 开启 `SetCaller(true)` 时存在（`file:line`）
- `error` — 调用 `Err/IfErr` 时存在

---

## xorm 集成

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

## 性能

```
pkg: github.com/zxysilent/logs
cpu: 12th Gen Intel(R) Core(TM) i5-12500H
count: 3 轮平均值

BenchmarkDisabled         1.2 ns/op,   0 B/op, 0 allocs   // 过滤快速路径
BenchmarkParallelSimple    11 ns/op,   0 B/op, 0 allocs   // 并行裸输出
BenchmarkParallelSpan      64 ns/op,   0 B/op, 0 allocs   // 并行 Trace+输出
BenchmarkParallel          58 ns/op,   0 B/op, 0 allocs   // 并行 With 7 字段
BenchmarkSimple            76 ns/op,   0 B/op, 0 allocs   // 基础 Info()
BenchmarkError            139 ns/op,   0 B/op, 0 allocs   // Error 日志
BenchmarkInfof            139 ns/op,  16 B/op, 1 allocs   // 格式化输出
BenchmarkWith5Fields      213 ns/op,   0 B/op, 0 allocs   // 5 个结构化字段
BenchmarkWith10Fields     310 ns/op,   0 B/op, 0 allocs   // 10 个结构化字段
BenchmarkSimpleCaller     491 ns/op,   0 B/op, 0 allocs   // Info + caller
BenchmarkParallelFile     346 ns/op,   0 B/op, 0 allocs   // 并行写入文件
```

### 优化要点

- 内置 key（`time`/`level`/`trace`/`caller`/`msg`）使用 `PutKeyRaw`，跳过 quoting 检查
- 单参数类型分派（string/int*/uint*/float*/bool/[]byte/`fmt.Stringer`），绕过 `fmt.Sprint` 的 interface dispatch
- buffer / fielder 均由 `sync.Pool` 复用，关键路径 0 分配

---

## 灵感来源

[zerolog](https://github.com/rs/zerolog/)

