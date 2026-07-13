package logs

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestSlogHandlerBasic(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("hello world")
	got := buf.String()
	if !strings.Contains(got, "level=INF") {
		t.Fatalf("missing level, got: %s", got)
	}
	if !strings.Contains(got, `msg="hello world"`) {
		t.Fatalf("missing msg, got: %s", got)
	}
	if !strings.Contains(got, "time=") {
		t.Fatalf("missing time, got: %s", got)
	}
}

func TestSlogHandlerLevels(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, &SlogHandlerOptions{Level: slog.LevelWarn})
	l := slog.New(h)

	buf.Reset()
	l.Debug("debug")
	if buf.Len() != 0 {
		t.Fatalf("debug should be filtered")
	}

	buf.Reset()
	l.Info("info")
	if buf.Len() != 0 {
		t.Fatalf("info should be filtered at warn level")
	}

	buf.Reset()
	l.Warn("warn")
	if !strings.Contains(buf.String(), "level=WRN") {
		t.Fatalf("warn should pass, got: %s", buf.String())
	}

	buf.Reset()
	l.Error("error")
	if !strings.Contains(buf.String(), "level=ERR") {
		t.Fatalf("error should pass, got: %s", buf.String())
	}
}

func TestSlogHandlerWithAttrs(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.With("user", "alice", "age", 30).Info("login")
	got := buf.String()
	if !strings.Contains(got, "user=alice") {
		t.Fatalf("missing user attr, got: %s", got)
	}
	if !strings.Contains(got, "age=30") {
		t.Fatalf("missing age attr, got: %s", got)
	}
}

func TestSlogHandlerWithGroup(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h.WithGroup("http"))

	l.Info("request", "method", "GET", "status", 200)
	got := buf.String()
	if !strings.Contains(got, "http.method=GET") {
		t.Fatalf("missing grouped attr, got: %s", got)
	}
	if !strings.Contains(got, "http.status=200") {
		t.Fatalf("missing grouped attr, got: %s", got)
	}
}

func TestSlogHandlerDuration(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("slow", slog.Duration("elapsed", 2*time.Second+30*time.Millisecond))
	got := buf.String()
	if !strings.Contains(got, "elapsed=2.03s") {
		t.Fatalf("missing duration, got: %s", got)
	}
}

func TestSlogHandlerTime(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	ts := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	l.Info("event", slog.Time("at", ts))
	got := buf.String()
	if !strings.Contains(got, "at=2026-07-13T12:00:00") {
		t.Fatalf("missing time, got: %s", got)
	}
}

func TestSlogHandlerBool(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("check", "ok", true, "fail", false)
	got := buf.String()
	if !strings.Contains(got, "ok=true") {
		t.Fatalf("missing bool, got: %s", got)
	}
	if !strings.Contains(got, "fail=false") {
		t.Fatalf("missing bool, got: %s", got)
	}
}

func TestSlogHandlerFloat(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("metric", "rate", 0.95)
	got := buf.String()
	if !strings.Contains(got, "rate=0.95") {
		t.Fatalf("missing float, got: %s", got)
	}
}

func TestSlogHandlerTypedArgs(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("stats",
		slog.Int("count", 42),
		slog.Int64("id", 123456789),
		slog.Uint64("uid", 999),
		slog.Float64("ratio", 0.999),
	)
	got := buf.String()
	if !strings.Contains(got, "count=42") {
		t.Fatalf("missing count, got: %s", got)
	}
	if !strings.Contains(got, "id=123456789") {
		t.Fatalf("missing id, got: %s", got)
	}
	if !strings.Contains(got, "uid=999") {
		t.Fatalf("missing uid, got: %s", got)
	}
	if !strings.Contains(got, "ratio=0.999") {
		t.Fatalf("missing ratio, got: %s", got)
	}
}

func TestSlogHandlerAny(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	type payload struct {
		Op   string `json:"op"`
		Code int    `json:"code"`
	}
	l.Info("rpc", slog.Any("body", payload{Op: "query", Code: 200}))
	got := buf.String()
	if !strings.Contains(got, "body=") {
		t.Fatalf("missing any, got: %s", got)
	}
	if !strings.Contains(got, `"op":"query"`) {
		t.Fatalf("missing json field, got: %s", got)
	}
}

func TestSlogHandlerLevelMapping(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, &SlogHandlerOptions{Level: slog.LevelDebug})
	l := slog.New(h)

	buf.Reset()
	l.Debug("dbg")
	if !strings.Contains(buf.String(), "level=DBG") {
		t.Fatalf("debug level missing, got: %s", buf.String())
	}

	buf.Reset()
	l.Info("inf")
	if !strings.Contains(buf.String(), "level=INF") {
		t.Fatalf("info level missing, got: %s", buf.String())
	}

	buf.Reset()
	l.Warn("wrn")
	if !strings.Contains(buf.String(), "level=WRN") {
		t.Fatalf("warn level missing, got: %s", buf.String())
	}

	buf.Reset()
	l.Error("err")
	if !strings.Contains(buf.String(), "level=ERR") {
		t.Fatalf("error level missing, got: %s", buf.String())
	}
}

func TestLoggerSlogHandler(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithLevel(LevelInfo))
	handler := l.SlogHandler()
	sl := slog.New(handler)

	sl.Info("from slog")
	got := buf.String()
	if !strings.Contains(got, `msg="from slog"`) {
		t.Fatalf("missing msg from Logger.SlogHandler, got: %s", got)
	}
}

func TestSlogHandlerLineBreak(t *testing.T) {
	var buf bytes.Buffer
	h := NewSlogHandler(&buf, nil)
	l := slog.New(h)

	l.Info("first")
	l.Info("second")
	got := buf.String()
	lines := strings.Count(got, "\n")
	if lines != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", lines, got)
	}
}
