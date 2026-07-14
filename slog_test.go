package logs

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSlogHandlerBasic(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithLevel(LevelWarn), WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
	l := slog.New(h)

	l.Info("slow", slog.Duration("elapsed", 2*time.Second+30*time.Millisecond))
	got := buf.String()
	if !strings.Contains(got, "elapsed=2.03s") {
		t.Fatalf("missing duration, got: %s", got)
	}
}

func TestSlogHandlerTime(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
	l := slog.New(h)

	l.Info("metric", "rate", 0.95)
	got := buf.String()
	if !strings.Contains(got, "rate=0.95") {
		t.Fatalf("missing float, got: %s", got)
	}
}

func TestSlogHandlerTypedArgs(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithHijack(false)).NewSlogHandler()
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
	h := New(&buf, WithLevel(LevelDebug), WithHijack(false)).NewSlogHandler()
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
	handler := l.NewSlogHandler()
	sl := slog.New(handler)

	sl.Info("from slog")
	got := buf.String()
	if !strings.Contains(got, `msg="from slog"`) {
		t.Fatalf("missing msg from Logger.SlogHandler, got: %s", got)
	}
}

func TestSlogHandlerLineBreak(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithHijack(false)).NewSlogHandler()
	l := slog.New(h)

	l.Info("first")
	l.Info("second")
	got := buf.String()
	lines := strings.Count(got, "\n")
	if lines != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", lines, got)
	}
}

func TestSlogHandlerCaller(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithCaller(true), WithHijack(false)).NewSlogHandler()
	l := slog.New(h)

	l.Info("with caller")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatalf("missing caller field, got: %s", got)
	}
	if !strings.Contains(got, "slog_test.go") {
		t.Fatalf("caller should point to slog_test.go, got: %s", got)
	}
}

func TestSlogHandlerCallerInherited(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf, WithCaller(true))
	sl := slog.New(l.NewSlogHandler())

	sl.Info("inherited caller")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatalf("missing inherited caller, got: %s", got)
	}
	if !strings.Contains(got, "slog_test.go") {
		t.Fatalf("caller should point to slog_test.go, got: %s", got)
	}
}

func TestDefaultSlogUsesDynamicLogsConfig(t *testing.T) {
	var first bytes.Buffer
	var second bytes.Buffer
	logger := New(&first, WithHijack(false))
	previous := slog.Default()
	slog.SetDefault(slog.New(logger.NewSlogHandler()))
	defer slog.SetDefault(previous)

	slog.Info("first")
	if !strings.Contains(first.String(), "msg=first") {
		t.Fatalf("first output missing: %s", first.String())
	}

	logger.cfg.setOutput(&second)
	logger.cfg.setLevel(LevelWarn)
	slog.Info("filtered")
	if second.Len() != 0 {
		t.Fatalf("info should be filtered after logs level change: %s", second.String())
	}
	slog.Warn("second")
	if !strings.Contains(second.String(), "msg=second") {
		t.Fatalf("updated output missing: %s", second.String())
	}
}

func TestRootNewSlogHandlerUsesPackageConfig(t *testing.T) {
	var buf bytes.Buffer
	previousOut := l.cfg.out
	previousLevel := l.cfg.level
	l.cfg.setOutput(&buf)
	l.cfg.setLevel(LevelInfo)
	defer func() {
		l.cfg.setOutput(previousOut)
		l.cfg.setLevel(previousLevel)
	}()

	logger := slog.New(NewSlogHandler())
	logger.Info("root handler")
	if !strings.Contains(buf.String(), `msg="root handler"`) {
		t.Fatalf("root handler did not use package config: %s", buf.String())
	}
}

func TestIndependentSlogIgnoresLogsConfig(t *testing.T) {
	var buf bytes.Buffer
	h := New(&buf, WithLevel(LevelDebug), WithHijack(false)).NewSlogHandler()
	logger := slog.New(h)
	previousOut := l.cfg.out
	previousLevel := l.cfg.level

	l.cfg.setOutput(io.Discard)
	l.cfg.setLevel(LevelMute)
	defer func() {
		l.cfg.setOutput(previousOut)
		l.cfg.setLevel(previousLevel)
	}()

	logger.Debug("independent")
	if !strings.Contains(buf.String(), "msg=independent") {
		t.Fatalf("independent handler should ignore package config: %s", buf.String())
	}
}

func TestSlogWithAttrsGroupOrdering(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(New(&buf, WithHijack(false)).NewSlogHandler()).
		With("root", 1).
		WithGroup("http").
		With("method", "GET").
		WithGroup("request")

	logger.Info("done", "id", 7)
	got := buf.String()
	for _, want := range []string{"root=1", "http.method=GET", "http.request.id=7"} {
		if !strings.Contains(got, want) {
			t.Fatalf("missing %q in %s", want, got)
		}
	}
	if strings.Contains(got, "http.root=") || strings.Contains(got, "request.method=") {
		t.Fatalf("group applied retroactively: %s", got)
	}
}

type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.b.Write(p)
}

func TestSlogHandlerConcurrent(t *testing.T) {
	var out lockedBuffer
	logger := slog.New(New(&out, WithHijack(false)).NewSlogHandler())
	const workers = 16
	const entries = 100

	var wg sync.WaitGroup
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < entries; i++ {
				logger.Info("parallel", "worker", id, "entry", i)
			}
		}(worker)
	}
	wg.Wait()

	out.mu.Lock()
	lines := strings.Count(out.b.String(), "\n")
	out.mu.Unlock()
	if lines != workers*entries {
		t.Fatalf("got %d lines, want %d", lines, workers*entries)
	}
}

func TestDefaultSlogMuteFiltersCustomHighLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, WithLevel(LevelMute), WithHijack(false))
	h := logger.NewSlogHandler()
	if h.Enabled(context.Background(), slog.Level(LevelMute+1)) {
		t.Fatal("LevelMute must disable every slog level")
	}
}

type slogTestValuer struct{}

func (slogTestValuer) LogValue() slog.Value {
	return slog.StringValue("resolved")
}

func TestSlogHandlerResolvesValuesAndQuotesKeys(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(New(&buf, WithHijack(false)).NewSlogHandler())
	logger.Info("attrs", slog.Any("user name", slogTestValuer{}))

	got := buf.String()
	if !strings.Contains(got, `"user name"=resolved`) {
		t.Fatalf("log valuer or quoted key missing: %s", got)
	}
}

func FuzzSlogHandler(f *testing.F) {
	f.Add("request", "method", "GET", int64(200))
	f.Add("line\nbreak", "key with space", "quoted\"value", int64(-1))
	f.Add("", "", "", int64(0))

	f.Fuzz(func(t *testing.T, msg, key, value string, number int64) {
		var buf bytes.Buffer
		logger := slog.New(New(&buf, WithHijack(false)).NewSlogHandler())

		logger.Info(msg, slog.String(key, value), slog.Int64("number", number))
		logger.With(slog.String(key, value)).Info(msg)
		logger.WithGroup("group").Info(msg, slog.String(key, value))

		got := buf.String()
		if lines := strings.Count(got, "\n"); lines != 3 {
			t.Fatalf("got %d records, want 3: %q", lines, got)
		}
		if levels := strings.Count(got, "level=INF"); levels != 3 {
			t.Fatalf("got %d INFO levels, want 3: %q", levels, got)
		}
		if messages := strings.Count(got, " msg="); messages != 3 {
			t.Fatalf("got %d message fields, want 3: %q", messages, got)
		}
	})
}

type fuzzSlogValuer struct {
	key    string
	value  string
	number int64
}

func (v fuzzSlogValuer) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String(v.key, v.value),
		slog.Int64("number", v.number),
	)
}

func FuzzSlogLogValuer(f *testing.F) {
	f.Add("resolved", "name", "alice", int64(42))
	f.Add("line\nbreak", "key with space", "quoted\"value", int64(-1))
	f.Add("", "", "", int64(0))

	f.Fuzz(func(t *testing.T, msg, key, value string, number int64) {
		var buf bytes.Buffer
		logger := slog.New(New(&buf, WithHijack(false)).NewSlogHandler())

		logger.Info(msg, slog.Any("value", fuzzSlogValuer{
			key:    key,
			value:  value,
			number: number,
		}))

		got := buf.String()
		if lines := strings.Count(got, "\n"); lines != 1 {
			t.Fatalf("got %d records, want 1: %q", lines, got)
		}
		if !strings.Contains(got, "level=INF") || !strings.Contains(got, " msg=") {
			t.Fatalf("incomplete log record: %q", got)
		}
		if !strings.Contains(got, "value.number=") {
			t.Fatalf("LogValuer group was not resolved: %q", got)
		}
	})
}

func BenchmarkSlogBasic(b *testing.B) {
	logger := slog.New(New(io.Discard, WithHijack(false)).NewSlogHandler())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("hello world")
	}
}

func BenchmarkSlogAttrs(b *testing.B) {
	logger := slog.New(New(io.Discard, WithHijack(false)).NewSlogHandler())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("request", "method", "GET", "status", 200, "cached", true)
	}
}

func BenchmarkSlogWithAttrs(b *testing.B) {
	logger := slog.New(New(io.Discard, WithHijack(false)).NewSlogHandler()).With(
		"service", "api",
		"version", 2,
		"production", true,
	)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("request")
	}
}

func BenchmarkSlogGroup(b *testing.B) {
	logger := slog.New(New(io.Discard, WithHijack(false)).NewSlogHandler())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("request", slog.Group("http",
			slog.String("method", "GET"),
			slog.Int("status", 200),
			slog.Bool("cached", true),
		))
	}
}

func BenchmarkSlogCaller(b *testing.B) {
	logger := slog.New(New(io.Discard, WithCaller(true), WithHijack(false)).NewSlogHandler())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("with caller")
	}
}

func BenchmarkSlogFiltered(b *testing.B) {
	logger := slog.New(New(io.Discard, WithLevel(LevelWarn), WithHijack(false)).NewSlogHandler())
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		logger.Info("filtered")
	}
}
