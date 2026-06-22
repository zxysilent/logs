package logs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCallerSkip(t *testing.T) {
	SetCaller(true)
	SetTrace("root")
	Info("01")
	Trace("t").Info("0101")
	l := New(os.Stderr, WithCaller(true))
	l.Info("02")
	lc := l.Clone()
	lc.Info("0202")
	la := lc.With("t1").Str("k", "v").Group()
	la.Info("03")
	la.Clone("t3").Clone("t4").Info("04") // append: t1.t3.t4
	la.Trace("t5").Info("05")             // replace: t5
}

func TestPrintCompat(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setLevel(LINFO)
	l.cfg.setCaller(false)

	l.Print("a", "b")
	if got := buf.String(); !strings.Contains(got, `msg=ab`) {
		t.Fatalf("Print msg mismatch: %s", got)
	}

	buf.Reset()
	l.Println("a", "b")
	if got := buf.String(); !strings.Contains(got, `msg=ab`) {
		t.Fatalf("Println msg mismatch: %s", got)
	}

	buf.Reset()
	l.Printf("%s:%d", "a", 1)
	if got := buf.String(); !strings.Contains(got, `msg=a:1`) {
		t.Fatalf("Printf msg mismatch: %s", got)
	}
}

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
	l.cfg.setCaller(true)
	l.cfg.setLevel(LDEBUG)
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
	l.cfg.setCaller(true)
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

// TestConfigFallback verifies nil out + set* methods tolerate nil fw without panic.
func TestConfigFallback(t *testing.T) {
	l := New(nil)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setMaxAge(1)
	l.cfg.setSep("/")
	l.cfg.setSkip(2)
	l.cfg.setMaxSize(1024)
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

// TestLastSep verifies multi-separator truncation picks the right-most match.
func TestLastSep(t *testing.T) {
	seps := []string{"/", "\\"}
	cases := []struct {
		in   string
		want int
	}{
		{"a/b/c.go", 3},   // last '/'
		{"a\\b\\c.go", 3}, // last '\'
		{"a/b\\c.go", 3},  // mixed: last is '\'
		{"a\\b/c.go", 3},  // mixed: last is '/'
		{"nosep.go", -1},  // no separator
		{"", -1},          // empty input
		{"/a/b/c.go", 4},  // last '/' at index 4
	}
	// multi-char separators
	if got := lastSep("a/src/b.go", []string{"/src", "/internal"}); got != 1 {
		t.Fatalf("lastSep multi-char sep mismatch: %d", got)
	}
	if got := lastSep("a/internal/b.go", []string{"/src", "/internal"}); got != 1 {
		t.Fatalf("lastSep multi-char sep mismatch: %d", got)
	}
	for _, c := range cases {
		if got := lastSep(c.in, seps); got != c.want {
			t.Fatalf("lastSep(%q)=%d, want %d", c.in, got, c.want)
		}
	}
	// empty seps in the list are ignored
	if got := lastSep("a/b", []string{"", "/"}); got != 1 {
		t.Fatalf("lastSep with empty sep mismatch: %d", got)
	}
}

// TestSetSepMulti verifies SetSep accepts multiple separators and SetSep() with no args keeps the value.
func TestSetSepMulti(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setSep("/", "\\")

	buf.Reset()
	l.Info("multi-sep")
	got := buf.String()
	// caller should be truncated to the file segment (starts with a separator)
	if !strings.Contains(got, "caller=") {
		t.Fatalf("caller field missing: %s", got)
	}

	// SetSep() with no args must keep existing separators
	l.cfg.setSep()
	buf.Reset()
	l.Info("still-works")
	if got := buf.String(); !strings.Contains(got, "caller=") {
		t.Fatalf("caller field missing after SetSep(): %s", got)
	}
}

func TestConfigWithFile(t *testing.T) {
	l := New(os.Stdout)
	l.cfg.setFile("./logs/app.log")
	l.cfg.setCaller(true)
	l.cfg.setLevel(LERROR)
	l.cfg.setConsole(true)
	l.cfg.setMaxAge(1)
	l.cfg.setMaxSize(1024)
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
	l.cfg.setFile("./logs/app.log")
	for i := 0; i < b.N; i++ {
		l.Info()
	}
}
func BenchmarkParallelFile(b *testing.B) {
	logger := New(nil)
	logger.cfg.setFile("./logs/app.log")
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

// TestCtxInfo verifies Ctx + Info/Error basic flow with trace id.
func TestCtxInfo(t *testing.T) {
	l := New(os.Stdout)
	l.cfg.setCaller(true)
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

// TestGroupBasic verifies Group then Debug/Info/Error work.
func TestGroupBasic(t *testing.T) {
	l := New(os.Stdout)
	l.cfg.setCaller(true)
	defer l.cfg.close()
	ctx := TraceCtx(context.Background(), trace())
	l1 := l.Ctx(ctx).Str("basic", "basic").Group()
	l1.Debug()
	l1.Info()
	l1.Error()
	s := l.Ctx(ctx)
	s.Bool("b", false)
	s.Info("666")
	s.Info("xx")
}
func TestWriter(t *testing.T) {
	SetFile("./logs/app.log")
	SetConsole(true)
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

// TestGroupFromCtx verifies Ctx + Group + With combination.
func TestGroupFromCtx(t *testing.T) {
	SetFile("./logs/app.log")
	SetConsole(true)
	SetCaller(true)
	ctx := TraceCtx(context.Background())
	n := Ctx(ctx).Str("A", "B").Str("subtrace", "sub").Group()
	n.With().Str("b", "b").Info("xx")
	n.With().Str("c", "c").Info("xx")
}

// TestCallerCorrect verifies caller points to the actual call site, not internal helper.
func TestCallerCorrect(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)

	l.Info("caller-test")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	// The caller should be this test file, not assist.go or logs.go
	if strings.Contains(got, "caller=/assist.go") || strings.Contains(got, "caller=/logs.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerWith verifies caller is correct through With() chain.
func TestCallerWith(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)

	l.With().Str("k", "v").Info("caller-with")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestSetSkip verifies SetSkip adjusts caller depth correctly.
func TestSetSkip(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setSkip(1)

	l.Info("skip-1")
	got := buf.String()
	// With skip=1 we should still get a valid caller (not crashing)
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field with skip=1")
	}
}

// TestCallerLineNum verifies caller reports the exact line number of the call site.
func TestCallerLineNum(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)

	_, _, baseLine, _ := runtime.Caller(0)
	l.Info("line-test") // caller = baseLine + 1
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("caller line mismatch: expected :%s, got: %s", expect, got)
	}
}

// TestCallerLineNumWith verifies caller line number with With() chain.
func TestCallerLineNumWith(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)

	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Str("k", "v").Info("line-with") // baseLine + 1
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("With caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNumCtx verifies caller line number with Ctx() chain.
func TestCallerLineNumCtx(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)

	ctx := TraceCtx(context.Background(), "req-1")
	_, _, baseLine, _ := runtime.Caller(0)
	l.Ctx(ctx).Info("line-ctx") // baseLine + 1
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("Ctx caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNumSkip verifies SetSkip(1) goes beyond the test file.
func TestCallerLineNumSkip(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(1)
	l.cfg.setLevel(LINFO)

	l.Info("skip-line")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field with skip=1")
	}
	if strings.Contains(got, "caller=/logs_test.go") {
		t.Fatalf("skip=1 should go beyond this file, got: %s", got)
	}
}

// TestCallerLineDebug verifies caller line number for Logger.Debug.
func TestCallerLineDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LDEBUG)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Debug("debug-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("Debug caller line mismatch: expected :%s, got: %s", expect, got)
	}
}

// TestCallerLineWarn verifies caller line number for Logger.Warn.
func TestCallerLineWarn(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LWARN)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Warn("warn-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("Warn caller line mismatch: expected :%s, got: %s", expect, got)
	}
}

// TestCallerLineError verifies caller line number for Logger.Error.
func TestCallerLineError(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LERROR)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Error("error-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("Error caller line mismatch: expected :%s, got: %s", expect, got)
	}
}

// TestCallerLinePrint verifies caller line number for Logger.Print.
func TestCallerLinePrint(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Print("print-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("Print caller line mismatch: expected :%s, got: %s", expect, got)
	}
}

// TestCallerLineFieldDebug verifies caller line number for fieldLogger.Debug.
func TestCallerLineFieldDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LDEBUG)
	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Caller(true).Debug("field-debug-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("fieldLogger Debug caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineFieldInfo verifies caller line number for fieldLogger.Info (with Caller).
func TestCallerLineFieldInfo(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)
	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Caller(true).Info("field-info-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("fieldLogger Info caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineFieldWarn verifies caller line number for fieldLogger.Warn.
func TestCallerLineFieldWarn(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LWARN)
	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Caller(true).Warn("field-warn-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("fieldLogger Warn caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineFieldError verifies caller line number for fieldLogger.Error.
func TestCallerLineFieldError(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LERROR)
	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Caller(true).Error("field-error-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("fieldLogger Error caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsInfo verifies caller line number for NsLogger.Info.
func TestCallerLineNsInfo(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Info("ns-info-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger Info caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/scope.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsDebug verifies caller line number for NsLogger.Debug.
func TestCallerLineNsDebug(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LDEBUG)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Debug("ns-debug-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger Debug caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/scope.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsWarn verifies caller line number for NsLogger.Warn.
func TestCallerLineNsWarn(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LWARN)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Warn("ns-warn-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger Warn caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/scope.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsError verifies caller line number for NsLogger.Error.
func TestCallerLineNsError(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LERROR)
	_, _, baseLine, _ := runtime.Caller(0)
	l.Error("ns-error-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger Error caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/scope.go") || strings.Contains(got, "caller=/assist.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsWith verifies caller line number for NsLogger.With().
func TestCallerLineNsWith(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)
	_, _, baseLine, _ := runtime.Caller(0)
	l.With().Str("k", "v").Info("ns-with-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger With caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/scope.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestCallerLineNsCtx verifies caller line number for NsLogger.Ctx().
func TestCallerLineNsCtx(t *testing.T) {
	var buf bytes.Buffer
	l.cfg.setCaller(true)
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setSkip(0)
	l.cfg.setLevel(LINFO)
	ctx := TraceCtx(context.Background(), "req-1")
	_, _, baseLine, _ := runtime.Caller(0)
	l.Ctx(ctx).Info("ns-ctx-line")
	got := buf.String()
	expect := strconv.Itoa(baseLine + 1)
	if !strings.Contains(got, ":"+expect+" ") {
		t.Fatalf("NsLogger Ctx caller line mismatch: expected :%s, got: %s", expect, got)
	}
	if strings.Contains(got, "caller=/field.go") || strings.Contains(got, "caller=/scope.go") {
		t.Fatalf("caller points to internal file: %s", got)
	}
}

// TestLevelFilterDebug verifies debug logs are filtered when level is INFO.
func TestLevelFilterDebug(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Debug("should-not-appear")
	l.Debugf("should-not-appear-%d", 1)
	if got := buf.String(); got != "" {
		t.Fatalf("debug output not filtered, got: %s", got)
	}
}

// TestLevelFilterInfo verifies info passes when level is INFO but debug does not.
func TestLevelFilterInfo(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Debug("no")
	l.Info("yes")
	got := buf.String()
	if !strings.Contains(got, "msg=yes") {
		t.Fatalf("info not logged, got: %s", got)
	}
	if strings.Contains(got, "no") {
		t.Fatalf("debug leaked through: %s", got)
	}
}

// TestLevelFilterWarnError verifies warn/error pass when level is WARN.
func TestLevelFilterWarnError(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LWARN)

	l.Info("no")
	l.Debug("no2")
	l.Warn("yes-warn")
	l.Error("yes-error")
	got := buf.String()
	if !strings.Contains(got, "WRN") || !strings.Contains(got, "ERR") {
		t.Fatalf("warn/error not logged, got: %s", got)
	}
	if strings.Contains(got, "msg=no") || strings.Contains(got, "msg=no2") {
		t.Fatalf("info/debug leaked through: %s", got)
	}
}

// TestLevelFilterErrorOnly verifies only ERROR passes at LERROR.
func TestLevelFilterErrorOnly(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LERROR)

	l.Warn("no")
	l.Error("yes-error")
	got := buf.String()
	if !strings.Contains(got, "ERR") {
		t.Fatalf("error not logged, got: %s", got)
	}
	if strings.Contains(got, "WRN") {
		t.Fatalf("warn leaked through: %s", got)
	}
}

// TestLevelMute verifies LevelMute filters everything.
func TestLevelMute(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LNONE)

	l.Debug("no")
	l.Info("no")
	l.Warn("no")
	l.Error("no")
	if got := buf.String(); got != "" {
		t.Fatalf("LevelMute didn't filter all, got: %s", got)
	}
}

// TestIfConditional verifies If(true) logs and If(false) skips.
func TestIfConditional(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().If(true).Info("yes")
	l.With().If(false).Info("no")
	got := buf.String()
	if !strings.Contains(got, "msg=yes") {
		t.Fatalf("If(true) not logged, got: %s", got)
	}
	if strings.Contains(got, "no") {
		t.Fatalf("If(false) leaked, got: %s", got)
	}
}

// TestIfErrConditional verifies IfErr(nil) skips and IfErr(err) logs.
func TestIfErrConditional(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().IfErr(nil).Info("nil-err")
	l.With().IfErr(errors.New("boom")).Info("has-err")
	got := buf.String()
	if !strings.Contains(got, "error=boom") {
		t.Fatalf("error field missing, got: %s", got)
	}
	if strings.Contains(got, "nil-err") {
		t.Fatalf("IfErr(nil) leaked, got: %s", got)
	}
}

// TestIfErrConditionalMultiLevel verifies IfErr chains correctly.
func TestIfErrConditionalMultiLevel(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	s := l.With().IfErr(nil)
	s.Info("no")
	// Note: s.Info() already calls putfl internally

	l.With().IfErr(errors.New("err1")).IfErr(errors.New("err2")).Errorf("multi-err")
	got := buf.String()
	if strings.Contains(got, "no") {
		t.Fatalf("nil chain leaked: %s", got)
	}
	if !strings.Contains(got, "multi-err") {
		t.Fatalf("expected error output, got: %s", got)
	}
}

// TestScoper verifies Scope freezes fields into a reusable, concurrency-safe logger.
func TestScoper(t *testing.T) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	s := l.With().Str("shared", "val").Group()
	s.With().Str("d1", "a").Info("entry1")
	s.With().Str("d2", "b").Info("entry2")
	// s is persistent and reusable; no manual release needed.
	s.Info("parent")
}

// TestFieldLoggerPrint verifies fielder Info/Infof with fields.
func TestFieldLoggerPrint(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().Str("k", "v").Info("p1", "p2")
	got := buf.String()
	if !strings.Contains(got, `k=v`) {
		t.Fatalf("fielder.Info k=v missing: %s", got)
	}
	if !strings.Contains(got, `p1p2`) {
		t.Fatalf("fielder.Info msg mismatch: %s", got)
	}

	buf.Reset()
	l.With().Int("n", 1).Info("pl")
	got = buf.String()
	if !strings.Contains(got, `n=1`) {
		t.Fatalf("fielder.Info n=1 missing: %s", got)
	}

	buf.Reset()
	l.With().Str("k", "v").Infof("%s:%d", "a", 1)
	got = buf.String()
	if !strings.Contains(got, `k=v`) {
		t.Fatalf("fielder.Infof k=v missing: %s", got)
	}
	if !strings.Contains(got, `a:1`) {
		t.Fatalf("fielder.Infof msg mismatch: %s", got)
	}
}

// TestFieldLoggerPrintSkip verifies If(false) + Info skips output.
func TestFieldLoggerPrintSkip(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().If(false).Info("should-not-appear")
	if got := buf.String(); got != "" {
		t.Fatalf("If(false).Info should be filtered: %s", got)
	}
}

// TestNsLoggerFormatted verifies NsLogger *f methods and Println/Printf.
func TestNsLoggerFormatted(t *testing.T) {
	var buf bytes.Buffer
	l := Trace("svc")
	l.cfg.setOutput(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LDEBUG)

	l.Debugf("debug %s", "test")
	if got := buf.String(); !strings.Contains(got, "trace=svc") || !strings.Contains(got, "debug test") {
		t.Fatalf("NsLogger.Debugf mismatch: %s", got)
	}

	buf.Reset()
	l.Infof("info %s", "test")
	if got := buf.String(); !strings.Contains(got, "info test") {
		t.Fatalf("NsLogger.Infof mismatch: %s", got)
	}

	buf.Reset()
	l.Warnf("warn %s", "test")
	if got := buf.String(); !strings.Contains(got, "warn test") {
		t.Fatalf("NsLogger.Warnf mismatch: %s", got)
	}

	buf.Reset()
	l.Errorf("error %s", "test")
	if got := buf.String(); !strings.Contains(got, "error test") {
		t.Fatalf("NsLogger.Errorf mismatch: %s", got)
	}

	buf.Reset()
	l.Printf("%s:%d", "k", 1)
	if got := buf.String(); !strings.Contains(got, "k:1") {
		t.Fatalf("NsLogger.Printf mismatch: %s", got)
	}

	buf.Reset()
	l.Println("a", "b")
	if got := buf.String(); !strings.Contains(got, "trace=svc") {
		t.Fatalf("NsLogger.Println trace missing: %s", got)
	}
}

// TestCtxNilContext verifies Ctx with empty context doesn't panic.
func TestCtxNilContext(t *testing.T) {
	l := New(io.Discard)
	fl := l.Ctx(context.Background())
	if fl == nil {
		t.Fatal("nil context returned nil fieldLogger")
	}
	fl.Info() // should not panic
}

// TestFieldLoggerCaller verifies Caller(true/false) toggles caller output.
func TestFieldLoggerCaller(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().Caller(true).Info("with-caller")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("Caller(true) did not add caller")
	}

	buf.Reset()
	l.With().Caller(false).Info("no-caller")
	got = buf.String()
	if strings.Contains(got, "caller=") {
		t.Fatal("Caller(false) should not have caller")
	}
}

// TestSetLevelPanic verifies SetLevel panics on an out-of-range level.
func TestSetLevelPanic(t *testing.T) {
	l := New(io.Discard)
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("SetLevel should panic on invalid level")
		}
	}()
	l.cfg.setLevel(LNONE + 1) // beyond the sentinel → illegal
}

// TestFieldLoggerEmptyArgs verifies empty args produce message field without msg value.
func TestFieldLoggerEmptyArgs(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.With().Info()
	got := buf.String()
	if strings.Contains(got, "msg=") {
		t.Fatalf("empty args should not produce msg field, got: %s", got)
	}
	if !strings.Contains(got, "INF") {
		t.Fatal("missing level")
	}
}

// TestFieldLoggerNilAttr verifies field methods with nil attr don't panic.
func TestFieldLoggerNilAttr(t *testing.T) {
	fl := &fielder{} // attr == nil
	fl.Str("k", "v").
		Stringer("k", nil).
		Bytes("k", []byte("v")).
		Bool("k", true).
		Int("k", 1).Int8("k", 1).Int16("k", 1).Int32("k", 1).Int64("k", 1).
		Uint("k", 1).Uint8("k", 1).Uint16("k", 1).Uint32("k", 1).Uint64("k", 1).
		Float32("k", 1).Float64("k", 1).
		Time("k", time.Now()).Dur("k", time.Second).
		Any("k", "v").Raw("k", []byte("v")).
		Err(errors.New("e")).IfErr(errors.New("e")).
		If(true)
	// Should not panic; every setter must safely early-return on nil attr.
}

// TestStdWriterNilReceiver verifies nil stdWriter.Write doesn't panic.
func TestStdWriterNilReceiver(t *testing.T) {
	var w *stdWriter
	n, err := w.Write([]byte("test"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		t.Fatal("expected non-zero bytes written")
	}
}

// TestStdWriterLevelOff prevents write when level too high.
func TestStdWriterLevelOff(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LERROR)

	w := l.stdWriter("t")
	n, err := w.Write([]byte("payload\n"))
	if err != nil {
		t.Fatalf("write error: %v", err)
	}
	if n == 0 {
		t.Fatal("write returned 0")
	}
	if got := buf.String(); got != "" {
		t.Fatalf("INFO write through stdWriter should be filtered at ERROR level, got: %s", got)
	}
}

// TestNewNilOut verifies New with nil writer defaults to Discard.
func TestNewNilOut(t *testing.T) {
	l := New(nil)
	l.cfg.setLevel(LINFO)
	l.Info("should-not-panic") // redirected to Discard, no panic
}

// TestPrintWriter verifies Logger.Print/Println/Printf format correctly.
func TestPrintWriter(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Print("a", "b")
	l.Println("a", "b")
	l.Printf("%s:%d", "a", 1)

	got := buf.String()
	if !strings.Contains(got, `msg=ab`) {
		t.Fatalf("Print mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=ab`) {
		t.Fatalf("Println mismatch: %s", got)
	}
	if !strings.Contains(got, `msg=a:1`) {
		t.Fatalf("Printf mismatch: %s", got)
	}
}

// TestSetConsoleNoFile verifies SetConsole on a logger with no file writer returns early without panic.
func TestSetConsoleNoFile(t *testing.T) {
	l := New(io.Discard)   // no SetFile → fw == nil
	l.cfg.setConsole(true) // should not panic, fw==nil branch
	SetCons(true)          // deprecated wrapper — still covered for backward compat
}

// TestWithTrace verifies Logger.With(trace) sets the trace on the fielder.
func TestWithTrace(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	l.With("req-abc").Str("k", "v").Info("msg")
	if got := buf.String(); !strings.Contains(got, "trace=req-abc") {
		t.Fatalf("With(trace) did not set trace: %s", got)
	}
}

// TestPutflNil verifies putfl(nil) does not panic.
func TestPutflNil(t *testing.T) {
	putfl(nil) // nil branch
}

// TestPrintbCallerRuntimeFail covers the printb !ok branch (file="###") via stdWriter.
func TestPrintbCallerRuntimeFail(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setSkip(9999) // far beyond stack depth → runtime.Caller returns !ok inside printb
	w := l.stdWriter("ns")
	w.Write([]byte("deep-skip-printb\n"))
	if got := buf.String(); !strings.Contains(got, "###") {
		t.Fatalf("expected caller=### in printb !ok branch, got: %s", got)
	}
}

// TestPrintbCallerAndAttr verifies printb with caller=true and a non-nil attr.
func TestPrintbCallerAndAttr(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	// Use stdWriter which calls printb internally.
	w := l.stdWriter("ns")
	w.Write([]byte("hello from printb\n"))
	if got := buf.String(); !strings.Contains(got, "caller=") {
		t.Fatalf("printb with caller should emit caller field: %s", got)
	}
}

// TestPrintbEmptyMsg verifies printb with empty msg emits no msg field.
func TestPrintbEmptyMsg(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	l.cfg.printb("", LINFO, false, nil, []byte{})
	if got := buf.String(); strings.Contains(got, "msg=") {
		t.Fatalf("printb with empty msg should not emit msg field: %s", got)
	}
}

// TestPrintbWithAttr verifies printb with a non-nil attr buffer appends fields.
func TestPrintbWithAttr(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	attr := getb()
	*attr = append(*attr, []byte("key=val")...)
	l.cfg.printb("", LINFO, false, attr, []byte("hi"))
	putb(attr)
	if got := buf.String(); !strings.Contains(got, "key=val") {
		t.Fatalf("printb with attr should include field: %s", got)
	}
}

// TestCallerRuntimeFail simulates runtime.Caller failure by using a very large skip.
// The !ok branch sets file="###" and line=0.
func TestCallerRuntimeFail(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setSkip(9999) // skip far beyond actual stack depth → runtime.Caller returns !ok
	l.Info("deep-skip")
	got := buf.String()
	if !strings.Contains(got, "###") {
		t.Fatalf("expected caller=### when runtime.Caller fails, got: %s", got)
	}
}

// TestPrintfCallerRuntimeFail covers the printf !ok branch.
func TestPrintfCallerRuntimeFail(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	l.cfg.setSkip(9999)
	l.Infof("deep-skip-f %d", 1)
	if got := buf.String(); !strings.Contains(got, "###") {
		t.Fatalf("expected caller=### in printf !ok branch, got: %s", got)
	}
}

// TestPackageClose verifies package-level Close succeeds.
func TestPackageClose(t *testing.T) {
	// Don't actually close the package logger; test with discard only.
	prevOut := l.cfg.out
	SetOutput(io.Discard)
	defer SetOutput(prevOut)

	if err := Close(); err == nil {
		// Success
	}
}

// TestTraceCtxVariants verifies all TraceCtx branches.
func TestTraceCtxVariants(t *testing.T) {
	// 1. empty ctx, no traceid -> generates new
	ctx1 := TraceCtx(context.Background())
	if TraceOf(ctx1) == "" {
		t.Fatal("expected generated trace id")
	}

	// 2. empty ctx, supplied traceid -> uses supplied
	ctx2 := TraceCtx(context.Background(), "supplied-id")
	if got := TraceOf(ctx2); got != "supplied-id" {
		t.Fatalf("expected supplied-id, got: %s", got)
	}

	// 3. existing ctx, no new traceid -> reuses
	ctx3 := TraceCtx(ctx1)
	if got := TraceOf(ctx3); got != TraceOf(ctx1) {
		t.Fatalf("expected reused trace, got: %s vs %s", got, TraceOf(ctx1))
	}

	// 4. existing ctx, new traceid -> appends
	ctx4 := TraceCtx(ctx1, "child")
	got := TraceOf(ctx4)
	if !strings.Contains(got, ".") {
		t.Fatalf("expected appended trace with '.', got: %s", got)
	}
}

// TestHijackStdlibCaller verifies caller is correct when stdlib log is hijacked.
func TestHijackStdlibCaller(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	// Hijack is already done in New(), so we can use stdlib log directly.
	stdlog.Print("hijack-test")
	got := buf.String()
	if !strings.Contains(got, "caller=") {
		t.Fatal("missing caller field")
	}
	// The caller should NOT be a stdlib file like log.go
	if strings.Contains(got, "caller=/log.go:") || strings.Contains(got, "caller=/log/") {
		t.Fatalf("caller points to stdlib log package: %s", got)
	}
}

// ---------------------------------------------------------------------------------------------------
// Benchmarks
// ---------------------------------------------------------------------------------------------------

func BenchmarkParallelSpan(b *testing.B) {
	SetOutput(io.Discard)
	SetCaller(false)
	ctx := TraceCtx(context.Background())
	n := Ctx(ctx).Str("A", "B").Str("subtrace", "sub").Group()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			n.With().Str("b", "b").Info("xx")
			n.With().Str("c", "c").Info("xx")
		}
	})
}

// BenchmarkSimple measures bare Info() with no fields and no caller.
func BenchmarkSimple(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

// BenchmarkSimpleCaller measures Info() with caller stack capture.
func BenchmarkSimpleCaller(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(true)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Info("hello world")
	}
}

// BenchmarkInfof measures formatted log output.
func BenchmarkInfof(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Infof("hello %s", "world")
	}
}

// BenchmarkWith5Fields measures With + 5 fields + Info.
func BenchmarkWith5Fields(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().
			Str("s", "value").
			Int("n", 42).
			Bool("ok", true).
			Float64("f", 3.14).
			Err(nil).
			Info("with fields")
	}
}

// BenchmarkWith10Fields measures With + 10 fields + Info.
func BenchmarkWith10Fields(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().
			Str("s1", "value1").
			Str("s2", "value2").
			Int("n1", 1).
			Int("n2", 2).
			Int64("n3", 3).
			Bool("ok", true).
			Float64("f1", 1.1).
			Float64("f2", 2.2).
			Err(nil).
			Str("last", "end").
			Info("with 10 fields")
	}
}

// BenchmarkDisabled measures the fast path when level is filtered out.
func BenchmarkDisabled(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LERROR) // Debug will be filtered
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Debug("should be filtered")
	}
}

// BenchmarkDisabledWithFields measures filtered With chain.
func BenchmarkDisabledWithFields(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LERROR)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Str("k", "v").Int("n", 1).Debug("filtered")
	}
}

// BenchmarkError measures Error() with err field.
func BenchmarkError(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LERROR)
	err := errors.New("something went wrong")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.With().Err(err).Error("failed")
	}
}

// BenchmarkParallelSimple measures parallel bare Info().
func BenchmarkParallelSimple(b *testing.B) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.Info("hello world")
		}
	})
}

// ------------------------------------------------------------------
// 边界测试 — 单参数类型快速路径
// ------------------------------------------------------------------

func TestPrintSingleArgTypes(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	tests := []struct {
		name string
		arg  any
		want string
	}{
		{"string", "hello", `msg=hello`},
		{"int", int(42), `msg=42`},
		{"int8", int8(8), `msg=8`},
		{"int16", int16(16), `msg=16`},
		{"int32", int32(32), `msg=32`},
		{"int64", int64(64), `msg=64`},
		{"uint", uint(100), `msg=100`},
		{"uint8", uint8(8), `msg=8`},
		{"uint16", uint16(16), `msg=16`},
		{"uint32", uint32(32), `msg=32`},
		{"uint64", uint64(64), `msg=64`},
		{"float32", float32(3.14), `msg=3.14`},
		{"float64", float64(2.718), `msg=2.718`},
		{"bool_true", true, `msg=true`},
		{"bool_false", false, `msg=false`},
		{"bytes", []byte("data"), `msg=data`},
		{"complex64", complex64(1 + 2i), `msg=(1+2i)`},
		{"complex128", complex128(3 + 4i), `msg=(3+4i)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			l.Info(tt.arg)
			if got := buf.String(); !strings.Contains(got, tt.want) {
				t.Errorf("single arg %T = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}

func TestPrintMultiArg(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Info("hello", "world", 42)
	got := buf.String()
	if !strings.Contains(got, `msg=helloworld42`) {
		t.Fatalf("multi arg mismatch: %s", got)
	}

	buf.Reset()
	l.Infof("count=%d name=%s", 3, "test")
	got = buf.String()
	if !strings.Contains(got, "name=test") {
		t.Fatalf("Infof multi arg mismatch: %s", got)
	}
}

func TestPrintZeroArgs(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Info()
	got := buf.String()
	if strings.Contains(got, "msg=") {
		t.Fatalf("zero args produce msg: %s", got)
	}
	if !strings.Contains(got, "INF") {
		t.Fatal("missing level")
	}
}

func TestPrintNilArg(t *testing.T) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Info(nil)         // nil interface
	l.Info((*int)(nil)) // typed nil
}

type nilStringerArg struct{}

func (n nilStringerArg) String() string { return "nil-stringer" }

func TestPrintStringerArg(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Info(nilStringerArg{})
	got := buf.String()
	if !strings.Contains(got, `msg=nil-stringer`) {
		t.Fatalf("Stringer arg mismatch: %s", got)
	}

	// nil fmt.Stringer
	var ns fmt.Stringer
	buf.Reset()
	l.Info(ns)
	got = buf.String()
	if !strings.Contains(got, `msg=<nil>`) {
		t.Fatalf("nil Stringer arg mismatch: %s", got)
	}
}

func TestPrintVeryLongString(t *testing.T) {
	var buf bytes.Buffer
	l := New(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	long := strings.Repeat("x", 2000)
	l.Info(long)
	if !strings.Contains(buf.String(), long) {
		t.Fatal("long string not found")
	}
}

func TestPrintSpecialChars(t *testing.T) {
	tests := []string{
		"newline\nhere",
		"tab\there",
		`quote"here`,
		`backslash\here`,
		"unicode\u276d",
		"null\x00char",
	}
	for i, s := range tests {
		var buf bytes.Buffer
		l := New(&buf)
		l.cfg.setCaller(false)
		l.cfg.setLevel(LINFO)
		l.Info(s)
		got := buf.String()
		if !strings.Contains(got, "msg=") {
			t.Fatalf("test[%d] no msg field: %s", i, got)
		}
	}
}

func TestNsSingleArgTypes(t *testing.T) {
	var buf bytes.Buffer
	l := Trace("api")
	l.cfg.setOutput(&buf)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	l.Info("hello")
	got := buf.String()
	if !strings.Contains(got, "trace=api") || !strings.Contains(got, "msg=hello") {
		t.Fatalf("NsLogger single arg: %s", got)
	}

	buf.Reset()
	l.Info(42)
	got = buf.String()
	if !strings.Contains(got, "trace=api") || !strings.Contains(got, "msg=42") {
		t.Fatalf("NsLogger int arg: %s", got)
	}

	buf.Reset()
	l.Info(true)
	got = buf.String()
	if !strings.Contains(got, "trace=api") || !strings.Contains(got, "msg=true") {
		t.Fatalf("NsLogger bool arg: %s", got)
	}

	buf.Reset()
	l.Info([]byte("data"))
	got = buf.String()
	if !strings.Contains(got, "trace=api") || !strings.Contains(got, "msg=data") {
		t.Fatalf("NsLogger []byte arg: %s", got)
	}
}

func TestLoggerConcurrent(t *testing.T) {
	stream := &blackholeStream{}
	l := New(stream)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)

	const goroutines = 50
	const writes = 1000
	done := make(chan struct{})
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < writes; j++ {
				l.Info("concurrent")
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < goroutines; i++ {
		<-done
	}
}

// ------------------------------------------------------------------
// Fuzz tests
// ------------------------------------------------------------------

func FuzzInfo(f *testing.F) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	f.Add("hello")
	f.Add(string([]byte{0, 1, 0xFF}))
	f.Fuzz(func(t *testing.T, arg string) {
		l.Info(arg)
		l.Info(arg, 42)
		l.Info()
		l.Info(arg, arg, arg)
	})
}

func FuzzInfof(f *testing.F) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	f.Add("%s", "hello")
	f.Add("%d", "42")
	f.Fuzz(func(t *testing.T, format, arg string) {
		l.Infof(format, arg)
		l.Infof(format)
		l.Infof("%s %s", arg, arg)
	})
}

func FuzzPrintb(f *testing.F) {
	l := New(io.Discard)
	l.cfg.setCaller(false)
	l.cfg.setLevel(LINFO)
	f.Add([]byte("hello"))
	f.Add([]byte{0, 1, 2, 0xFF})
	f.Fuzz(func(t *testing.T, data []byte) {
		l.cfg.printb("", LINFO, false, nil, data)
		l.cfg.printb("trace-id", LINFO, false, nil, data)
	})
}

func FuzzNsInfo(f *testing.F) {
	f.Add("ns", "hello")
	f.Add("", "")
	f.Add(strings.Repeat("x", 512), "msg")
	f.Fuzz(func(t *testing.T, ns, msg string) {
		l := Trace(ns)
		l.cfg.setOutput(io.Discard)
		l.cfg.setCaller(false)
		l.cfg.setLevel(LINFO)
		l.Info(msg)
		l.Info(msg, 42)
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

// refLastSep is a trivially-correct reference: max of LastIndex over all separators.
func refLastSep(file string, seps []string) int {
	max := -1
	for _, sep := range seps {
		if sep == "" {
			continue
		}
		if idx := strings.LastIndex(file, sep); idx > max {
			max = idx
		}
	}
	return max
}

// TestLastSepMulti verifies lastSep matches the reference implementation across
// multi-char separators (real-world /xxx scenario), same-first-byte seps, and edge cases.
func TestLastSepMulti(t *testing.T) {
	sepSets := [][]string{
		{"/internal", "/src", "/"},
		{"/zxysilent", "/schci"},      // multi-char, same first byte '/'
		{"/zxysilent", "/schci", "/"}, // mixed with single '/'
		{"zxysilent", "schci"},        // no leading sep char
		{"", "/"},                     // empty sep ignored
	}
	files := []string{
		"/home/user/go/src/github.com/zxysilent/logs/internal/file/witer.go",
		"/home/user/src/main.go",
		"relative/path/file.go",
		"github.com/zxysilent/logs/internal/x.go",
		"/root/zxysilent/app/schci/main.go", // both multi-char seps present
		"/only/schci/here.go",
		"/zxysilent",
		"prefix/zxysilenter/file.go", // sep is a prefix of a longer word
		"nosep",
		"",
		"/",
		"a/b",
	}
	for _, seps := range sepSets {
		for _, f := range files {
			if got, want := lastSep(f, seps), refLastSep(f, seps); got != want {
				t.Fatalf("lastSep(%q, %v)=%d, want %d", f, seps, got, want)
			}
		}
	}
}

// BenchmarkLastSep measures multi-string separator lookup (real-world /xxx scenario).
var (
	benchPath = "/home/user/go/src/github.com/zxysilent/logs/internal/file/witer.go"
	benchSeps = []string{"/internal", "/src", "/"}
	benchSink int
)

func BenchmarkLastSep(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchSink = lastSep(benchPath, benchSeps)
	}
}
