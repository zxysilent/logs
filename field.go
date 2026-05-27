package logs

import (
	"fmt"
	"strings"
)

type fieldLogger struct {
	attr   *buffer //调用输出后清空
	logger *Logger
	trace  string
	caller bool
	skip   bool
}

// Dup duplicate for group field, the caller of Dup need to call Rel、Debug*、Info*、Warn* or Error* to release resources.
func (s *fieldLogger) Dup() *fieldLogger {
	f := getfl()
	f.logger = s.logger
	f.trace = s.trace
	f.caller = s.caller
	f.attr = getb()
	*f.attr = append(*f.attr, *s.attr...)
	return f
}

// Rel release any message
func (s *fieldLogger) Rel() {
	putfl(s)
}

func (s *fieldLogger) Caller(b bool) *fieldLogger {
	s.caller = b
	return s
}

func (fl *fieldLogger) Debug(args ...any) {
	if !fl.skip && LDEBUG >= fl.logger.level {
		print(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Debugf(foramt string, args ...any) {
	if !fl.skip && LDEBUG >= fl.logger.level {
		printf(fl.trace, LDEBUG, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Info(args ...any) {
	if !fl.skip && LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Infof(foramt string, args ...any) {
	if !fl.skip && LINFO >= fl.logger.level {
		printf(fl.trace, LINFO, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Warn(args ...any) {
	if !fl.skip && LWARN >= fl.logger.level {
		print(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Warnf(foramt string, args ...any) {
	if !fl.skip && LWARN >= fl.logger.level {
		printf(fl.trace, LWARN, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}
func (fl *fieldLogger) Error(args ...any) {
	if !fl.skip && LERROR >= fl.logger.level {
		print(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Errorf(foramt string, args ...any) {
	if !fl.skip && LERROR >= fl.logger.level {
		printf(fl.trace, LERROR, fl.caller, fl.logger, fl.attr, foramt, args...)
	}
	putfl(fl)
}

func (fl *fieldLogger) Print(args ...any) {
	if fl != nil && fl.logger != nil && LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, nil, args...)
	}
}

func (fl *fieldLogger) Println(args ...any) {
	if fl != nil && fl.logger != nil && LINFO >= fl.logger.level {
		print(fl.trace, LINFO, fl.caller, fl.logger, nil, strings.TrimSuffix(fmt.Sprintln(args...), "\n"))
	}
}

func (fl *fieldLogger) Printf(foramt string, args ...any) {
	if fl != nil && fl.logger != nil && LINFO >= fl.logger.level {
		printf(fl.trace, LINFO, fl.caller, fl.logger, nil, foramt, args...)
	}
}
