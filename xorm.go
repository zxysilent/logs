package logs

/*

import (
	"io"

	xlog "xorm.io/xorm/log"
)

// XormLogger is the default implment of ILogger
type XormLogger struct {
	loger   *Logger
	level   xlog.LogLevel
	showSQL bool
}

var _ xlog.Logger = &XormLogger{}

// NewXormLogger let you customrize your logger prefix and flag and logLevel
func NewXormLogger(out io.Writer, l xlog.LogLevel) *XormLogger {
	return &XormLogger{
		loger: New(out),
		level: l,
	}
}

// Error implement ILogger
func (s *XormLogger) Error(v ...interface{}) {
	if s.level <= xlog.LOG_ERR {
		s.loger.Error(v...)
	}
}

// Errorf implement ILogger
func (s *XormLogger) Errorf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_ERR {
		s.loger.Errorf(format, v...)
	}
}

// Debug implement ILogger
func (s *XormLogger) Debug(v ...interface{}) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.Debug(v...)
	}
}

// Debugf implement ILogger
func (s *XormLogger) Debugf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_DEBUG {
		s.loger.Debugf(format, v...)
	}
}

// Info implement ILogger
func (s *XormLogger) Info(v ...interface{}) {
	if s.level <= xlog.LOG_INFO {
		s.loger.Info(v...)
	}
}

// Infof implement ILogger
func (s *XormLogger) Infof(format string, v ...interface{}) {
	if s.level <= xlog.LOG_INFO {
		s.loger.Infof(format, v...)
	}
}

// Warn implement ILogger
func (s *XormLogger) Warn(v ...interface{}) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.Warn(v...)
	}
}

// Warnf implement ILogger
func (s *XormLogger) Warnf(format string, v ...interface{}) {
	if s.level <= xlog.LOG_WARNING {
		s.loger.Warnf(format, v...)
	}
}

// Level implement ILogger
func (s *XormLogger) Level() xlog.LogLevel {
	return s.level
}

// SetLevel implement ILogger
func (s *XormLogger) SetLevel(l xlog.LogLevel) {
	s.level = l
}

// ShowSQL implement ILogger
func (s *XormLogger) ShowSQL(show ...bool) {
	if len(show) == 0 {
		s.showSQL = true
		return
	}
	s.showSQL = show[0]
}

// IsShowSQL implement ILogger
func (s *XormLogger) IsShowSQL() bool {
	return s.showSQL
}

*/
