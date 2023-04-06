package logs

import (
	"context"
	"io"
	"os"
)

var log = New(os.Stdout)

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
func Debug(args ...interface{}) {
	if LDEBUG >= log.level {
		print(nil, LDEBUG, log.caller, log, nil, args...)
	}
}

func Debugf(foramt string, args ...interface{}) {
	if LDEBUG >= log.level {
		printf(nil, LDEBUG, log.caller, log, nil, foramt, args...)
	}
}

func Info(args ...interface{}) {
	if LINFO >= log.level {
		print(nil, LINFO, log.caller, log, nil, args...)
	}
}

func Infof(foramt string, args ...interface{}) {
	if LINFO >= log.level {
		printf(nil, LINFO, log.caller, log, nil, foramt, args...)
	}
}

func Warn(args ...interface{}) {
	if LWARN >= log.level {
		print(nil, LWARN, log.caller, log, nil, args...)
	}
}

func Warnf(foramt string, args ...interface{}) {
	if LWARN >= log.level {
		printf(nil, LWARN, log.caller, log, nil, foramt, args...)
	}
}

func Error(args ...interface{}) {
	if LERROR >= log.level {
		print(nil, LERROR, log.caller, log, nil, args...)
	}
}

func Errorf(foramt string, args ...interface{}) {
	if LERROR >= log.level {
		printf(nil, LERROR, log.caller, log, nil, foramt, args...)
	}
}

func With() *FieldLogger {
	return log.With()
}

func Ctx(ctx context.Context) *FieldLogger {
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
