package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
func TestWriterLifecycle(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path)
	defer w.Close()

	w.SetCons(false)
	if w.cons {
		t.Fatalf("SetCons(false) not applied")
	}

	maxSize := w.maxsize
	w.SetMaxSize(0)
	if w.maxsize != maxSize {
		t.Fatalf("SetMaxSize(0) should be ignored")
	}

	w.SetMaxSize(2)
	w.SetMaxAge(7)
	if w.maxage != 7 {
		t.Fatalf("SetMaxAge not applied")
	}

	if !w.equaldate([]byte("2026-05-09"), []byte("time=2026-05-09")) {
		t.Fatalf("equaldate should match same date")
	}
	if w.equaldate([]byte("2026-05-09"), []byte("time=2026-05-10")) {
		t.Fatalf("equaldate should not match different dates")
	}

	if got := w.time2name(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)); got != ".2024-01-02-030405" {
		t.Fatalf("time2name mismatch: %s", got)
	}

	if _, err := w.Write([]byte("a")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	if err := w.flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if !strings.Contains(string(content), "a") {
		t.Fatalf("written content missing: %s", content)
	}

	fixed := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	w.created = fixed
	if err := w.rotate(); err != nil {
		t.Fatalf("rotate failed: %v", err)
	}
	backup := filepath.Join(dir, w.fname+w.time2name(fixed)+w.fsuffix)
	if _, err := os.Stat(backup); err != nil {
		t.Fatalf("backup log missing after rotate: %v", err)
	}

	stale := filepath.Join(dir, "app.2000-01-01-000000.log")
	if err := os.WriteFile(stale, []byte("old"), 0o644); err != nil {
		t.Fatalf("create stale file failed: %v", err)
	}
	w.maxage = 1
	w.delete()
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file should be deleted, err=%v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if err := w.close(); err != nil {
		t.Fatalf("close on already closed writer failed: %v", err)
	}
}
