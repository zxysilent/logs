package logs

import (
	"context"
)

// NsLogger 根命名空间日志，ns 不为空时始终作为 trace 输出。
type NsLogger struct {
	lg *Logger
	ns string
}

func (n *NsLogger) Debug(args ...any) {
	if LDEBUG >= n.lg.level {
		print(n.ns, LDEBUG, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Debugf(fmt string, args ...any) {
	if LDEBUG >= n.lg.level {
		printf(n.ns, LDEBUG, n.lg.caller, n.lg, nil, fmt, args...)
	}
}

func (n *NsLogger) Info(args ...any) {
	if LINFO >= n.lg.level {
		print(n.ns, LINFO, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Infof(fmt string, args ...any) {
	if LINFO >= n.lg.level {
		printf(n.ns, LINFO, n.lg.caller, n.lg, nil, fmt, args...)
	}
}

func (n *NsLogger) Warn(args ...any) {
	if LWARN >= n.lg.level {
		print(n.ns, LWARN, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Warnf(fmt string, args ...any) {
	if LWARN >= n.lg.level {
		printf(n.ns, LWARN, n.lg.caller, n.lg, nil, fmt, args...)
	}
}

func (n *NsLogger) Error(args ...any) {
	if LERROR >= n.lg.level {
		print(n.ns, LERROR, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Errorf(fmt string, args ...any) {
	if LERROR >= n.lg.level {
		printf(n.ns, LERROR, n.lg.caller, n.lg, nil, fmt, args...)
	}
}

func (n *NsLogger) Print(args ...any) {
	if LINFO >= n.lg.level {
		print(n.ns, LINFO, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Println(args ...any) {
	if LINFO >= n.lg.level {
		print(n.ns, LINFO, n.lg.caller, n.lg, nil, args...)
	}
}

func (n *NsLogger) Printf(fmt string, args ...any) {
	if LINFO >= n.lg.level {
		printf(n.ns, LINFO, n.lg.caller, n.lg, nil, fmt, args...)
	}
}

func (n *NsLogger) With() *fieldLogger {
	f := getfl()
	f.logger = n.lg
	f.caller = n.lg.caller
	f.trace = n.ns
	f.attr = getb()
	return f
}

func (n *NsLogger) Ctx(ctx context.Context) *fieldLogger {
	f := getfl()
	f.trace, _ = ctx.Value(traceKey).(string)
	f.logger = n.lg
	f.caller = n.lg.caller
	f.attr = getb()
	if f.trace != "" {
		f.trace = n.ns + "\u00b7" + f.trace
	} else {
		f.trace = n.ns
	}
	return f
}
