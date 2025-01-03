package logs

import (
	"context"
	"testing"
)

func TestTrace(t *testing.T) {
	t.Log(trace())
	t.Log(TraceId())
}

func BenchmarkTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trace()
	}
}

func TestTraceOf(t *testing.T) {
	ctx := TraceCtx(context.TODO())
	t.Log(TraceOf(ctx))
}
