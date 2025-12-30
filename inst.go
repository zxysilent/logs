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

func SetSep(sep string) {
	log.SetSep(sep)
}

func SetCaller(b bool) {
	log.SetCaller(b)
}

func SetSkip(skip int) {
	log.SetSkip(skip)
}

func SetOutput(out io.Writer) {
	log.SetOutput(out)
}

func Debug(args ...any) {
	if LDEBUG >= log.level {
		print("", LDEBUG, log.caller, log, nil, args...)
	}
}

func Debugf(foramt string, args ...any) {
	if LDEBUG >= log.level {
		printf("", LDEBUG, log.caller, log, nil, foramt, args...)
	}
}

func Info(args ...any) {
	if LINFO >= log.level {
		print("", LINFO, log.caller, log, nil, args...)
	}
}

func Infof(foramt string, args ...any) {
	if LINFO >= log.level {
		printf("", LINFO, log.caller, log, nil, foramt, args...)
	}
}

func Warn(args ...any) {
	if LWARN >= log.level {
		print("", LWARN, log.caller, log, nil, args...)
	}
}

func Warnf(foramt string, args ...any) {
	if LWARN >= log.level {
		printf("", LWARN, log.caller, log, nil, foramt, args...)
	}
}

func Error(args ...any) {
	if LERROR >= log.level {
		print("", LERROR, log.caller, log, nil, args...)
	}
}

func Errorf(foramt string, args ...any) {
	if LERROR >= log.level {
		printf("", LERROR, log.caller, log, nil, foramt, args...)
	}
}

func With() *fieldLogger {
	return log.With()
}

func Ctx(ctx context.Context) *fieldLogger {
	return log.Ctx(ctx)
}

func Close() error {
	return log.Close()
}

func SetFile(path string) {
	log.SetFile(path)
}

func SetMaxAge(ma int) {
	log.SetMaxAge(ma)
}

func SetMaxSize(ms int64) {
	log.SetMaxSize(ms)
}

func SetCons(b bool) {
	log.SetCons(b)
}
