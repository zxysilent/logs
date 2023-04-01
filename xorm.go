package logs

import (
	xlog "xorm.io/xorm/log"
)

// XLogger
type XLogger struct {
	loger   *Logger
	level   xlog.LogLevel
	showSQL bool
}

var _ xlog.Logger = &XLogger{}

// Error implement ILogger
func (s *XLogger) Error(v ...interface{}) {
	if s.level <= xlog.LOG_ERR {
		s.loger.With().Caller(false).Str("log", "xorm").Error(v...)
	}
}

// Errorf implement ILogger
func (s *XLogger) Errorf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_ERR {
		s.loger.With().Caller(false).Str("log", "xorm").Errorf(format, v...)
	}
}

// Debug implement ILogger
func (s *XLogger) Debug(v ...interface{}) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.With().Caller(false).Str("log", "xorm").Debug(v...)
	}
}

// Debugf implement ILogger
func (s *XLogger) Debugf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.With().Caller(false).Str("log", "xorm").Debugf(format, v...)
	}
}

// Info implement ILogger
func (s *XLogger) Info(v ...interface{}) {
	if s.level <= xlog.LOG_INFO {
		s.loger.With().Caller(false).Str("log", "xorm").Info(v...)
	}
}

// Infof implement ILogger
func (s *XLogger) Infof(format string, v ...interface{}) {
	if s.level <= xlog.LOG_INFO {
		s.loger.With().Caller(false).Str("log", "xorm").Infof(format, v...)
	}
}

// Warn implement ILogger
func (s *XLogger) Warn(v ...interface{}) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.With().Caller(false).Str("log", "xorm").Warn(v...)
	}
}

// Warnf implement ILogger
func (s *XLogger) Warnf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.With().Caller(false).Str("log", "xorm").Warnf(format, v...)
	}
}

// Level implement ILogger
func (s *XLogger) Level() xlog.LogLevel {
	return s.level
}

// SetLevel implement ILogger
func (s *XLogger) SetLevel(l xlog.LogLevel) {
	s.level = l
}

// ShowSQL implement ILogger
func (s *XLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (s *XLogger) IsShowSQL() bool {
	return s.showSQL
}

func (l *Logger) Xorm(lv xlog.LogLevel) *XLogger {
	return &XLogger{
		loger: l,
		level: lv,
	}
}

func Xorm(lv xlog.LogLevel) *XLogger {
	return log.Xorm(lv)
}
