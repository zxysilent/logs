package logs

import (
	"bytes"
	"io"
	stdlog "log"
)

func (l *Logger) HijackLog() {
	stdlog.SetFlags(0)
	ns := stdlog.Prefix()
	stdlog.SetPrefix("")
	stdlog.SetOutput(l.StdWriter(ns))
}

type stdWriter struct {
	logger *Logger
	ns     string
}

func (w *stdWriter) Write(p []byte) (int, error) {
	if w == nil || w.logger == nil {
		return len(p), nil
	}
	if LINFO < w.logger.level {
		return len(p), nil
	}
	msg := bytes.TrimRight(p, "\n")
	nsb := []byte(w.ns)
	if w.ns != "" && bytes.HasPrefix(msg, nsb) {
		msg = bytes.TrimPrefix(msg, nsb)
	}
	printb(w.ns, LINFO, w.logger.caller, w.logger, nil, msg)
	return len(p), nil
}

func (l *Logger) Print(args ...any) {
	if LINFO >= l.level {
		print("", LINFO, l.caller, l, nil, args...)
	}
}

func (l *Logger) Println(args ...any) {
	if LINFO >= l.level {
		print("", LINFO, l.caller, l, nil, args...)
	}
}

func (l *Logger) Printf(foramt string, args ...any) {
	if LINFO >= l.level {
		printf("", LINFO, l.caller, l, nil, foramt, args...)
	}
}

func (l *Logger) StdWriter(ns string) io.Writer {
	return &stdWriter{logger: l, ns: ns}
}

func (l *Logger) Writer() io.Writer {
	return l
}
