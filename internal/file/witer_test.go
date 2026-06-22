package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestName2time(t *testing.T) {
	f := New("../../logs/app.log", false)
	t.Logf("%+v", f)
	f.delete(f.maxage)
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
	w := New(path, false)
	defer w.Close()

	w.SetConsole(false)
	if w.console {
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
	w.delete(w.maxage)
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale file should be deleted, err=%v", err)
	}

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
}

// TestWriteSizeRotate verifies a write exceeding maxsize triggers rotation.
func TestWriteSizeRotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, false)
	defer w.Close()

	// 1 MiB cap; write enough to exceed it and force a size-based rotate.
	w.SetMaxSize(1)
	chunk := make([]byte, 256*1024)
	for i := 0; i < 6; i++ {
		if _, err := w.Write(chunk); err != nil {
			t.Fatalf("write %d failed: %v", i, err)
		}
	}
	w.flush()

	// At least one rotated backup file should exist.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	rotated := 0
	for _, e := range entries {
		if e.Name() != "app.log" && strings.HasSuffix(e.Name(), ".log") {
			rotated++
		}
	}
	if rotated == 0 {
		t.Fatalf("expected a rotated backup after exceeding maxsize, found none")
	}
}

// TestWriteCrossDayRotate verifies a write on a new day triggers rotation.
func TestWriteCrossDayRotate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, false)
	defer w.Close()

	// First write establishes the current file and its creation date.
	if _, err := w.Write([]byte("time=2026-05-09T00:00:00 first\n")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	// Force creates to a known past date so the next write looks like a new day.
	w.creates = []byte("2000-01-01")
	if _, err := w.Write([]byte("time=2026-05-10T00:00:00 second\n")); err != nil {
		t.Fatalf("cross-day write failed: %v", err)
	}
	w.flush()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	rotated := 0
	for _, e := range entries {
		if e.Name() != "app.log" && strings.HasSuffix(e.Name(), ".log") {
			rotated++
		}
	}
	if rotated == 0 {
		t.Fatalf("expected a rotated backup after crossing day, found none")
	}
}

// TestWriteAfterClose verifies writing after Close returns os.ErrClosed.
func TestWriteAfterClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, false)

	if _, err := w.Write([]byte("before close\n")); err != nil {
		t.Fatalf("write before close failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if _, err := w.Write([]byte("after close\n")); err != os.ErrClosed {
		t.Fatalf("expected os.ErrClosed after close, got: %v", err)
	}
}

// TestCloseIdempotent verifies Close is safe to call multiple times.
func TestCloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, false)

	if _, err := w.Write([]byte("data\n")); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("second close should be idempotent, got: %v", err)
	}
}

// TestRotateFallbackOnRenameFailure verifies rotate continues even if rename fails.
func TestRotateFallbackOnRenameFailure(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, false)
	defer w.Close()

	// Write something to create the file.
	if _, err := w.Write([]byte("first\n")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}
	// Remove the directory to force rename to fail.
	os.RemoveAll(dir)
	// rotate should still create a new file (fallback).
	if err := w.rotate(); err != nil {
		t.Fatalf("rotate should not fail even when rename fails: %v", err)
	}
	// The new file should exist because rotate recreates it.
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("new file should exist after rotate fallback: %v", err)
	}
}

// TestWriteCons verifies cons=true also writes to stderr without error.
func TestWriteCons(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path, true)
	defer w.Close()

	if !w.console {
		t.Fatalf("cons should be true")
	}
	if _, err := w.Write([]byte("to file and stderr\n")); err != nil {
		t.Fatalf("write with cons failed: %v", err)
	}
}
