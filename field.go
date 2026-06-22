package logs

// fielder is a one-time chain builder for accumulating fields.
type fielder struct {
	attr   *buffer //调用输出后清空
	cfg    *config
	trace  string
	caller bool
	skip   bool
}

// Group 将当前 fielder 攒好的字段固化为持久、可复用的 *Logger。调用后原 fielder 被释放，不可再使用。
// Group solidifies the accumulated fields into a persistent *Logger. The original fielder is released and must not be used again.
func (s *fielder) Group() *Logger {
	c := &Logger{cfg: s.cfg, trace: s.trace}
	if s.attr != nil && len(*s.attr) > 0 {
		c.attr = make([]byte, len(*s.attr))
		copy(c.attr, *s.attr)
	}
	putfl(s)
	return c
}

// Caller sets whether to output caller information.
func (s *fielder) Caller(b bool) *fielder {
	s.caller = b
	return s
}

// Debug emits the accumulated fields at debug level, then releases the fielder.
func (fl *fielder) Debug(args ...any) {
	if !fl.skip && LevelDebug >= fl.cfg.level {
		fl.cfg.print(fl.trace, LevelDebug, fl.caller, fl.attr, args...)
	}
	putfl(fl)
}

// Debugf emits the accumulated fields with a formatted message at debug level, then releases the fielder.
func (fl *fielder) Debugf(format string, args ...any) {
	if !fl.skip && LevelDebug >= fl.cfg.level {
		fl.cfg.printf(fl.trace, LevelDebug, fl.caller, fl.attr, format, args...)
	}
	putfl(fl)
}

// Info emits the accumulated fields at info level, then releases the fielder.
func (fl *fielder) Info(args ...any) {
	if !fl.skip && LevelInfo >= fl.cfg.level {
		fl.cfg.print(fl.trace, LevelInfo, fl.caller, fl.attr, args...)
	}
	putfl(fl)
}

// Infof emits the accumulated fields with a formatted message at info level, then releases the fielder.
func (fl *fielder) Infof(format string, args ...any) {
	if !fl.skip && LevelInfo >= fl.cfg.level {
		fl.cfg.printf(fl.trace, LevelInfo, fl.caller, fl.attr, format, args...)
	}
	putfl(fl)
}

// Warn emits the accumulated fields at warn level, then releases the fielder.
func (fl *fielder) Warn(args ...any) {
	if !fl.skip && LevelWarn >= fl.cfg.level {
		fl.cfg.print(fl.trace, LevelWarn, fl.caller, fl.attr, args...)
	}
	putfl(fl)
}

// Warnf emits the accumulated fields with a formatted message at warn level, then releases the fielder.
func (fl *fielder) Warnf(format string, args ...any) {
	if !fl.skip && LevelWarn >= fl.cfg.level {
		fl.cfg.printf(fl.trace, LevelWarn, fl.caller, fl.attr, format, args...)
	}
	putfl(fl)
}

// Error emits the accumulated fields at error level, then releases the fielder.
func (fl *fielder) Error(args ...any) {
	if !fl.skip && LevelError >= fl.cfg.level {
		fl.cfg.print(fl.trace, LevelError, fl.caller, fl.attr, args...)
	}
	putfl(fl)
}

// Errorf emits the accumulated fields with a formatted message at error level, then releases the fielder.
func (fl *fielder) Errorf(format string, args ...any) {
	if !fl.skip && LevelError >= fl.cfg.level {
		fl.cfg.printf(fl.trace, LevelError, fl.caller, fl.attr, format, args...)
	}
	putfl(fl)
}
