## 简单 golang 结构化日志记录库
>旧版本请使用 `github.com/zxysilent/logs v0.2.1`
- 日志等级 ```DEBUG、INFO、WARN、ERROR```
- 每天切分日志文件
- 默认保留```31```天日志记录(可修改)
- 可同时输出到文件和标准输出
- 单文件大小限制 ```64MB```(可修改)
- 可配置输出等级
- 可配置调用信息
- 可链路追踪
- 适配xorm日志
- 直接使用、维护默认实例
- 可新建日志实例 ```New(io.Writer)```


### 日志结构
``` golang
// logger
type Logger struct {
	out    io.Writer  // 输出
	sep    string     // 路径分隔
	caller bool       // 调用信息
	level  logLevel   // 日志等级
	skip   int        //
	mu     sync.Mutex // logger🔒
	fw     *file.Writer
}
```

### 使用示例
``` golang
package main

import "github.com/zxysilent/logs"

func main() {
	// 使用默认实例
	// 开发环境下设置输出等级为DEBUG，线上环境设置为INFO
	logs.SetLevel(logs.DEBUG)
	// 设置输出调用信息
	logs.SetCaller(true)
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
    // 结构化
    logger.With().
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
    // 链路追踪
    ctx := TraceCtx(context.Background())
	logger.Ctx(ctx).Str("basic", "basic").Debug()
	// ------------------------- 使用自定义实例
	// 适用于不同业务模块
	applog := logs.New(nil)
    applog.SetFile("./logs/applog.log")
    defer applog.Close()
	// 设置日志输出等级
	// 开发环境下设置输出等级为DEBUG，线上环境设置为INFO
	applog.SetLevel(logs.DEBUG)
	// 设置输出调用信息
	applog.SetCallInfo(true)
	// 设置同时显示到控制台
	// 默认只输出到文件
	applog.SetConsole(true)
	applog.Debug("Debug Logger")
	applog.Debugf("Debugf %s", "Logger")

	applog.Info("Info Logger")
	applog.Infof("Infof %s", "Logger")

	applog.Warn("Warn Logger")
	applog.Warnf("Warnf %s", "Logger")

	applog.Error("Error Logger")
	applog.Errorf("Errorf %s", "Logger")
}

```

 ### 性能 

```
pkg: github.com/zxysilent/logs
cpu: 12th Gen Intel(R) Core(TM) i5-12500H
BenchmarkParallel
BenchmarkParallel-16
13344735	        83.23 ns/op	      48 B/op	       1 allocs/op
PASS
ok  	github.com/zxysilent/logs	1.236s
```