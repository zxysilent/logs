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

func TestNew(t *testing.T) {
	applog := NewLogger("logs/xxx.log")
	defer applog.Flush()
	// 设置日志输出等级
	// 开发环境下设置输出等级为DEBUG，线上环境设置为INFO
	applog.SetLevel(DEBUG)
	// 设置输出调用信息
	applog.SetCallInfo(true)
	// 设置同时显示到控制台
	// 默认只输出到文件
	// applog.SetConsole(true)
	applog.Debug("Debug Logger")
	applog.Debugf("Debugf %s", "Logger")

	applog.Info("Info Logger")
	applog.Infof("Infof %s", "Logger")

	applog.Warn("Warn Logger")
	applog.Warnf("Warnf %s", "Logger")

	applog.Error("Error Logger")
	applog.Errorf("Errorf %s", "Logger")
}
