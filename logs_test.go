package logs

import (
	"context"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

type blackholeStream struct {
	writeCount uint64
}

func (s *blackholeStream) WriteCount() uint64 {
	return atomic.LoadUint64(&s.writeCount)
}

func (s *blackholeStream) Write(p []byte) (int, error) {
	atomic.AddUint64(&s.writeCount, 1)
	return len(p), nil
}
func BenchmarkParallel(b *testing.B) {
	stream := &blackholeStream{}
	logger := New(stream)
	// logger.SetCaller(true)
	// logger.caller = true
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.With().
				Str("rate", "15").
				Int("low", 16).
				Bool("key", true).
				Int8("key2", 2).
				Int16("key3", 3).
				Int32("key4", 4).
				Int64("key5", 5).
				Uint("key", 6).
				Uint8("key", 7).
				Float32("high", 123.2).Info()
		}
	})

	if stream.WriteCount() != uint64(b.N) {
		b.Fatalf("Log write count")
	}
}
func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trace()
	}
}
func TestUUID(t *testing.T) {
	t.Log(trace())
}
func TestLogger(t *testing.T) {
	log.SetCaller(true)
	Info("")
	Info()
}
func BenchmarkLogger(b *testing.B) {
	n := New(nil)
	f := n.With()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Info("info info info info")
	}
}

func TestField(t *testing.T) {
	n := New(os.Stdout)
	f := n.With()
	f.Bool("out", false).
		Bool("key", true).
		Int("key", 1).
		Int8("key", 2).
		Int16("key", 3).
		Int32("key", 4).
		Int64("key", 5).
		Uint("key", 6).
		Uint8("key", 7).
		Uint16("key", 8).
		Uint32("key", 9).
		Uint64("key", 10).
		Float32("key", 11.98122).
		Float64("key", 12.987654321).
		Str("key", "a").
		Bytes("key", []byte("b")).
		Hex("key", []byte{0x1f}).
		Time("key", time.Time{}).
		Dur("key", 0).Any("key-any", runtime.BlockProfileRecord{})
	f.Info()
}

const lines = `
line1
line2
	tab
space	
`

func TestLog(t *testing.T) {
	l := New(os.Stdout)
	l.With().
		Str("k", "xx").
		Bool("out", false).
		Bool("key", true).
		Int("key", 1).
		Int8("key", 2).
		Int16("key", 3).
		Int32("key", 4).
		Int64("key", 5).
		Uint("key", 6).
		Uint8("key", 7).
		Uint16("key", 8).
		Uint32("key", 9).
		Uint64("key", 10).
		Float32("key", 11.98122).
		Float64("key", 12.987654321).
		Str("key", "a").
		Bytes("key", []byte("b")).
		Hex("key", []byte{0x1f}).
		Time("key", time.Now()).
		Dur("key", 0).Any("key-any", runtime.BlockProfileRecord{}).Info("xx")
}
func TestLog1(t *testing.T) {
	SetCaller(true)
	l := New(os.Stdout)
	l.SetCaller(true)
	l.With().Info()
	Info("")
	Errorf("")
	l.With().RawJSON("x", []byte(lines)).Info()
}
func TestLog2(t *testing.T) {
	l := New(os.Stdout)
	l.SetCaller(true)
	ctx := TrackCtx(context.Background(), trace())
	l.Ctx(ctx).Info()
	l.Ctx(ctx).Info()
	l.Ctx(ctx).Str("t", "xx").Str("tx", "tt").Info()
	l.Ctx(ctx).Info()
	l.Ctx(ctx).Error()
	s := l.Ctx(ctx)
	s.Bool("b", false)
	s.Info("666")
	s.Info("xx")
}

// # 使用benchmark采集3秒的内存维度的数据，并生成文件
// go test run=^$ -bench=^BenchmarkZerologJSONNegative$ github.com/zxysilent/logs -benchmem  -benchtime=3s -memprofile=mem_profile.out
// # 采集CPU维度的数据
// go test -benchmem -benchtime=3s -bench=^BenchmarkZerologJSONNegative1$ -cpuprofile=cpu_profile.out1
// # 查看pprof文件，指定http方式查看
// go tool pprof -http="127.0.0.1:8080" mem_profile.out
// go tool pprof -http="127.0.0.1:8080" cpu_profile.out1
// # 查看pprof文件，直接在命令行查看
// go tool pprof mem_profile.out
// go test -benchmem -run=^$ -bench ^BenchmarkZerologJSONNegative$ github.com/zxysilent/logs -count=1 -v -benchtime=3s -cpuprofile=cpu_profile.out
// go tool pprof -http="127.0.0.1:8080" cpu_profile.out
// go test -benchmem -run=^$ -bench ^BenchmarkZerologJSONPositive1$ github.com/zxysilent/logs -count=1 -v -benchtime=3s -cpuprofile=cpu_profile.out1
// go tool pprof -http="127.0.0.1:8080" cpu_profile.out1
