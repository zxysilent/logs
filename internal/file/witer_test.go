package file

import (
	"os"
	"testing"
)

func TestName2time(t *testing.T) {
	f := New("./log/app.log")
	f.delete()
}

func TestReadDir(t *testing.T) {
	dirs, err := os.ReadDir("./log")
	if err != nil {
		return
	}
	for _, dir := range dirs {
		t.Log(dir.Name())
	}
}
