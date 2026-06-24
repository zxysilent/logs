// Package logs_test demonstrates the default global logger API.
package logs_test

import (
	"github.com/zxysilent/logs"
)

// Example: basic level methods on the default logger.
func Example_defaultLevels() {
	logs.Debug("debug message")
	logs.Debugf("debugf: %s", "detail")
	logs.Info("info message")
	logs.Infof("infof: %s", "detail")
	logs.Warn("warn message")
	logs.Warnf("warnf: %s", "detail")
	logs.Error("error message")
	logs.Errorf("errorf: %s", "detail")
	// Output:
}

// Example: SetCaller enables file:line in log output.
func Example_defaultSetCaller() {
	logs.SetCaller(true)
	logs.SetSep("/")
	logs.Info("caller enabled")
	// Output:
}

// Example: SetLevel filters log output.
func Example_defaultSetLevel() {
	logs.SetLevel(logs.LevelDebug)
	logs.Debug("visible when level is DEBUG")
	logs.SetLevel(logs.LevelWarn)
	logs.Debug("not visible, level is WARN")
	logs.Warn("visible, level is WARN")
	// Output:
}

// Example: Print/Println/Printf compatibility with standard library signatures.
func Example_defaultPrint() {
	logs.Print("hello", "world")    // msg=helloworld
	logs.Println("hello", "world")  // msg=helloworld
	logs.Printf("%s:%d", "key", 42) // msg=key:42
	// Output:
}

// Example: With starts a field chain; must call a terminal method.
func Example_defaultWith() {
	logs.With().
		Str("user", "alice").
		Int("age", 30).
		Info("fields attached")
	// Output:
}
