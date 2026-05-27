package logs

import (
	"context"
	"strings"
	"testing"
)

func TestTrace(t *testing.T) {
	t.Log(trace())
	t.Log(TraceId())
}

func TestTraceId(t *testing.T) {
	id := TraceId()
	if len(id) == 0 {
		t.Fatal("TraceId returned empty")
	}
	// All characters should be from the trace alphabet
	for _, c := range id {
		if !strings.ContainsRune(traceStr, c) {
			t.Fatalf("invalid character in trace id: %c", c)
		}
	}
}

func TestTraceUniqueness(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for i := 0; i < 1000; i++ {
		id := trace()
		if seen[id] {
			t.Fatalf("duplicate trace id generated: %s", id)
		}
		seen[id] = true
	}
}

func BenchmarkTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trace()
	}
}

func BenchmarkTraceId(b *testing.B) {
	for i := 0; i < b.N; i++ {
		TraceId()
	}
}

func TestTraceOf(t *testing.T) {
	ctx := TraceCtx(context.TODO())
	t.Log(TraceOf(ctx))
}

func TestTraceOfEmpty(t *testing.T) {
	// Context without trace key returns empty string
	got := TraceOf(context.Background())
	if got != "" {
		t.Fatalf("expected empty trace for bare context, got: %s", got)
	}
}

func TestTraceCtxAppend(t *testing.T) {
	ctx := TraceCtx(context.Background(), "parent")
	ctx = TraceCtx(ctx, "child")

	got := TraceOf(ctx)
	parts := strings.Split(got, ".")
	if len(parts) != 2 {
		t.Fatalf("expected parent.child, got: %s", got)
	}
	if parts[0] != "parent" || parts[1] != "child" {
		t.Fatalf("expected parent.child, got: %s", got)
	}
}

func TestTraceCtxReuse(t *testing.T) {
	ctx := TraceCtx(context.Background(), "only")
	// Calling again without traceid should reuse
	ctx = TraceCtx(ctx)
	got := TraceOf(ctx)
	if got != "only" {
		t.Fatalf("expected reuse of 'only', got: %s", got)
	}
}

func TestTraceCtxEmptyNew(t *testing.T) {
	// Empty traceid in new context generates a random one
	ctx := TraceCtx(context.Background(), "")
	got := TraceOf(ctx)
	if got == "" {
		t.Fatal("expected generated trace id")
	}
}
