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

// SetSep 设置调用信息分隔符
func SetSep(sep string) {
	log.SetSep(sep)
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
func With() *fieldLogger {
	return log.With()
}

func Ctx(ctx context.Context) *fieldLogger {
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
func Ns(ns string) *NsLogger {
	return &NsLogger{lg: log, ns: ns}
}
