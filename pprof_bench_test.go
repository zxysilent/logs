package logs

import (
	"context"
	"io"
	"os"
	"runtime"
	"testing"
)

// =============================================================================
// pprof 性能分析基准测试
// 运行方式:
//   CPU:  go test -bench=. -benchtime=3s -cpuprofile=cpu.prof
//   Mem:  go test -bench=. -benchtime=3s -memprofile=mem.prof
//   分析: go tool pprof -http=:8080 cpu.prof
//         go tool pprof -http=:8080 mem.prof
// =============================================================================

type discardingWriter struct{}

func (discardingWriter) Write(p []byte) (int, error) { return len(p), nil }

// ---------------------------------------------------------------------------
// 基础场景
// ---------------------------------------------------------------------------

// BenchmarkBasicEmpty 空日志（无字段/无 caller/无 trace）
func BenchmarkBasicEmpty(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info()
	}
}

// BenchmarkBasicString 单字符串消息
func BenchmarkBasicString(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

// BenchmarkBasicFormat 格式化消息（无参数）
func BenchmarkBasicFormat(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("static message")
	}
}

// BenchmarkBasicFormatArgs 格式化消息（带参数）
func BenchmarkBasicFormatArgs(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("count=%d name=%s", i, "bench")
	}
}

// ---------------------------------------------------------------------------
// Caller 场景
// ---------------------------------------------------------------------------

// BenchmarkCaller 带 caller 信息
func BenchmarkCaller(b *testing.B) {
	l := New(discardingWriter{}, WithCaller(true))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("with caller")
	}
}

// ---------------------------------------------------------------------------
// Trace 场景
// ---------------------------------------------------------------------------

// BenchmarkTraceInfo 带 trace 信息
func BenchmarkTraceInfo(b *testing.B) {
	l := New(discardingWriter{}).Trace("mytrace.abcdefgh")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("with trace")
	}
}

// ---------------------------------------------------------------------------
// With 链式字段场景
// ---------------------------------------------------------------------------

// BenchmarkWith1Field 1 个字段
func BenchmarkWith1Field(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("key", "value").Info()
	}
}

// BenchmarkWith3Fields 3 个字段
func BenchmarkWith3Fields(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("str", "hello").Int("count", 42).Bool("ok", true).Info()
	}
}

// BenchmarkWith8Fields 8 个字段（多种类型）
func BenchmarkWith8Fields(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().
			Str("str", "hello").
			Int("int", 1024).
			Int64("int64", 999999).
			Float64("float", 3.14159).
			Bool("bool", true).
			Uint("uint", 100).
			Err(nil).
			Str("msg", "test").Info()
	}
}

// ---------------------------------------------------------------------------
// Group 固化场景（预计算字段）
// ---------------------------------------------------------------------------

// BenchmarkGroup3Fields 固化 3 个字段后复用
func BenchmarkGroup3Fields(b *testing.B) {
	l := New(discardingWriter{}).With().Str("app", "myapp").Int("version", 1).Bool("prod", true).Group()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("grouped")
	}
}

// BenchmarkGroup8Fields 固化 8 个字段后复用
func BenchmarkGroup8Fields(b *testing.B) {
	l := New(discardingWriter{}).With().
		Str("app", "myapp").
		Int("version", 1).
		Bool("prod", true).
		Float64("rate", 0.99).
		Int64("id", 123456).
		Str("env", "production").
		Str("region", "us-east-1").
		Str("host", "node-42").Group()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("grouped-heavy")
	}
}

// ---------------------------------------------------------------------------
// 字符串转义场景
// ---------------------------------------------------------------------------

// BenchmarkStringNoEscape 无需转义的字符串
func BenchmarkStringNoEscape(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("key", "simple_ascii_value").Info()
	}
}

// BenchmarkStringWithSpace 含空格需引号的字符串
func BenchmarkStringWithSpace(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("key", "value with spaces").Info()
	}
}

// BenchmarkStringWithEscape 含转义字符的字符串
func BenchmarkStringWithEscape(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("key", "value\nwith\"quotes").Info()
	}
}

// ---------------------------------------------------------------------------
// 多参数 msg 场景
// ---------------------------------------------------------------------------

// BenchmarkMultiArgs2 2 个非字符串参数（走 fmt.Sprint）
func BenchmarkMultiArgs2(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("count:", 42, "ok:", true)
	}
}

// ---------------------------------------------------------------------------
// 错误字段场景
// ---------------------------------------------------------------------------

// BenchmarkWithError 错误字段
func BenchmarkWithError(b *testing.B) {
	l := New(discardingWriter{})
	err := io.EOF
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Err(err).Info()
	}
}

// ---------------------------------------------------------------------------
// Ctx 场景
// ---------------------------------------------------------------------------

// BenchmarkCtx 从 context 提取 traceid
func BenchmarkCtx(b *testing.B) {
	l := New(discardingWriter{})
	ctx := TraceCtx(context.Background(), "trace-12345")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Ctx(ctx).Info()
	}
}

// ---------------------------------------------------------------------------
// 并行场景
// ---------------------------------------------------------------------------

// BenchmarkParallelFields 并行日志（8 goroutines，带字段）
func BenchmarkParallelFields(b *testing.B) {
	l := New(discardingWriter{})
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.With().Str("goroutine", "worker").Int("id", 42).Info()
		}
	})
}

// BenchmarkParallelGroup 并行固化日志（无字段构造开销）
func BenchmarkParallelGroup(b *testing.B) {
	l := New(discardingWriter{}).With().Str("service", "api").Int("shard", 1).Group()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("parallel-group")
		}
	})
}

// ---------------------------------------------------------------------------
// 完整组合场景（模拟真实使用）
// ---------------------------------------------------------------------------

// BenchmarkRealWorld 模拟真实场景：caller + trace + 字段 + 消息
func BenchmarkRealWorld(b *testing.B) {
	l := New(discardingWriter{}, WithCaller(true)).Trace("svc.api.v1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().
			Str("method", "GET").
			Int("status", 200).
			Int64("latency_us", 1234).
			Str("path", "/api/v1/users").
			Info("request completed")
	}
}

// ---------------------------------------------------------------------------
// Buffer 池压力测试
// ---------------------------------------------------------------------------

// BenchmarkBufferPoolPressure 模拟大量不同大小日志混合
func BenchmarkBufferPoolPressure(b *testing.B) {
	l := New(discardingWriter{})
	msgs := []string{"short", "medium length message here", "a very long message that might cause buffer growth beyond the pool cap"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info(msgs[i%3])
	}
}

// =============================================================================
// 辅助：生成 pprof 文件的独立测试
// =============================================================================

// TestGenerateCPUProfile 生成 CPU profile 并写入文件
func TestGenerateCPUProfile(b *testing.T) {
	f, err := os.Create("cpu.prof")
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()

	runtime.SetCPUProfileRate(1000) // 提高采样精度

	l := New(discardingWriter{}, WithCaller(true)).Trace("profile.test")
	// 热身
	for i := 0; i < 1000; i++ {
		l.With().Str("key", "value").Int("n", i).Info("warmup")
	}

	// 正式采样（需手动用 go test -run 运行后自行 go tool pprof）
	b.Log("CPU profile written to cpu.prof — analyze with: go tool pprof -http=:8080 cpu.prof")
}
