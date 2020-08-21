package logs

import (
	"testing"
)

func BenchmarkLogger(b *testing.B) {
	Info("info info info info")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Info("info info info info")
	}
	Flush()
}

func TestLogger(t *testing.T) {
	SetCallInfo(true)
	Debug("Debug Logger")
	Debugf("Debugf %s", "Logger")

	Info("Info Logger")
	Infof("Infof %s", "Logger")

	Warn("Warn Logger")
	Warnf("Warnf %s", "Logger")

	Error("Error Logger")
	Errorf("Errorf %s", "Logger")

	s := []int{1, 2, 3}
	for range s {
		Debug("Debug Logger")
	}
	Flush()
}
