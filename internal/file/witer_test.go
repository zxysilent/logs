package file

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestName2time(t *testing.T) {
	f := New("../../logs/app.log")
	f.console = false
	t.Logf("%+v", f)
	f.delete(f.maxage)
}

func TestArchive2timeRejectsMalformedNames(t *testing.T) {
	w := &Writer{fname: "app", loc: time.UTC}
	for _, name := range []string{
		"app.tar.gz",
		"appx.2024-01-02.tar.gz",
		"app.2024-13-02.tar.gz",
		"app.2024-01-02x.tar.gz",
	} {
		if _, err := w.archive2time(name); err == nil {
			t.Errorf("archive2time(%q) expected an error", name)
		}
	}
}

func TestName2timeRejectsMalformedNames(t *testing.T) {
	w := &Writer{fname: "app", fsuffix: ".log", loc: time.UTC}
	for _, name := range []string{
		"app.log",
		"appx.2024-01-02-030405.log",
		"app.2024-13-02-030405.log",
		"app.2024-01-02-030405.txt",
	} {
		if _, err := w.name2time(name); err == nil {
			t.Errorf("name2time(%q) expected an error", name)
		}
	}
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
	w.SetConsole(false)
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
	w := New(path)
	w.SetConsole(false)
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
	w := New(path)
	w.SetConsole(false)
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
	w := New(path)
	w.SetConsole(false)
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

// createBackup writes a fake rotated backup file for the given offset in days.
func createBackup(t *testing.T, dir, fname, content string, day time.Time) string {
	t.Helper()
	name := fname + day.Format(".2006-01-02-150405") + ".log"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("create backup %s failed: %v", name, err)
	}
	return name
}

// readArchive extracts an archive and returns name -> content.
func readArchive(t *testing.T, path string) map[string]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open archive failed: %v", err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatalf("gzip reader failed: %v", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	out := map[string]string{}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar reader failed: %v", err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read %s failed: %v", hdr.Name, err)
		}
		out[hdr.Name] = string(data)
	}
	return out
}

// TestCompressDisabled verifies no archives are created when compress is off.
func TestCompressDisabled(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(64)

	old := time.Now().AddDate(0, 0, -9)
	createBackup(t, dir, "app", "old data", old)
	w.sweep()

	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), archiveSuffix) {
			t.Fatalf("archive created while compress disabled: %s", e.Name())
		}
	}
}

// TestCompressByDay verifies per-day archiving, original entry names,
// source removal, and the uncompressed window.
func TestCompressByDay(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(64) // retention far beyond keepDays; days-9 compresses, day-7 stays

	today := time.Now()
	nameOld1 := createBackup(t, dir, "app", "day-nine file one", today.AddDate(0, 0, -9))
	nameOld2 := createBackup(t, dir, "app", "day-nine file two", today.AddDate(0, 0, -9).Add(time.Hour))
	nameKept1 := createBackup(t, dir, "app", "day-two kept", today.AddDate(0, 0, -2))
	nameKept2 := createBackup(t, dir, "app", "today kept", today)
	nameRecent := createBackup(t, dir, "app", "seven-days-ago kept", today.AddDate(0, 0, -7))

	w.compress = true // set directly; SetCompress only toggles the flag
	w.sweep()
	archive := filepath.Join(dir, "app."+today.AddDate(0, 0, -9).Format("2006-01-02")+".tar.gz")
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive missing after sweep: %v", err)
	}

	got := readArchive(t, archive)
	if len(got) != 2 {
		t.Fatalf("expected 2 entries in archive, got %d: %v", len(got), got)
	}
	if got[nameOld1] != "day-nine file one" || got[nameOld2] != "day-nine file two" {
		t.Fatalf("archive entries mismatch: %v", got)
	}

	// sources of archived day removed
	if _, err := os.Stat(filepath.Join(dir, nameOld1)); !os.IsNotExist(err) {
		t.Fatalf("archived source should be removed, err=%v", err)
	}
	// window files untouched
	for _, kept := range []string{nameKept1, nameKept2, nameRecent} {
		if _, err := os.Stat(filepath.Join(dir, kept)); err != nil {
			t.Fatalf("in-window file %s should be kept: %v", kept, err)
		}
	}
}

// TestCompressIdempotent verifies repeated sweeps do not duplicate archives.
func TestCompressIdempotent(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(64)

	today := time.Now()
	createBackup(t, dir, "app", "old data", today.AddDate(0, 0, -9))
	w.compress = true // set directly; SetCompress only toggles the flag

	archive := filepath.Join(dir, "app."+today.AddDate(0, 0, -9).Format("2006-01-02")+".tar.gz")
	for i := 0; i < 3; i++ {
		w.sweep()
	}
	got := readArchive(t, archive)
	wantKey := "app" + today.AddDate(0, 0, -9).Format(".2006-01-02-150405") + ".log"
	if len(got) != 1 || got[wantKey] != "old data" {
		t.Fatalf("archive content corrupted by repeated sweep: want key %#v, got %#v", wantKey, got)
	}
}

// TestCompressShortRetention verifies that with maxage < keepDays nothing is
// archived: retention deletes old backups before they reach the 7-day window.
func TestCompressShortRetention(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(4)

	today := time.Now()
	nameStale := createBackup(t, dir, "app", "five days old", today.AddDate(0, 0, -5))
	nameInWindow := createBackup(t, dir, "app", "three days old", today.AddDate(0, 0, -3))
	nameKept := createBackup(t, dir, "app", "two days old", today.AddDate(0, 0, -2))
	w.compress = true // set directly; SetCompress only toggles the flag
	w.sweep()

	// No archive of any kind should exist.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir failed: %v", err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), archiveSuffix) {
			t.Fatalf("no archive expected with maxage < keepDays, found: %s", e.Name())
		}
	}
	// The 5-days-old backup is removed by retention, not compression.
	if _, err := os.Stat(filepath.Join(dir, nameStale)); !os.IsNotExist(err) {
		t.Fatalf("5-days-old file should be removed by retention, err=%v", err)
	}
	// Backups inside the uncompressed window stay as plain files.
	if _, err := os.Stat(filepath.Join(dir, nameInWindow)); err != nil {
		t.Fatalf("3-days-old file should stay plain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, nameKept)); err != nil {
		t.Fatalf("2-days-old file should stay plain: %v", err)
	}
}

func TestCompressUsesLocalDateBoundary(t *testing.T) {
	location := time.FixedZone("UTC-8", -8*60*60)

	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.loc = location
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(64)
	w.SetCompress(true)

	today := time.Now().In(location)
	boundary := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, location).AddDate(0, 0, -keepDays)
	older := boundary.AddDate(0, 0, -1)
	nameBoundary := createBackup(t, dir, "app", "boundary day", boundary)
	nameOlder := createBackup(t, dir, "app", "older day", older)

	parsed, err := w.name2time(nameBoundary)
	if err != nil {
		t.Fatalf("parse local backup timestamp: %v", err)
	}
	if parsed.Location() != location {
		t.Fatalf("backup timestamp location = %v, want %v", parsed.Location(), location)
	}
	parsedArchive, err := w.archive2time("app." + boundary.Format("2006-01-02") + archiveSuffix)
	if err != nil {
		t.Fatalf("parse local archive date: %v", err)
	}
	if parsedArchive.Location() != location {
		t.Fatalf("archive date location = %v, want %v", parsedArchive.Location(), location)
	}

	w.sweep()
	if _, err := os.Stat(filepath.Join(dir, nameBoundary)); err != nil {
		t.Fatalf("boundary-day backup should remain uncompressed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, nameOlder)); !os.IsNotExist(err) {
		t.Fatalf("older backup should be archived, err=%v", err)
	}
	archive := filepath.Join(dir, "app."+older.Format("2006-01-02")+archiveSuffix)
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("older-day archive missing: %v", err)
	}
}

// TestDeleteArchives verifies archives beyond maxage are removed.
func TestDeleteArchives(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(7)

	oldArchive := filepath.Join(dir, "app."+time.Now().AddDate(0, 0, -30).Format("2006-01-02")+".tar.gz")
	if err := os.WriteFile(oldArchive, []byte("stub"), 0o644); err != nil {
		t.Fatalf("create old archive failed: %v", err)
	}
	freshArchive := filepath.Join(dir, "app."+time.Now().AddDate(0, 0, -3).Format("2006-01-02")+".tar.gz")
	if err := os.WriteFile(freshArchive, []byte("stub"), 0o644); err != nil {
		t.Fatalf("create fresh archive failed: %v", err)
	}

	w.delete(w.maxage)
	if _, err := os.Stat(oldArchive); !os.IsNotExist(err) {
		t.Fatalf("expired archive should be deleted, err=%v", err)
	}
	if _, err := os.Stat(freshArchive); err != nil {
		t.Fatalf("fresh archive should be kept: %v", err)
	}
}

// TestCompressTempCleanup verifies a failed archive leaves no temp file and
// keeps its sources intact.
func TestCompressTempCleanup(t *testing.T) {
	dir := t.TempDir()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	defer w.Close()
	w.SetMaxAge(64)

	today := time.Now()
	name := createBackup(t, dir, "app", "old data", today.AddDate(0, 0, -9))
	// Pre-create the .tmp target as a directory so os.Create fails.
	if err := os.Mkdir(filepath.Join(dir, "app."+today.AddDate(0, 0, -9).Format("2006-01-02")+".tar.gz.tmp"), 0o755); err != nil {
		t.Fatalf("mkdir stub failed: %v", err)
	}

	w.sweep()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.IsDir() && strings.HasSuffix(e.Name(), ".tmp") {
			continue // the stub directory itself
		}
		if strings.HasSuffix(e.Name(), ".tar.gz") {
			t.Fatalf("archive created despite failure: %s", e.Name())
		}
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("source must survive failed archive: %v", err)
	}
}

// TestCloseIdempotent verifies Close is safe to call multiple times.
func TestCloseIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w := New(path)
	w.SetConsole(false)
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
	w := New(path)
	w.SetConsole(false)
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
	w := New(path)
	defer w.Close()

	if !w.console {
		t.Fatalf("cons should be true")
	}
	if _, err := w.Write([]byte("to file and stderr\n")); err != nil {
		t.Fatalf("write with cons failed: %v", err)
	}
}
