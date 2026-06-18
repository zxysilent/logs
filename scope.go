package logs

import "context"

// Scoper 持久的作用域日志，预绑定一组公共字段（attr）与可选的 trace/命名空间，
// 可被多次复用且并发安全。通过 fielder.Scope() 或 Logger.Ns() 创建。
// 与 fielder 不同，Scoper 不会在输出后被回收，可长期持有。
type Scoper struct {
	logger *Logger
	attr   []byte // 预设字段，独立存储（不来自池），只读复用
	trace  string // 命名空间 / trace
	caller bool
}

// Scope 将当前 fielder 攒好的字段固化为持久的 ScopeLogger。
// 调用后原 fielder 被释放，不可再使用。
func (fl *fielder) Scope() *Scoper {
	s := &Scoper{
		logger: fl.logger,
		trace:  fl.trace,
		caller: fl.caller,
	}
	if fl.attr != nil && len(*fl.attr) > 0 {
		s.attr = make([]byte, len(*fl.attr))
		copy(s.attr, *fl.attr)
	}
	putfl(fl)
	return s
}

// Ns 基于 Logger 创建一个命名空间作用域，ns 作为 trace 输出，调用信息沿用 Logger 当前配置。
func (l *Logger) Ns(ns string) *Scoper {
	return &Scoper{logger: l, trace: ns, caller: l.caller}
}

// Scope 基于 Logger 创建一个空作用域，可通过 With 追加字段后复用，调用信息沿用 Logger 当前配置。
func (l *Logger) Scope() *Scoper {
	return &Scoper{logger: l, caller: l.caller}
}

// With 基于已有作用域派生一个一次性 fielder，继承预设字段与 trace，可继续追加字段。
// 可选 trace 参数与作用域的命名空间拼接（parent.trace），作用域无命名空间时直接使用该 trace。
func (s *Scoper) With(trace ...string) *fielder {
	f := getfl()
	f.logger = s.logger
	f.caller = s.caller
	f.attr = getb()
	*f.attr = append(*f.attr, s.attr...)
	ntrace := ""
	if len(trace) > 0 {
		ntrace = trace[0]
	}
	f.trace = joinTrace(s.trace, ntrace)
	return f
}

// Ctx 从 context 取出 traceid，与作用域的命名空间拼接后派生一次性 fielder。
func (s *Scoper) Ctx(ctx context.Context) *fielder {
	f := getfl()
	f.logger = s.logger
	f.caller = s.caller
	f.attr = getb()
	*f.attr = append(*f.attr, s.attr...)
	tid, _ := ctx.Value(traceKey).(string)
	f.trace = joinTrace(s.trace, tid)
	return f
}

// joinTrace 拼接作用域命名空间与子 trace：两者都存在时以点号连接，否则取非空的一方。
func joinTrace(base, sub string) string {
	if base != "" && sub != "" {
		return base + "." + sub
	}
	if base != "" {
		return base
	}
	return sub
}

// preb 把预设字段包装成临时 *buffer 供 print/printf 复用（只读，不修改 s.attr）。
func (s *Scoper) preb() *buffer {
	if len(s.attr) == 0 {
		return nil
	}
	b := buffer(s.attr)
	return &b
}

func (s *Scoper) Debug(args ...any) {
	if LDEBUG >= s.logger.level {
		print(s.trace, LDEBUG, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Debugf(format string, args ...any) {
	if LDEBUG >= s.logger.level {
		printf(s.trace, LDEBUG, s.caller, s.logger, s.preb(), format, args...)
	}
}

func (s *Scoper) Info(args ...any) {
	if LINFO >= s.logger.level {
		print(s.trace, LINFO, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Infof(format string, args ...any) {
	if LINFO >= s.logger.level {
		printf(s.trace, LINFO, s.caller, s.logger, s.preb(), format, args...)
	}
}

func (s *Scoper) Warn(args ...any) {
	if LWARN >= s.logger.level {
		print(s.trace, LWARN, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Warnf(format string, args ...any) {
	if LWARN >= s.logger.level {
		printf(s.trace, LWARN, s.caller, s.logger, s.preb(), format, args...)
	}
}

func (s *Scoper) Error(args ...any) {
	if LERROR >= s.logger.level {
		print(s.trace, LERROR, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Errorf(format string, args ...any) {
	if LERROR >= s.logger.level {
		printf(s.trace, LERROR, s.caller, s.logger, s.preb(), format, args...)
	}
}

func (s *Scoper) Print(args ...any) {
	if LINFO >= s.logger.level {
		print(s.trace, LINFO, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Println(args ...any) {
	if LINFO >= s.logger.level {
		print(s.trace, LINFO, s.caller, s.logger, s.preb(), args...)
	}
}

func (s *Scoper) Printf(format string, args ...any) {
	if LINFO >= s.logger.level {
		printf(s.trace, LINFO, s.caller, s.logger, s.preb(), format, args...)
	}
}
