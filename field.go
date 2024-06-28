package logs

import (
	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/encoder"
)

type FieldLogger struct {
	attr   *buffer.Buffer //调用输出后清空
	buf    *buffer.Buffer //每次输出的时候重置
	logger *Logger
	enc    encoder.Encoder
	trace  string
	caller bool
}

func (s *FieldLogger) Caller(b bool) *FieldLogger {
	s.caller = b
	return s
}

func (fl *FieldLogger) Debug(args ...any) {
	if LDEBUG >= fl.logger.level {
		print(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Debugf(foramt string, args ...any) {
	if LDEBUG >= fl.logger.level {
		printf(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Info(args ...any) {
	if LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Infof(foramt string, args ...any) {
	if LINFO >= fl.logger.level {
		printf(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Warn(args ...any) {
	if LWARN >= fl.logger.level {
		print(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Warnf(foramt string, args ...any) {
	if LWARN >= fl.logger.level {
		printf(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}
func (fl *FieldLogger) Error(args ...any) {
	if LERROR >= fl.logger.level {
		print(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}

func (fl *FieldLogger) Errorf(foramt string, args ...any) {
	if LERROR >= fl.logger.level {
		printf(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	buffer.Put(fl.attr)
	fl.attr = nil
}
