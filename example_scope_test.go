// Package logs_test demonstrates the ScopeLogger (reusable namespace / field scope) API.
package logs_test

import (
	"context"

	"github.com/zxysilent/logs"
)

// Example: Ns creates a namespaced scope that always emits trace=ns.
func Example_scopeNs() {
	l := logs.Ns("myapp")
	l.Info("namespaced log")
	l.Debug("debug with namespace")
	l.Warn("warn with namespace")
	l.Error("error with namespace")
	// Output:
}

// Example: Ns print family.
func Example_scopeNsPrint() {
	l := logs.Ns("myapp")
	l.Print("hello")
	l.Println("hello")
	l.Printf("%s:%d", "key", 1)
	// Output:
}

// Example: Scope freezes a field chain into a persistent, reusable scope.
func Example_scopeFields() {
	app := logs.With().Str("svc", "api").Int("pid", 1).Scope()
	app.Info("started")             // svc=api pid=1
	app.With().Int("uid", 9).Info("login") // svc=api pid=1 uid=9
	// Output:
}

// Example: Logger.Scope creates an empty reusable scope; With adds fields per entry.
func Example_scopeEmpty() {
	s := logs.Scope()
	s.With().Str("user", "alice").Info("field log")
	// Output:
}

// Example: Ns With carries the namespace and joins an optional sub trace.
func Example_scopeWith() {
	l := logs.Ns("myapp")
	l.With().Str("user", "alice").Info("trace=myapp")
	l.With("req-1").Info("trace=myapp.req-1")
	// Output:
}

// Example: Ns Ctx combines the namespace root with a context trace.
func Example_scopeCtx() {
	l := logs.Ns("myapp")
	ctx := logs.TraceCtx(context.Background(), "req-1")
	l.Ctx(ctx).Info("trace=myapp.req-1")
	// Output:
}

// Example: Ns Ctx without context trace falls back to namespace only.
func Example_scopeCtxEmpty() {
	l := logs.Ns("myapp")
	l.Ctx(context.Background()).Info("trace=myapp only")
	// Output:
}
