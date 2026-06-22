// Package logs_test demonstrates the Group (reusable field scope) and Trace (namespace) API.
package logs_test

import (
	"context"

	"github.com/zxysilent/logs"
)

// Example: Trace creates a namespaced logger that always emits trace=...
func Example_traceNs() {
	l := logs.Trace("myapp")
	l.Info("namespaced log")
	l.Debug("debug with namespace")
	l.Warn("warn with namespace")
	l.Error("error with namespace")
	// Output:
}

// Example: Trace with Print/Println/Printf.
func Example_traceNsPrint() {
	l := logs.Trace("myapp")
	l.Print("hello")
	l.Println("hello")
	l.Printf("%s:%d", "key", 1)
	// Output:
}

// Example: Group freezes a field chain into a persistent, reusable logger.
func Example_groupFields() {
	app := logs.With().Str("svc", "api").Int("pid", 1).Group()
	app.Info("started")                    // svc=api pid=1
	app.With().Int("uid", 9).Info("login") // svc=api pid=1 uid=9
	// Output:
}

// Example: Trace With carries the namespace and joins an optional sub trace.
func Example_traceWith() {
	l := logs.Trace("myapp")
	l.With().Str("user", "alice").Info("trace=myapp")
	l.With("req-1").Info("trace=myapp.req-1")
	// Output:
}

// Example: Trace Ctx combines the namespace root with a context trace.
func Example_traceCtx() {
	l := logs.Trace("myapp")
	ctx := logs.TraceCtx(context.Background(), "req-1")
	l.Ctx(ctx).Info("trace=myapp.req-1")
	// Output:
}

// Example: Trace Ctx without context trace falls back to namespace only.
func Example_traceCtxEmpty() {
	l := logs.Trace("myapp")
	l.Ctx(context.Background()).Info("trace=myapp only")
	// Output:
}
