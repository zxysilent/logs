// Package logs_test demonstrates fieldLogger (With/Ctx field chain) API.
package logs_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/zxysilent/logs"
)

type myStringer struct {
	prefix string
}

func (m myStringer) String() string {
	return m.prefix + "-stringer"
}

// Example: fieldLogger field types overview.
func Example_fieldTypes() {
	l := logs.New(nil)
	l.With().
		Str("str", "hello").
		Bytes("bytes", []byte("world")).
		Bool("active", true).
		Int("count", 42).
		Int8("i8", 8).
		Int16("i16", 16).
		Int32("i32", 32).
		Int64("i64", 64).
		Uint("u", 1).
		Uint8("u8", 2).
		Uint16("u16", 3).
		Uint32("u32", 4).
		Uint64("u64", 5).
		Float32("f32", 3.14).
		Float64("f64", 2.718).
		Time("ts", time.Time{}).
		Dur("latency", time.Second).
		Stringer("st", myStringer{"my"}).
		Info("all field types")
	// Output:
}

// Example: Err / IfErr / If control flow.
func Example_fieldError() {
	l := logs.New(nil)
	err := errors.New("something failed")

	// Err always logs the error field
	l.With().Err(err).Error("request failed")
	l.With().Err(nil).Info("no error")

	// IfErr only logs when err != nil
	l.With().IfErr(err).Error("logged because err is non-nil")
	l.With().IfErr(nil).Info("skipped because err is nil") // not logged

	// If controls logging based on a boolean
	l.With().If(true).Info("logged, condition is true")
	l.With().If(false).Warn("skipped, condition is false") // not logged
	// Output:
}

// Example: Scope freezes a field chain into a reusable ScopeLogger.
func Example_fieldScope() {
	l := logs.New(nil)
	base := l.With().Str("app", "myapp").Str("env", "prod").Scope()

	// base is persistent; derive a one-shot fielder via With for each entry
	base.With().Int("step", 1).Info("first step")
	base.With().Int("step", 2).Info("second step")
	base.With().Str("status", "done").Info("final step")
	// Output:
}

// Example: Caller enables/disables caller info on a per-chain basis.
func Example_fieldCaller() {
	l := logs.New(nil)
	l.With().Caller(true).Info("includes caller line")
	l.With().Caller(false).Info("no caller line")
	// Output:
}

// Example: Ctx with TraceCtx attaches a trace id to the log.
func Example_fieldCtx() {
	ctx := logs.TraceCtx(context.Background())
	l := logs.New(nil)
	l.Ctx(ctx).Str("op", "query").Info("traced operation")
	// Output:
}

// Example: Any accepts arbitrary types via fmt.Stringer or fmt.Sprint.
func Example_fieldAny() {
	l := logs.New(nil)
	// Any uses fmt.Sprint-like fallback for unrecognized types
	l.With().Info(fmt.Stringer(myStringer{"x"}))
	l.With().Info(struct{ A int }{A: 1})
	// Output:
}

// Example: PutNil on nil value.
func Example_fieldNil() {
	l := logs.New(nil)
	l.With().Err(nil).Info("nil error logged as null")
	l.With().Stringer("key", nil).Info("nil Stringer logged as null")
	// Output:
}
