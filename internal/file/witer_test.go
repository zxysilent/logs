package file

import (
	"os"
	"testing"
)

func TestName2time(t *testing.T) {
	f := New("../../logs/app.log")
	t.Logf("%+v", f)
	f.delete()
}

func TestReadDir(t *testing.T) {
	dirs, err := os.ReadDir("../../logs")
	if err != nil {
		return
	}
	for _, dir := range dirs {
		t.Log(dir.Name())
	}
}
