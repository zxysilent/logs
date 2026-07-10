package file

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkWriteNoConsole measures raw Write throughput (console disabled).
func BenchmarkWriteNoConsole(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path)
	w.SetConsole(false)
	defer w.Close()

	msg := []byte("time=2026-07-10T12:00:00Z level=INF msg=benchmark payload\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(msg)
	}
}

// BenchmarkWriteWithConsole measures Write throughput with console mirroring.
func BenchmarkWriteWithConsole(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path)
	defer w.Close()

	// Discard stderr during benchmark.
	old := os.Stderr
	os.Stderr, _ = os.Open(os.DevNull)
	defer func() { os.Stderr = old }()

	msg := []byte("time=2026-07-10T12:00:00Z level=INF msg=benchmark payload\n")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Write(msg)
	}
}

// BenchmarkRotate measures rotate cost (flush + sync + rename + open).
func BenchmarkRotate(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path)
	w.SetConsole(false)
	defer w.Close()

	// Pre-populate with one write so rotate has a file to close.
	w.Write([]byte("time=2026-07-10T12:00:00Z level=INF msg=seed\n"))
	w.created = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.rotate()
	}
}

// BenchmarkName2time measures date extraction from backup filenames.
func BenchmarkName2time(b *testing.B) {
	w := New(filepath.Join(b.TempDir(), "app.log"))
	w.SetConsole(false)
	defer w.Close()

	name := "app.2025-06-15-120000.log"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = w.name2time(name)
	}
}
