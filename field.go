package logs

import (
	"context"

	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/encoder"
)

type FieldLogger struct {
	ctx    context.Context
	attr   *buffer.Buffer //调用输出后清空
	buf    *buffer.Buffer //每次输出的时候重置
	logger *Logger
	enc    encoder.Encoder
	caller bool
}

func (s *FieldLogger) Caller(b bool) *FieldLogger {
	s.caller = b
	return s
}

func (fl *FieldLogger) Debug(args ...interface{}) {
	if LDEBUG >= fl.logger.level {
		print(fl.ctx, LDEBUG, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Debugf(foramt string, args ...interface{}) {
	if LDEBUG >= fl.logger.level {
		printf(fl.ctx, LDEBUG, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Info(args ...interface{}) {
	if LINFO >= fl.logger.level {
		print(fl.ctx, LINFO, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Infof(foramt string, args ...interface{}) {
	if LINFO >= fl.logger.level {
		printf(fl.ctx, LINFO, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Warn(args ...interface{}) {
	if LWARN >= fl.logger.level {
		print(fl.ctx, LWARN, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Warnf(foramt string, args ...interface{}) {
	if LWARN >= fl.logger.level {
		printf(fl.ctx, LWARN, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}
func (fl *FieldLogger) Error(args ...interface{}) {
	if LERROR >= fl.logger.level {
		print(fl.ctx, LERROR, fl.caller && fl.logger.caller, fl.logger, fl.attr, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}

func (fl *FieldLogger) Errorf(foramt string, args ...interface{}) {
	if LERROR >= fl.logger.level {
		printf(fl.ctx, LERROR, fl.caller && fl.logger.caller, fl.logger, fl.attr, foramt, args...)
		buffer.Put(fl.attr)
		fl.attr = nil
	}
}
