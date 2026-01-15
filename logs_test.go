package logs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestInst(t *testing.T) {
	SetCaller(true)
	SetLevel(LDEBUG)
	SetSep("/")
	SetSkip(1)
	SetOutput(io.Discard)
	SetMaxAge(1)
	SetMaxSize(1024)
	Debug("Debug")
	Debugf("%s", "Debugf")
	Info("Info")
	Infof("%s", "Infof")
	Warn("Warn")
	Warnf("%s", "Warnf")
	Error("Error")
	Errorf("%s", "Errorf")
	Ctx(context.TODO()).Info()
}

func TestBase(t *testing.T) {
	l := New(os.Stdout)
	l.SetCaller(true)
	l.SetLevel(LDEBUG)
	l.Debug("Debug")
	l.Debugf("%s", "Debugf")
	l.Info("Info")
	l.Infof("%s", "Infof")
	l.Warn("Warn")
	l.Warnf("%s", "Warnf")
	l.Error("Error")
	l.Errorf("%s", "Errorf")
}

func TestWithBase(t *testing.T) {
	l := New(os.Stdout)
	l.SetCaller(true)
	ctx := TraceCtx(context.TODO())
	l.Ctx(ctx).Debug("Debug")
	l.Ctx(ctx).Debugf("%s", "Debugf")
	l.Ctx(ctx).Info("Info")
	l.Ctx(ctx).Infof("%s", "Infof")
	l.Ctx(ctx).Warn("Warn")
	l.Ctx(ctx).Warnf("%s", "Warnf")
	l.Ctx(ctx).Error("Error")
	l.Ctx(ctx).Errorf("%s", "Errorf")
	l.Ctx(ctx).If(false).Error("Error")
	l.Ctx(ctx).If(false).Errorf("%s", "Errorf")
}
func TestConfig(t *testing.T) {
	l := New(nil)
	l.SetCaller(true)
	l.SetLevel(LINFO)
	l.SetMaxAge(1)
	l.SetSep("/")
	l.SetSkip(2)
	l.SetMaxSize(1024)
	ctx := TraceCtx(context.TODO())
	l.Ctx(ctx).Debug("Debug")
	l.Ctx(ctx).Debugf("%s", "Debugf")
	l.Ctx(ctx).Info("Info")
	l.Ctx(ctx).Infof("%s", "Infof")
	l.Ctx(ctx).Warn("Warn")
	l.Ctx(ctx).Warnf("%s", "Warnf")
	l.Ctx(ctx).Error("Error")
	l.Ctx(ctx).Errorf("%s", "Errorf")
}

func TestConfigWithFile(t *testing.T) {
	l := New(os.Stdout)
	l.SetFile("./logs/app.log")
	l.SetCaller(true)
	l.SetLevel(LERROR)
	l.SetCons(true)
	l.SetMaxAge(1)
	l.SetMaxSize(1024)
	ctx := TraceCtx(context.TODO())
	l.Ctx(ctx).Debug("Debug")
	l.Ctx(ctx).Debugf("%s", "Debugf")
	l.Ctx(ctx).Info("Info")
	l.Ctx(ctx).Infof("%s", "Infof")
	l.Ctx(ctx).Warn("Warn")
	l.Ctx(ctx).Warnf("%s", "Warnf")
	l.Ctx(ctx).Error("Error")
	l.Ctx(ctx).Errorf("%s", "Errorf")
}

// ---------------------------------------------------------------------------------------------------Parallel
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
				Str("str", "str").
				Int("int", 1025).
				Bool("bool", true).
				Int8("int8", 8).
				Int16("int16", 16).
				Int32("int32", 32).
				Int64("int64", 64).
				Uint("uint", 6).
				Uint8("uin8", 8).
				Err(nil).
				Float32("float32", 3.14).Info()
		}
	})

	if stream.WriteCount() != uint64(b.N) {
		b.Fatalf("Log write count")
	}
}
func BenchmarkLog(b *testing.B) {
	l := New(os.Stdout)
	l.SetFile("./logs/app.log")
	for i := 0; i < b.N; i++ {
		l.Info()
	}
}
func BenchmarkParallelFile(b *testing.B) {
	logger := New(nil)
	logger.SetFile("./logs/app.log")
	// logger.SetCaller(true)
	// logger.caller = true
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.With().
				Str("str", "str").
				Int("int", 1025).
				Bool("bool", true).
				Int("int", 64).
				Int64("int64", 64).Info()
		}
	})
}

type mint int

func (mi mint) String() string {
	return fmt.Sprintf("int:%d", mi)
}

func TestField(t *testing.T) {
	n := New(os.Stdout)
	f := n.With()
	f.Bool("out", false).
		Caller(true).
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
		Err(nil).
		Err(errors.New("err")).
		Raw("key", []byte("")).
		Bytes("key", []byte("b")).
		Time("key", time.Time{}).
		Stringer("key", mint(10)).
		Stringer("key", nil).
		Dur("key", 0).Any("key-any", runtime.BlockProfileRecord{})
	f.Info()
}

func TestLog(t *testing.T) {
	l := New(os.Stdout)
	l.SetCaller(true)
	ctx := TraceCtx(context.Background(), trace())
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

func TestLog1(t *testing.T) {
	l := New(os.Stdout)
	l.SetCaller(true)
	// l.SetFile("./logs1/app.log")
	defer l.Close()
	ctx := TraceCtx(context.Background(), trace())
	l1 := l.Ctx(ctx).Str("basic", "basic")
	l1.Dup().Debug()
	l1.Dup().Info()
	l1.Dup().Error()
	s := l.Ctx(ctx)
	s.Bool("b", false)
	s.Info("666")
	s.Info("xx")
}
func TestWriter(t *testing.T) {
	SetFile("./logs/app.log")
	// SetText()
	SetCons(true)
	SetCaller(true)
	for i := 0; i < 10; i++ {
		With().Int("idx", i).Debug()
		With().Int("idx", i).Debug("debug")
		With().Int("idx", i).Debugf("debugf")
		With().Int("idx", i).Info()
		With().Int("idx", i).Info("info")
		With().Int("idx", i).Infof("infof")
		With().Int("idx", i).Warn()
		With().Int("idx", i).Warn("warn")
		With().Int("idx", i).Warnf("warnf")
		With().Int("idx", i).Error()
		With().Int("idx", i).Error("erro")
		With().Int("idx", i).Errorf("errorf")
	}
	With().Str("idx", "sp ce").Errorf("omit empty")
	Close()
}
func TestSpan(t *testing.T) {
	SetFile("./logs/app.log")
	SetCons(true)
	SetCaller(true)
	ctx := TraceCtx(context.Background())
	n := Ctx(ctx).Str("A", "B").Str("subtrace", "sub")
	defer n.Rel()
	n.Dup().Str("b", "b").Info("xx")
	n.Dup().Str("c", "c").Info("xx")
}

func BenchmarkParallelSpan(b *testing.B) {
	ctx := TraceCtx(context.Background())
	// SetOutput(io.Discard)
	n := Ctx(ctx).Str("A", "B").Str("subtrace", "sub")
	defer n.Rel()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.Dup().Str("b", "b").Info("xx")
			n.Dup().Str("c", "c").Info("xx")
		}
	})
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
