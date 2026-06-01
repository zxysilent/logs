// Package logs_test demonstrates NsLogger (namespaced logger) API.
package logs_test

import (
	"context"

	"github.com/zxysilent/logs"
)

// Example: Ns creates a namespaced logger that always emits trace=ns.
func Example_nsBasic() {
	l := logs.Ns("myapp")
	l.Info("namespaced log")
	l.Debug("debug with namespace")
	l.Warn("warn with namespace")
	l.Error("error with namespace")
	// Output:
}

// Example: Ns print family.
func Example_nsPrint() {
	l := logs.Ns("myapp")
	l.Print("hello")
	l.Println("hello")
	l.Printf("%s:%d", "key", 1)
	// Output:
}

// Example: Ns With returns a fieldLogger that also carries the namespace.
func Example_nsWith() {
	l := logs.Ns("myapp")
	l.With().Str("user", "alice").Info("namespaced field log")
	// Output:
}

// Example: Ns Ctx combines the namespace root with a context trace.
func Example_nsCtx() {
	l := logs.Ns("myapp")
	ctx := logs.TraceCtx(context.Background(), "req-1")
	l.Ctx(ctx).Info("trace=myapp.req-1")
	// Output:
}

// Example: Ns Ctx without context trace falls back to namespace only.
func Example_nsCtxEmpty() {
	l := logs.Ns("myapp")
	l.Ctx(context.Background()).Info("trace=myapp only")
	// Output:
}
