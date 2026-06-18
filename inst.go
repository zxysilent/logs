package logs

import (
	"context"
	"io"
	"os"
)

var log = New(os.Stderr)

func SetLevel(lv logLevel) {
	log.SetLevel(lv)
}

// SetSep 设置调用信息(caller)路径分隔符，可传入多个。
// 输出 caller 时按这些分隔符截断文件路径，取最靠后的匹配段（保留分隔符本身），
// 例如 SetSep("/internal", "/src") 会把绝对路径截到对应模块段。
// 不传参数时保持当前配置不变。
func SetSep(sep ...string) {
	log.SetSep(sep...)
}

// SetCaller
func SetCaller(b bool) {
	log.SetCaller(b)
}

// SetSkip 设置调用信息跳过层数
func SetSkip(skip int) {
	log.SetSkip(skip)
}

// SetOutput 设置输出
func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

func Debug(args ...any) {
	if LDEBUG >= log.level {
		print("", LDEBUG, log.caller, log, nil, args...)
	}
}

func Debugf(format string, args ...any) {
	if LDEBUG >= log.level {
		printf("", LDEBUG, log.caller, log, nil, format, args...)
	}
}

func Info(args ...any) {
	if LINFO >= log.level {
		print("", LINFO, log.caller, log, nil, args...)
	}
}

func Infof(format string, args ...any) {
	if LINFO >= log.level {
		printf("", LINFO, log.caller, log, nil, format, args...)
	}
}

func Warn(args ...any) {
	if LWARN >= log.level {
		print("", LWARN, log.caller, log, nil, args...)
	}
}

func Warnf(format string, args ...any) {
	if LWARN >= log.level {
		printf("", LWARN, log.caller, log, nil, format, args...)
	}
}

func Error(args ...any) {
	if LERROR >= log.level {
		print("", LERROR, log.caller, log, nil, args...)
	}
}

func Errorf(format string, args ...any) {
	if LERROR >= log.level {
		printf("", LERROR, log.caller, log, nil, format, args...)
	}
}

func Print(args ...any) {
	if LINFO >= log.level {
		print("", LINFO, log.caller, log, nil, args...)
	}
}

func Println(args ...any) {
	if LINFO >= log.level {
		print("", LINFO, log.caller, log, nil, args...)
	}
}

func Printf(format string, args ...any) {
	if LINFO >= log.level {
		printf("", LINFO, log.caller, log, nil, format, args...)
	}
}

// With 字段日志
func With(trace ...string) *fielder {
	return log.With(trace...)
}

func Ctx(ctx context.Context) *fielder {
	return log.Ctx(ctx)
}

// Close 关闭日志文件
func Close() error {
	return log.Close()
}

// SetFile 设置日志文件路径
func SetFile(path string) {
	log.SetFile(path)
}

// SetMaxAge 日志最大保存天数
func SetMaxAge(ma int) {
	log.SetMaxAge(ma)
}

// SetMaxSize 单个日志最大容量 MiB
func SetMaxSize(ms int64) {
	log.SetMaxSize(ms)
}

// SetCons 同时输出控制台
func SetCons(b bool) {
	log.SetCons(b)
}

// Ns 命名空间日志
func Ns(ns string) *Scoper {
	return log.Ns(ns)
}

// Scope 空作用域日志
func Scope() *Scoper {
	return log.Scope()
}
