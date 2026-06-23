package logs

import (
	"bytes"
	"context"
	stdlog "log"
	"os"
	"strings"
	"testing"
)

func TestHijackLog(t *testing.T) {
	var buf bytes.Buffer
	prevOut := l.cfg.out
	prevCaller := l.cfg.caller
	prevLevel := l.cfg.level
	SetOutput(&buf)
	SetCaller(false)
	SetLevel(LINFO)
	defer SetOutput(os.Stderr)
	defer SetCaller(false)
	defer SetLevel(prevLevel)
	defer SetOutput(prevOut)
	defer SetCaller(prevCaller)
	l.hijackstd()
	stdlog.Println("xxxxxxxxxxxxxxx")
	if got := buf.String(); !strings.Contains(got, `msg=xxxxxxxxxxxxxxx`) {
		t.Fatalf("hijack msg mismatch: %s", got)
	}
}

func TestPackagePrintCompat(t *testing.T) {
	var buf bytes.Buffer
	prevOut := l.cfg.out
	prevCaller := l.cfg.caller
	prevLevel := l.cfg.level
	SetOutput(&buf)
	SetCaller(false)
	SetLevel(LINFO)
	defer SetOutput(prevOut)
	defer SetCaller(prevCaller)
	defer SetLevel(prevLevel)

	Print("a", "b")
	Println("a", "b")
	Printf("%s:%d", "a", 1)

	got := buf.String()
	if !strings.Contains(got, `msg=ab`) {
		t.Fatalf("package Print msg mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=ab`) {
		t.Fatalf("package Println msg mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=a:1`) {
		t.Fatalf("package Printf msg mismatch: %s", got)
	}
}

func TestTraceRootTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false)).Trace("myapp")
	l.Info("hello")
	got := buf.String()
	if !strings.Contains(got, "trace=myapp") {
		t.Fatalf("trace root trace missing, got: %s", got)
	}
}

func TestTraceSubTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false)).Trace("myapp")
	ctx := TraceCtx(context.Background(), "req-1")
	l.Ctx(ctx).Info("sub")
	got := buf.String()
	if !strings.Contains(got, "trace=myapp.req-1") {
		t.Fatalf("trace join missing, got: %s", got)
	}
}

func TestTraceEmptyCtx(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false)).Trace("myapp")
	l.Ctx(context.Background()).Info("x")
	got := buf.String()
	if !strings.Contains(got, "trace=myapp") {
		t.Fatalf("trace with empty ctx missing, got: %s", got)
	}
	if strings.Contains(got, "\u00b7") {
		t.Fatalf("unexpected separator with empty ctx, got: %s", got)
	}
}

func TestTraceWithFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false)).Trace("svc")
	l.With().Str("k", "v").Info("x")
	got := buf.String()
	if !strings.Contains(got, "trace=svc") {
		t.Fatalf("trace in With missing, got: %s", got)
	}
}

func TestTraceLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false), WithLevel(LevelWarn)).Trace("svc")
	l.Info("no")
	l.Warn("yes")
	got := buf.String()
	if strings.Contains(got, "no") {
		t.Fatalf("info leaked through WARN")
	}
	if !strings.Contains(got, "trace=svc") {
		t.Fatalf("trace missing, got: %s", got)
	}
}

func TestTracePrint(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false)).Trace("api")
	l.Print("a", "b")
	got := buf.String()
	if !strings.Contains(got, "trace=api") {
		t.Fatalf("print trace missing, got: %s", got)
	}
	if !strings.Contains(got, "msg=ab") {
		t.Fatalf("print msg missing, got: %s", got)
	}
}

func TestTraceNoTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false))
	l.Info("no trace")
	got := buf.String()
	if strings.Contains(got, "trace=") {
		t.Fatalf("trace should not appear, got: %s", got)
	}
}

func TestStdWriterDirectWrite(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false))

	writer := l.stdWriter("ns-x")
	if _, err := writer.Write([]byte("ns-xpayload\n")); err != nil {
		t.Fatalf("write error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, `trace=ns-x`) {
		t.Fatalf("trace mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=payload`) {
		t.Fatalf("payload mismatch: %s", got)
	}
}

func TestStdWriterLevelFilter(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false), WithLevel(LevelError))

	writer := l.stdWriter("ns")
	if _, err := writer.Write([]byte("nspayload\n")); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("expected filtered output, got: %s", got)
	}
}

func TestTraceCallerLine(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(true), WithSkip(0)).Trace("myapp")

	l.Info("caller-line")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	// caller should point to this test file, not internal files
	if strings.Contains(got, "caller=/logger.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
	if !strings.Contains(got, "stdlib_test.go") {
		t.Fatalf("caller should be in stdlib_test.go, got: %s", got)
	}
}

func TestTraceCallerWith(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(true), WithSkip(0)).Trace("myapp")

	l.With().Str("k", "v").Info("caller-with")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/logger.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
	if !strings.Contains(got, "stdlib_test.go") {
		t.Fatalf("caller should be in stdlib_test.go, got: %s", got)
	}
}

func TestTraceCallerCtx(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(true), WithSkip(0)).Trace("myapp")

	ctx := TraceCtx(context.Background(), "req-1")
	l.Ctx(ctx).Info("caller-ctx")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/logger.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
	if !strings.Contains(got, "stdlib_test.go") {
		t.Fatalf("caller should be in stdlib_test.go, got: %s", got)
	}
}

func TestStdLoggerWithTraceCtx(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false))

	ctx := TraceCtx(context.Background(), "trace-ctx")
	l.Ctx(ctx).Info("ctx")
	got := buf.String()
	if !strings.Contains(got, `trace=trace-ctx`) {
		t.Fatalf("context trace mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=ctx`) {
		t.Fatalf("context msg mismatch: %s", got)
	}
}
