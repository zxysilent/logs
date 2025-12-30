package logs

type fieldLogger struct {
	attr   *buffer //调用输出后清空
	logger *logger
	trace  string
	caller bool
}

// Dup for group field, the caller of Dup need to call Omit、Debug*、Info*、Warn* or Error* to release resources.
func (s *fieldLogger) Dup() *fieldLogger {
	f := getfl()
	f.logger = s.logger
	f.caller = s.caller
	f.attr = getb()
	*f.attr = append(*f.attr, *s.attr...)
	return f
}

// Omit any message
func (s *fieldLogger) Omit() {
	putfl(s)
}

func (s *fieldLogger) Caller(b bool) *fieldLogger {
	s.caller = b
	return s
}

func (fl *fieldLogger) Debug(args ...any) {
	if LDEBUG >= fl.logger.level {
		print(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Debugf(foramt string, args ...any) {
	if LDEBUG >= fl.logger.level {
		printf(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Info(args ...any) {
	if LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Infof(foramt string, args ...any) {
	if LINFO >= fl.logger.level {
		printf(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Warn(args ...any) {
	if LWARN >= fl.logger.level {
		print(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Warnf(foramt string, args ...any) {
	if LWARN >= fl.logger.level {
		printf(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}
func (fl *fieldLogger) Error(args ...any) {
	if LERROR >= fl.logger.level {
		print(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Errorf(foramt string, args ...any) {
	if LERROR >= fl.logger.level {
		printf(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}
