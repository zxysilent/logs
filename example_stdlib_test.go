// Package logs_test demonstrates standard library integration API.
package logs_test

import (
	stdlog "log"
	"os"

	"github.com/zxysilent/logs"
)

// The global logger is initialized with New(os.Stderr), which calls hijackstd().
// After that, all stdlib log output is converted to logfmt and respects the
// logs level/caller settings.

// Example: stdlib log is automatically hijacked into logfmt.
func Example_stdlibHijack() {
	stdlog.Println("hello from stdlib log")
	// Output:
}

// Example: stdlib log respects the logs level filter.
func Example_stdlibHijackLevel() {
	logs.SetLevel(logs.LWARN)
	stdlog.Println("this INFO message is suppressed")
	logs.SetLevel(logs.LINFO) // restore
	// Output:
}

// Example: stdlib log respects the caller setting.
func Example_stdlibHijackCaller() {
	logs.SetCaller(true)
	logs.SetSep("/")
	stdlog.Println("caller visible")
	logs.SetCaller(false) // restore
	// Output:
}

// Example: stdlib prefix set before New becomes the log namespace.
func Example_stdlibHijackPrefix() {
	// Set stdlib prefix before creating the logger
	stdlog.SetPrefix("myprefix")
	l := logs.New(os.Stderr, logs.WithCaller(false)) // hijackstd reads prefix as namespace
	_ = l
	defer stdlog.SetPrefix("") // restore

	stdlog.Println("message with ns")
	// Output:
}

// Example: Logger.Print / Println / Printf mirror stdlib signatures.
func Example_stdlibPrintCompat() {
	l := logs.New(os.Stderr)
	l.Print("a", "b")         // msg=ab
	l.Println("a", "b")       // msg=a b
	l.Printf("%s:%d", "k", 1) // msg=k:1
	// Output:
}
