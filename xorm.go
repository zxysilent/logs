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
func (s *XLogger) Error(v ...any) {
	if s.level <= xlog.LOG_ERR {
		s.loger.With().Caller(false).Str("caller", "xorm").Error(v...)
	}
}

// Errorf implement ILogger
func (s *XLogger) Errorf(format string, v ...any) {
	if s.level <= xlog.LOG_ERR {
		s.loger.With().Caller(false).Str("caller", "xorm").Errorf(format, v...)
	}
}

// Debug implement ILogger
func (s *XLogger) Debug(v ...any) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.With().Caller(false).Str("caller", "xorm").Debug(v...)
	}
}

// Debugf implement ILogger
func (s *XLogger) Debugf(format string, v ...any) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.With().Caller(false).Str("caller", "xorm").Debugf(format, v...)
	}
}

// Info implement ILogger
func (s *XLogger) Info(v ...any) {
	if s.level <= xlog.LOG_INFO {
		s.loger.With().Caller(false).Str("caller", "xorm").Info(v...)
	}
}

// Infof implement ILogger
func (s *XLogger) Infof(format string, v ...any) {
	if s.level <= xlog.LOG_INFO {
		s.loger.With().Caller(false).Str("caller", "xorm").Infof(format, v...)
	}
}

// Warn implement ILogger
func (s *XLogger) Warn(v ...any) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.With().Caller(false).Str("caller", "xorm").Warn(v...)
	}
}

// Warnf implement ILogger
func (s *XLogger) Warnf(format string, v ...any) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.With().Caller(false).Str("caller", "xorm").Warnf(format, v...)
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
