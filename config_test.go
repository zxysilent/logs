package logs

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestNewWithOptions verifies New applies functional options at construction.
func TestNewWithOptions(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithLevel(LWARN), WithCaller(true), WithSep("/internal", "/"))

	buf.Reset()
	l.Info("filtered") // below WARN
	if buf.Len() != 0 {
		t.Fatalf("WithLevel(LWARN) should filter Info: %s", buf.String())
	}

	buf.Reset()
	l.Warn("shown")
	got := buf.String()
	if !strings.Contains(got, "msg=shown") {
		t.Fatalf("Warn should pass at WARN level: %s", got)
	}
	if !strings.Contains(got, "caller=") {
		t.Fatalf("WithCaller(true) should emit caller: %s", got)
	}
}

// TestNewDefaultsNilOut verifies New(nil) defaults to Discard and INFO level.
func TestNewDefaultsNilOut(t *testing.T) {
	l := New(nil)
	if l.cfg.level != LINFO {
		t.Fatalf("default level should be LINFO, got %v", l.cfg.level)
	}
	if l.cfg.out != io.Discard {
		t.Fatalf("nil out should default to io.Discard")
	}
	l.Info("no panic")
}

// TestConfigSharedAcrossTrace verifies sub-loggers from Trace/Clone share the root Config:
// changing the root config affects all derived loggers.
func TestConfigSharedAcrossTrace(t *testing.T) {
	var buf bytes.Buffer
	root := New(&buf, WithLevel(LINFO))
	api := root.Trace("api")
	pay := api.Clone("pay") // Clone appends -> api.pay

	// Mutate the shared config via the root's cfg.
	root.cfg.setLevel(LWARN)

	buf.Reset()
	api.Info("api-info") // should now be filtered (shared cfg level=WARN)
	pay.Info("pay-info")
	if buf.Len() != 0 {
		t.Fatalf("sub-loggers should observe shared config level change: %s", buf.String())
	}

	buf.Reset()
	pay.Warn("pay-warn")
	if got := buf.String(); !strings.Contains(got, "trace=api.pay") {
		t.Fatalf("expected nested trace api.pay, got: %s", got)
	}
}

// TestSetFile verifies the package-level SetFile/SetMaxAge/SetMaxSize/SetCons path
// (the runtime-mutable default instance).
func TestSetFile(t *testing.T) {
	// Save & restore the default instance's output so the test is isolated.
	prevOut := l.cfg.out
	defer SetOutput(prevOut)

	dir := t.TempDir()
	SetFile(dir + "/app.log")
	SetMaxAge(7)
	SetMaxSize(2)
	SetCons(false)
	SetLevel(LINFO)
	defer Close()

	if l.cfg.fw == nil {
		t.Fatal("SetFile should set a file writer on the default instance")
	}
	Info("to file")
}

// TestNewFile verifies NewFile returns a Writer plus an idempotent close handle.
func TestNewFile(t *testing.T) {
	dir := t.TempDir()
	w, closeFn := NewFile(dir+"/app.log", WithMaxAge(7), WithMaxSize(2), WithConsole(false))
	l := New(w, WithLevel(LINFO))
	l.Info("hello file")

	// close handle works and is idempotent (multiple calls → nil, no panic).
	if err := closeFn(); err != nil {
		t.Fatalf("first close error: %v", err)
	}
	if err := closeFn(); err != nil {
		t.Fatalf("second close should be a no-op nil, got: %v", err)
	}
	// writer is closed: further writes fail.
	if _, err := w.Write([]byte("after\n")); err == nil {
		t.Fatal("expected error writing to a closed file writer")
	}
}

// TestWithSkip verifies WithSkip sets the caller skip depth.
func TestWithSkip(t *testing.T) {
	l := New(nil, WithSkip(3))
	if l.cfg.skip != 3 {
		t.Fatalf("WithSkip(3) not applied, got %d", l.cfg.skip)
	}
}

// TestWithLevelPanic verifies WithLevel rejects out-of-range levels.
func TestWithLevelPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("WithLevel should panic on illegal level")
		}
	}()
	New(nil, WithLevel(LNONE+1))
}

// TestTraceCopiesAttr verifies Trace/Clone copy frozen preset fields to the child.
func TestTraceCopiesAttr(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	base := l.With().Str("svc", "api").Group() // base has frozen attr svc=api
	child := base.Trace("sub")                 // Trace must copy attr
	buf.Reset()
	child.Info("hi")
	if got := buf.String(); !strings.Contains(got, "svc=api") || !strings.Contains(got, "trace=sub") {
		t.Fatalf("Trace should carry copied attr and trace, got: %s", got)
	}
}

// TestWriterRaw verifies Logger.Write passes bytes straight through to out.

// TestSetOutputNil verifies setOutput(nil) falls back to io.Discard.
func TestSetOutputNil(t *testing.T) {
	l := New(nil)
	l.cfg.setOutput(nil) // nil → io.Discard branch
	if l.cfg.out != io.Discard {
		t.Fatal("setOutput(nil) should default to io.Discard")
	}
	l.Info("no panic")
}

// TestLevelString verifies String() covers standard, intermediate, and OFF levels.
func TestLevelString(t *testing.T) {
	cases := []struct {
		lv   Level
		want string
	}{
		{LDEBUG, "DBG"},
		{LDEBUG + 1, "DBG"}, // -3, still below INFO
		{LINFO, "INF"},
		{LINFO + 2, "INF"}, // intermediate, below WARN
		{LWARN, "WRN"},
		{LWARN + 1, "WRN"},
		{LERROR, "ERR"},
		{LERROR + 100, "ERR"}, // below sentinel
		{LNONE, "OFF"},
	}
	for _, c := range cases {
		if got := c.lv.String(); got != c.want {
			t.Fatalf("Level(%d).String()=%q, want %q", c.lv, got, c.want)
		}
	}
}

// TestWithHijack verifies WithHijack controls hijacking.
func TestWithHijack(t *testing.T) {
	// WithHijack(false): logger constructed without hijacking stdlib
	l := New(nil, WithHijack(false), WithCaller(false), WithLevel(LINFO))
	if l.cfg.hijack {
		t.Fatal("WithHijack(false) should set hijack=false")
	}
	l.Info("no panic") // logger still works

	// WithHijack(true): default behavior, already covered by New()
	l2 := New(nil, WithHijack(true))
	if !l2.cfg.hijack {
		t.Fatal("WithHijack(true) should set hijack=true")
	}
}

// TestSetConsole verifies the package-level SetConsole function.
func TestSetConsole(t *testing.T) {
	prevOut := l.cfg.out
	defer SetOutput(prevOut)

	SetConsole(true) // fw==nil, no-op branch
	// should not panic
}

// TestClone verifies the package-level Clone function.
func TestClone(t *testing.T) {
	var buf bytes.Buffer
	prevOut := l.cfg.out
	prevLevel := l.cfg.level
	defer SetOutput(prevOut)
	defer SetLevel(prevLevel)

	SetOutput(&buf)
	SetLevel(LINFO)
	SetCaller(false)

	child := Clone()
	buf.Reset()
	child.Info("from clone")
	if got := buf.String(); !strings.Contains(got, `msg="from clone"`) {
		t.Fatalf("Clone() should produce a working Logger: %s", got)
	}
}

// TestLoggerClonePresetFields verifies Clone preserves frozen attr.
func TestLoggerClonePresetFields(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(false), WithLevel(LINFO))

	base := l.With().Str("svc", "auth").Group() // base has preset attr
	child := base.Clone()

	buf.Reset()
	child.Info("preserved")
	got := buf.String()
	if !strings.Contains(got, "svc=auth") {
		t.Fatalf("Clone should preserve preset fields: %s", got)
	}
}
