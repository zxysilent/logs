package logs

type fielder struct {
	attr   *buffer //调用输出后清空
	logger *Logger
	trace  string
	caller bool
	skip   bool
}

func (s *fielder) Caller(b bool) *fielder {
	s.caller = b
	return s
}

func (fl *fielder) Debug(args ...any) {
	if !fl.skip && LDEBUG >= fl.logger.level {
		print(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fielder) Debugf(format string, args ...any) {
	if !fl.skip && LDEBUG >= fl.logger.level {
		printf(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, format, args...)
	}
	putfl(fl)
}

func (fl *fielder) Info(args ...any) {
	if !fl.skip && LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fielder) Infof(format string, args ...any) {
	if !fl.skip && LINFO >= fl.logger.level {
		printf(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, format, args...)
	}
	putfl(fl)
}

func (fl *fielder) Warn(args ...any) {
	if !fl.skip && LWARN >= fl.logger.level {
		print(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fielder) Warnf(format string, args ...any) {
	if !fl.skip && LWARN >= fl.logger.level {
		printf(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, format, args...)
	}
	putfl(fl)
}

func (fl *fielder) Error(args ...any) {
	if !fl.skip && LERROR >= fl.logger.level {
		print(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fielder) Errorf(format string, args ...any) {
	if !fl.skip && LERROR >= fl.logger.level {
		printf(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, format, args...)
	}
	putfl(fl)
}
