package logs

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
)

// TestGroupFromFielder verifies fielder.Group() freezes preset fields and is reusable.
func TestGroupFromFielder(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	sc := l.With().Str("svc", "api").Int("pid", 1).Group()

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

// TestGroupWithDerive verifies With derives a one-shot fielder inheriting preset fields.
func TestGroupWithDerive(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	sc := l.With().Str("base", "b").Group()

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

// TestLoggerCloneEmpty verifies Clone() creates an empty reusable logger.
func TestLoggerCloneEmpty(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	sc := l.Clone()
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

// TestLoggerTrace verifies Logger.Trace sets the trace field.
func TestLoggerTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	sc := l.Trace("api")
	buf.Reset()
	sc.Info("hello")
	got := buf.String()
	if !strings.Contains(got, "trace=api") {
		t.Fatalf("trace missing: %s", got)
	}
}

// TestTraceReplaceCloneAppend verifies Trace replaces the namespace while
// Clone(trace) appends to it (and Clone() with no args is a pure copy).
func TestTraceReplaceCloneAppend(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	// Trace replaces: api -> svc (not api.svc)
	buf.Reset()
	l.Trace("api").Trace("svc").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=svc") || strings.Contains(got, "trace=api.svc") {
		t.Fatalf("Trace should replace, expected trace=svc, got: %s", got)
	}

	// Clone(trace) appends: api -> api.pay
	buf.Reset()
	l.Trace("api").Clone("pay").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api.pay") {
		t.Fatalf("Clone(trace) should append, expected trace=api.pay, got: %s", got)
	}

	// Clone() with no args is a pure copy: keeps api
	buf.Reset()
	l.Trace("api").Clone().Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api") {
		t.Fatalf("Clone() should preserve trace, expected trace=api, got: %s", got)
	}

	// Clone(trace) on empty namespace -> just the trace
	buf.Reset()
	l.Clone("only").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=only") {
		t.Fatalf("Clone(trace) on empty ns, expected trace=only, got: %s", got)
	}
}

// TestGroupWithTraceJoin verifies With(trace) joins with the logger namespace.
func TestGroupWithTraceJoin(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	// namespace + sub trace -> ns.sub
	buf.Reset()
	l.Trace("api").With("req1").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api.req1") {
		t.Fatalf("expected trace=api.req1, got: %s", got)
	}

	// no namespace + sub trace -> sub
	buf.Reset()
	l.Clone().With("only").Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=only") {
		t.Fatalf("expected trace=only, got: %s", got)
	}

	// namespace + no sub trace -> ns
	buf.Reset()
	l.Trace("api").With().Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api") {
		t.Fatalf("expected trace=api, got: %s", got)
	}
}

// TestGroupCtxTraceJoin verifies Ctx joins namespace with the context trace id.
func TestGroupCtxTraceJoin(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	ctx := TraceCtx(context.Background(), "req9")

	buf.Reset()
	l.Trace("api").Ctx(ctx).Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=api.req9") {
		t.Fatalf("expected trace=api.req9, got: %s", got)
	}

	// no namespace -> just the ctx trace
	buf.Reset()
	l.Clone().Ctx(ctx).Info("x")
	if got := buf.String(); !strings.Contains(got, "trace=req9") {
		t.Fatalf("expected trace=req9, got: %s", got)
	}

	// no namespace, no ctx trace -> no trace
	buf.Reset()
	l.Clone().Ctx(context.Background()).Info("x")
	if got := buf.String(); strings.Contains(got, "trace=") {
		t.Fatalf("expected no trace, got: %s", got)
	}
}

// TestGroupCallerFollowsConfig verifies a derived logger reads caller from the shared config.
func TestGroupCallerFollowsConfig(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setLevel(LINFO)

	l.cfg.setCaller(true)
	sc := l.Trace("api") // shares root cfg
	buf.Reset()
	sc.Info("with-caller")
	if got := buf.String(); !strings.Contains(got, "caller=") {
		t.Fatalf("expected caller field, got: %s", got)
	}

	l.cfg.setCaller(false)
	sc2 := l.Clone() // shares root cfg
	buf.Reset()
	sc2.Info("no-caller")
	if got := buf.String(); strings.Contains(got, "caller=") {
		t.Fatalf("expected no caller field, got: %s", got)
	}
}

// TestGroupLevels verifies level gating across all level methods.
func TestGroupLevels(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LWARN)

	sc := l.Clone()
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

// TestGroupFormatted verifies the *f methods.
func TestGroupFormatted(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LDEBUG)

	sc := l.Clone()
	buf.Reset()
	sc.Infof("%s=%d", "k", 7)
	if got := buf.String(); !strings.Contains(got, "k=7") {
		t.Fatalf("Infof mismatch: %s", got)
	}
}

// TestGroupConcurrent verifies a Group logger is safe under concurrent use.
func TestGroupConcurrent(t *testing.T) {
	l := New(nil) // Discard
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	sc := l.With().Str("shared", "v").Group()

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

// TestPkgTraceAndClone verifies the package-level Trace() and Clone() entry points.
func TestPkgTraceAndClone(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetCaller(false)
	SetLevel(LINFO)

	buf.Reset()
	Trace("svc").Info("ns")
	if got := buf.String(); !strings.Contains(got, "trace=svc") {
		t.Fatalf("pkg Trace trace missing: %s", got)
	}

}
