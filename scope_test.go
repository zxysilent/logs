package logs

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
)

// TestScopeFromFielder verifies fielder.Scope() freezes preset fields and is reusable.
func TestScopeFromFielder(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	sc := l.With().Str("svc", "api").Int("pid", 1).Scope()

	buf.Reset()
	sc.Info("first")
	got := buf.String()
	if !strings.Contains(got, "svc=api") || !strings.Contains(got, "pid=1") {
		t.Fatalf("preset fields missing: %s", got)
	}
	if !strings.Contains(got, "msg=first") {
		t.Fatalf("msg missing: %s", got)
	}

	// reusable: preset fields persist across calls
	buf.Reset()
	sc.Error("second")
	got = buf.String()
	if !strings.Contains(got, "svc=api") || !strings.Contains(got, "pid=1") {
		t.Fatalf("preset fields not reused: %s", got)
	}
	if !strings.Contains(got, "level=ERR") {
		t.Fatalf("level mismatch: %s", got)
	}
}

// TestScopeWithDerive verifies With derives a one-shot fielder inheriting preset fields.
func TestScopeWithDerive(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	sc := l.With().Str("base", "b").Scope()

	buf.Reset()
	sc.With().Int("extra", 9).Info("derived")
	got := buf.String()
	if !strings.Contains(got, "base=b") || !strings.Contains(got, "extra=9") {
		t.Fatalf("derived should carry base+extra: %s", got)
	}

	// derived fields must NOT leak back into the scope
	buf.Reset()
	sc.Info("clean")
	got = buf.String()
	if strings.Contains(got, "extra=9") {
		t.Fatalf("derived field leaked into scope: %s", got)
	}
	if !strings.Contains(got, "base=b") {
		t.Fatalf("base field missing: %s", got)
	}
}

// TestLoggerScopeEmpty verifies Logger.Scope() creates an empty reusable scope.
func TestLoggerScopeEmpty(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	sc := l.Scope()
	buf.Reset()
	sc.With().Str("k", "v").Info("msg")
	got := buf.String()
	if !strings.Contains(got, "k=v") {
		t.Fatalf("empty scope With should add field: %s", got)
	}
	if strings.Contains(got, "trace=") {
		t.Fatalf("empty scope should have no trace: %s", got)
	}
}

// TestLoggerNsTrace verifies Logger.Ns sets the trace field.
func TestLoggerNsTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	sc := l.Ns("api")
	buf.Reset()
	sc.Info("hello")
	got := buf.String()
	if !strings.Contains(got, "trace=api") {
		t.Fatalf("Ns trace missing: %s", got)
	}
}

// TestScopeWithTraceJoin verifies With(trace) joins with the scope namespace.
func TestScopeWithTraceJoin(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	// namespace + sub trace -> ns.sub
	buf.Reset()
	l.Ns("api").With("req1").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api.req1") {
		t.Fatalf("expected trace=api.req1, got: %s", got)
	}

	// no namespace + sub trace -> sub
	buf.Reset()
	l.Scope().With("only").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=only") {
		t.Fatalf("expected trace=only, got: %s", got)
	}

	// namespace + no sub trace -> ns
	buf.Reset()
	l.Ns("api").With().Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api") {
		t.Fatalf("expected trace=api, got: %s", got)
	}
}

// TestScopeCtxTraceJoin verifies Ctx joins namespace with the context trace id.
func TestScopeCtxTraceJoin(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LINFO)

	ctx := TraceCtx(context.Background(), "req9")

	buf.Reset()
	l.Ns("api").Ctx(ctx).Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api.req9") {
		t.Fatalf("expected trace=api.req9, got: %s", got)
	}

	// no namespace -> just the ctx trace
	buf.Reset()
	l.Scope().Ctx(ctx).Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=req9") {
		t.Fatalf("expected trace=req9, got: %s", got)
	}

	// no namespace, no ctx trace -> no trace
	buf.Reset()
	l.Scope().Ctx(context.Background()).Info("x")
	if got := buf.String(); strings.Contains(got, "trace=") {
		t.Fatalf("expected no trace, got: %s", got)
	}
}

// TestScopeCallerSnapshot verifies the scope snapshots caller from the logger at creation.
func TestScopeCallerSnapshot(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetLevel(LINFO)

	l.SetCaller(true)
	sc := l.Ns("api") // snapshots caller=true
	buf.Reset()
	sc.Info("with-caller")
	if got := buf.String(); !strings.Contains(got, "caller=") {
		t.Fatalf("expected caller field, got: %s", got)
	}

	l.SetCaller(false)
	sc2 := l.Scope() // snapshots caller=false
	buf.Reset()
	sc2.Info("no-caller")
	if got := buf.String(); strings.Contains(got, "caller=") {
		t.Fatalf("expected no caller field, got: %s", got)
	}
}

// TestScopeLevels verifies level gating across all level methods.
func TestScopeLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LWARN)

	sc := l.Scope()
	buf.Reset()
	sc.Debug("d")
	sc.Info("i")
	if buf.Len() != 0 {
		t.Fatalf("debug/info should be filtered at WARN level: %s", buf.String())
	}
	sc.Warn("w")
	sc.Error("e")
	got := buf.String()
	if !strings.Contains(got, "msg=w") || !strings.Contains(got, "msg=e") {
		t.Fatalf("warn/error should pass: %s", got)
	}
}

// TestScopeFormatted verifies the *f methods.
func TestScopeFormatted(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.SetCaller(false)
	l.SetLevel(LDEBUG)

	sc := l.Scope()
	buf.Reset()
	sc.Infof("%s=%d", "k", 7)
	if got := buf.String(); !strings.Contains(got, "k=7") {
		t.Fatalf("Infof mismatch: %s", got)
	}
}

// TestScopeConcurrent verifies a shared ScopeLogger is safe under concurrent use.
func TestScopeConcurrent(t *testing.T) {
	l := New(nil) // Discard
	l.SetCaller(false)
	l.SetLevel(LINFO)

	sc := l.With().Str("shared", "v").Scope()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			sc.Info("direct")
			sc.With().Int("n", n).Info("derived")
		}(i)
	}
	wg.Wait()
	// Passes if no race/panic; run with -race for full verification.
}

// TestPkgScopeAndNs verifies the package-level Scope() and Ns() entry points.
func TestPkgScopeAndNs(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetCaller(false)
	SetLevel(LINFO)

	buf.Reset()
	Ns("svc").Info("ns")
	if got := buf.String(); !strings.Contains(got, "trace=svc") {
		t.Fatalf("pkg Ns trace missing: %s", got)
	}

	buf.Reset()
	Scope().With().Str("a", "b").Info("scope")
	if got := buf.String(); !strings.Contains(got, "a=b") {
		t.Fatalf("pkg Scope field missing: %s", got)
	}
}
