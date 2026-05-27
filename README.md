## 简单 golang 结构化日志记录库

> 旧版本请使用 `github.com/zxysilent/logs v0.2.1`

-   日志等级 `DEBUG、INFO、WARN、ERROR`
-   每天切分日志文件
-   默认保留`31`天日志记录(可修改)
-   可同时输出到文件和标准输出
-   单文件大小限制 `64MB`(可修改)
-   可配置输出等级
-   可配置调用信息
-   可链路追踪
-   **命名空间 (Ns) 根 trace 支持**
-   适配 xorm 日志
-   兼容标准库 `log.Print/Printf/Println`
-   自动劫持标准库 `log`
-   直接使用、维护默认实例
-   可新建日志实例 `New(io.Writer)`

### 日志结构

```golang
// Logger
type Logger struct {
    out    io.Writer  // 输出
    sep    string     // 路径分隔
    caller bool       // 调用信息
    level  logLevel   // 日志等级
    skip   int        // 额外跳过帧数
    mu     sync.Mutex // logger 锁
    fw     *file.Writer
}
```

### NsLogger — 命名空间日志

```golang
// NsLogger 仅暴露输出方法，ns 始终作为根 trace 输出。
type NsLogger struct {
    lg *Logger
    ns string
}
```

### 使用示例

```golang
package main

import (
    "context"
    stdlog "log"

    "github.com/zxysilent/logs"
)

func main() {
    // 使用默认实例
    // 开发环境下设置输出等级为 LDEBUG，线上环境设置为 LINFO
    logs.SetLevel(logs.LDEBUG)
    // 设置输出调用信息
    logs.SetCaller(true)
    // 直接使用默认实例
    logs.Debug()
    logs.Debug("debug")
    logs.Debugf("debugf")
    logs.Info()
    logs.Info("info")
    logs.Infof("infof")
    logs.Warn()
    logs.Warn("warn")
    logs.Warnf("warnf")
    logs.Error()
    logs.Error("erro")
    logs.Errorf("errorf")

    // 命名空间 — NsLogger
    apiLog := logs.Ns("api")
    apiLog.Info("server started")     // trace=api
    ctx := logs.TraceCtx(context.Background(), "req-1")
    apiLog.Ctx(ctx).Info("handle")    // trace=api·req-1

    // 结构化
    logs.With().
        Str("str", "str").
        Int("int", 1025).
        Bool("bool", true).
        Int8("int8", 8).
        Int16("int16", 16).
        Int32("int32", 32).
        Int64("int64", 64).
        Uint("uint", 6).
        Uint8("uin8", 8).
        Float32("float32", 3.14).Info()

    // 复用结构字段
    d := logs.With().Str("str", "str")
    defer d.Rel()
    d.Dup().Info()
    d.Dup().Warn()

    // 链路追踪
    ctx := logs.TraceCtx(context.Background())
    logs.Ctx(ctx).Str("basic", "basic").Debug()

    // ------------------------- 使用自定义实例
    // 适用于不同业务模块
    applog := logs.New(nil)
    applog.SetFile("./logs/applog.log")
    defer applog.Close()
    // 设置日志输出等级
    // 开发环境下设置输出等级为 LDEBUG，线上环境设置为 LINFO
    applog.SetLevel(logs.LDEBUG)
    // 设置输出调用信息
    applog.SetCaller(true)
    // 设置同时显示到控制台
    // 默认只输出到文件
    applog.SetCons(true)
    applog.Debug("Debug Logger")
    applog.Debugf("Debugf %s", "Logger")

    applog.Info("Info Logger")
    applog.Infof("Infof %s", "Logger")

    applog.Warn("Warn Logger")
    applog.Warnf("Warnf %s", "Logger")

    applog.Error("Error Logger")
    applog.Errorf("Errorf %s", "Logger")

    // ------------------------- 标准库兼容
    // 包级 log.Print/Printf/Println 直接写入当前默认实例
    logs.Print("stdlib", " message")
    logs.Printf("stdlib %s", "format")
    logs.Println("stdlib", "line")

    // stdlog prefix 作为 ns trace，自动劫持
    // HijackLog 在 New() 时自动调用
    stdlog.SetPrefix("[uploader] ")
    stdlog.Println("hello")
}
```

### 命名空间 (Ns)

```golang
// 获取不同命名空间的日志实例
apiLog := logs.Ns("api")
dbLog  := logs.Ns("db")

// NsLogger 与 Logger 共用全局配置（level / caller / output）
apiLog.Info("start")              // trace=api
dbLog.Ctx(ctx).Info("query")      // trace=db·req-1
apiLog.With().Str("k", "v").Info() // trace=api k=v

// Ctx 时 ns 与 trace 用 · 拼接
ctx := logs.TraceCtx(context.Background(), "abc")
logs.Ns("svc").Ctx(ctx).Info()    // trace=svc·abc
```

### 标准库兼容

本库提供标准库兼容入口，输出仍然走当前 `Logger`。

-   `logs.Print/Printf/Println`：默认实例直接调用。
-   `logs.HijackLog()`：劫持 `stdlog` 包级，**`New()` 自动调用**，prefix 作为 trace。
-   `l.StdWriter(ns)`：返回 `io.Writer`，桥接任意需要 `io.Writer` 的库。

#### HijackLog 示例

```golang
package main

import (
    "log"
    "github.com/zxysilent/logs"
)

func main() {
    logs.SetLevel(logs.LINFO)
    // New() 自动调用 HijackLog，后续 stdlog 输出均走本库
    // prefix 自动作为 trace
    log.SetPrefix("[uploader] ")
    log.Println("hello world")
    // 输出: trace=[uploader] level=INF msg="hello world"
}
```

#### StdWriter 示例

```golang
// 绑定到自定义实例
applog := logs.New(nil)
applog.SetLevel(logs.LINFO)
writer := applog.StdWriter("service-ns")
stdlog.New(writer, "service-ns", 0).Println("hello")
// 输出: trace=service-ns level=INF msg=hello
```

### xorm 用例

```golang
db.AddHook(&repoHook{showSql: true})

type repoHook struct {
    showSql bool
}

func (rh *repoHook) BeforeProcess(ctx *contexts.ContextHook) (context.Context, error) {
    return ctx.Ctx, nil
}

func (rh *repoHook) AfterProcess(ctx *contexts.ContextHook) error {
    if ctx.Err != nil {
        logs.Ctx(ctx.Ctx).Caller(false).Err(ctx.Err).Str("SQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Error()
    } else if ctx.ExecuteTime > 200*time.Millisecond {
        logs.Ctx(ctx.Ctx).Caller(false).Str("SlowSQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Warn()
    } else if rh.showSql {
        logs.Ctx(ctx.Ctx).Caller(false).Str("SQL", ctx.SQL).Any("args", ctx.Args).Dur("dur", ctx.ExecuteTime).Debug()
    }
    return ctx.Err
}
```

### 性能

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

### 性能优化细节

- 内置 key（`time`/`level`/`trace`/`caller`/`msg`）使用 `PutKeyRaw` 跳过 quoting 检查
- 单参数输出类型分派（string/int*/uint*/float*/bool/[]byte/`fmt.Stringer`）直接调用 `textenc` 编码器，绕过 `fmt.Sprint`
- 底层 `sync.Pool` 复用 buffer，关键路径 0 分配

## Take ideas from

[zerolog](https://github.com/rs/zerolog/)
