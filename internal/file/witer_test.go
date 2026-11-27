package file

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// testLoc and testNow pin the clock so day boundaries stay deterministic.
var (
	testLoc = time.FixedZone("CST", 8*60*60)
	testNow = time.Date(2026, 3, 15, 10, 0, 0, 0, testLoc)
)

// newTestWriter returns a writer over dir/app.log with a frozen clock.
func newTestWriter(t *testing.T, dir string, at time.Time) *Writer {
	t.Helper()
	w := New(filepath.Join(dir, "app.log"))
	w.SetConsole(false)
	w.loc = at.Location()
	setNow(w, at)
	t.Cleanup(func() { _ = w.Close() })
	return w
}

// setNow moves the writer's frozen clock.
func setNow(w *Writer, at time.Time) {
	ts := at
	w.now = func() time.Time { return ts }
}

// archiveName is the archive file a sweep at `at` produces for `day`.
func archiveName(day, at time.Time) string {
	stamp := time.Date(day.Year(), day.Month(), day.Day(), at.Hour(), at.Minute(), at.Second(), 0, day.Location())
	return "app" + stamp.Format(nameLayout) + archiveSuffix
}

// dirNames lists the file names in dir.
func dirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s failed: %v", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// countArchives counts tar.gz files in dir.
func countArchives(t *testing.T, dir string) int {
	t.Helper()
	n := 0
	for _, name := range dirNames(t, dir) {
		if strings.HasSuffix(name, archiveSuffix) {
			n++
		}
	}
	return n
}

// TestPasename verifies one parser covers backups and archives, differing only
// by suffix.
func TestPasename(t *testing.T) {
	w := &Writer{fname: "app", fsuffix: ".log", loc: testLoc}
	cases := []struct {
		name string
		kind nameKind
		want time.Time
	}{
		{"app.2026-01-02-030405.log", kindBackup, time.Date(2026, 1, 2, 3, 4, 5, 0, testLoc)},
		{"app.2026-01-02-100530.tar.gz", kindArchive, time.Date(2026, 1, 2, 10, 5, 30, 0, testLoc)},
	}
	for _, c := range cases {
		got, kind, err := w.pasename(c.name)
		if err != nil {
			t.Fatalf("pasename(%q) failed: %v", c.name, err)
		}
		if kind != c.kind {
			t.Errorf("pasename(%q) kind = %v, want %v", c.name, kind, c.kind)
		}
		if !got.Equal(c.want) {
			t.Errorf("pasename(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

// TestPasenameRejectsMalformedNames verifies unmanaged files are ignored
// instead of being archived or deleted.
func TestPasenameRejectsMalformedNames(t *testing.T) {
	w := &Writer{fname: "app", fsuffix: ".log", loc: testLoc}
	for _, name := range []string{
		"app.log",
		"app.tar.gz",
		"appx.2026-01-02-030405.log",
		"app.2026-13-02-030405.log",
		"app.2026-1-2-030405.log",
		"app.2026-01-02-030405.txt",
		"app.2026-01-02-030405.log.tmp",
		"app.2026-01-02-030405.tar.gz.tmp",
		"app.2026-01-02x-030405.log",
	} {
		if _, _, err := w.pasename(name); err == nil {
			t.Errorf("pasename(%q) expected an error", name)
		}
	}
	// The backup name is only valid while its suffix matches the writer.
	if _, _, err := w.pasename("app.2026-01-02-030405.log"); err != nil {
		t.Errorf("valid backup rejected: %v", err)
	}
	if _, _, err := (&Writer{fname: "app", fsuffix: ".txt", loc: testLoc}).pasename("app.2026-01-02-030405.log"); err == nil {
		t.Errorf("backup with foreign suffix should be rejected")
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
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)

	old := createBackup(t, dir, "app", "old data", testNow.AddDate(0, 0, -9))
	w.sweep()

	if n := countArchives(t, dir); n != 0 {
		t.Fatalf("archives created while compress disabled: %d", n)
	}
	if _, err := os.Stat(filepath.Join(dir, old)); err != nil {
		t.Fatalf("backup should stay plain when compress is off: %v", err)
	}
}

// TestCompressByDay verifies per-day archiving, the derived archive name,
// original entry names, source removal, and the uncompressed window.
func TestCompressByDay(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow) // 2026-03-15 10:00:00
	w.SetMaxAge(64)
	w.SetCompress(true)

	oldDay := testNow.AddDate(0, 0, -9)
	edgeDay := testNow.AddDate(0, 0, -7)
	nameOld1 := createBackup(t, dir, "app", "day-nine file one", oldDay)
	nameOld2 := createBackup(t, dir, "app", "day-nine file two", oldDay.Add(time.Hour))
	nameKept1 := createBackup(t, dir, "app", "six days kept", testNow.AddDate(0, 0, -6))
	nameKept2 := createBackup(t, dir, "app", "today kept", testNow)
	nameEdge := createBackup(t, dir, "app", "seven days archived", edgeDay)

	w.sweep()

	archive := filepath.Join(dir, archiveName(oldDay, testNow))
	if want := "app.2026-03-06-100000.tar.gz"; filepath.Base(archive) != want {
		t.Fatalf("archive name = %s, want %s", filepath.Base(archive), want)
	}
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

	// Day -7 is the first day outside the uncompressed window.
	edgeArchive := filepath.Join(dir, archiveName(edgeDay, testNow))
	if _, err := os.Stat(edgeArchive); err != nil {
		t.Fatalf("day-7 backup should be archived: %v", err)
	}
	if got := readArchive(t, edgeArchive); got[nameEdge] != "seven days archived" {
		t.Fatalf("day-7 archive content mismatch: %v", got)
	}

	// sources of archived days removed
	for _, removed := range []string{nameOld1, nameOld2, nameEdge} {
		if _, err := os.Stat(filepath.Join(dir, removed)); !os.IsNotExist(err) {
			t.Fatalf("archived source %s should be removed, err=%v", removed, err)
		}
	}
	// window files untouched
	for _, kept := range []string{nameKept1, nameKept2} {
		if _, err := os.Stat(filepath.Join(dir, kept)); err != nil {
			t.Fatalf("in-window file %s should be kept: %v", kept, err)
		}
	}
	if n := countArchives(t, dir); n != 2 {
		t.Fatalf("archives created = %d, want 2", n)
	}
}

// TestCompressIdempotent verifies repeated sweeps, including ones running at a
// different time of day, never produce a second archive for a day.
func TestCompressIdempotent(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	oldDay := testNow.AddDate(0, 0, -9)
	createBackup(t, dir, "app", "old data", oldDay)

	archive := filepath.Join(dir, archiveName(oldDay, testNow))
	for i := 0; i < 3; i++ {
		w.sweep()
	}
	got := readArchive(t, archive)
	wantKey := "app" + oldDay.Format(nameLayout) + ".log"
	if len(got) != 1 || got[wantKey] != "old data" {
		t.Fatalf("archive content corrupted by repeated sweep: want key %#v, got %#v", wantKey, got)
	}

	// A later sweep at a different time of day must not add a second archive.
	setNow(w, testNow.Add(90*time.Minute+15*time.Second))
	w.sweep()
	if n := countArchives(t, dir); n != 1 {
		t.Fatalf("archives after clock moved = %d, want 1: %v", n, dirNames(t, dir))
	}
	if _, err := os.Stat(filepath.Join(dir, "app.2026-03-06-113015.tar.gz")); !os.IsNotExist(err) {
		t.Fatalf("second archive for the same day should not exist, err=%v", err)
	}
}

// TestCompressShortRetention verifies archiving is skipped when maxage cannot
// outlive the uncompressed window: nothing is packed just to be dropped.
func TestCompressShortRetention(t *testing.T) {
	for _, maxage := range []int{keepDays, 0, -1} {
		dir := t.TempDir()
		w := newTestWriter(t, dir, testNow)
		w.SetMaxAge(maxage)
		w.SetCompress(true)

		stale := createBackup(t, dir, "app", "nine days old", testNow.AddDate(0, 0, -9))
		w.sweep()

		if n := countArchives(t, dir); n != 0 {
			t.Fatalf("maxage=%d: archives = %d, want 0", maxage, n)
		}
		for _, name := range dirNames(t, dir) {
			if strings.HasSuffix(name, ".tmp") {
				t.Fatalf("maxage=%d: left a temp file: %s", maxage, name)
			}
		}
		_, err := os.Stat(filepath.Join(dir, stale))
		if maxage <= 0 {
			if err != nil {
				t.Fatalf("maxage=%d: retention disabled, backup should stay: %v", maxage, err)
			}
			continue
		}
		if !os.IsNotExist(err) {
			t.Fatalf("maxage=%d: 9-days-old backup should expire, err=%v", maxage, err)
		}
	}
}

func TestCompressUsesLocalDateBoundary(t *testing.T) {
	location := time.FixedZone("UTC-8", -8*60*60)
	now := time.Date(2026, 3, 15, 10, 0, 0, 0, location)

	dir := t.TempDir()
	w := newTestWriter(t, dir, now)
	w.SetMaxAge(64)
	w.SetCompress(true)

	inWindow := createBackup(t, dir, "app", "six days old", now.AddDate(0, 0, -6))
	older := createBackup(t, dir, "app", "seven days old", now.AddDate(0, 0, -7))

	parsed, kind, err := w.pasename(older)
	if err != nil {
		t.Fatalf("parse local backup timestamp: %v", err)
	}
	if parsed.Location() != location {
		t.Fatalf("backup timestamp location = %v, want %v", parsed.Location(), location)
	}
	if kind != kindBackup {
		t.Fatalf("kind = %v, want kindBackup", kind)
	}

	w.sweep()
	if _, err := os.Stat(filepath.Join(dir, inWindow)); err != nil {
		t.Fatalf("six-days-old backup should stay plain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, older)); !os.IsNotExist(err) {
		t.Fatalf("seven-days-old backup should be archived, err=%v", err)
	}
	archive := filepath.Join(dir, archiveName(now.AddDate(0, 0, -7), now))
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("archive missing: %v", err)
	}
}

// TestCompressCrossMonthYear checks the day window across a year boundary.
func TestCompressCrossMonthYear(t *testing.T) {
	now := time.Date(2027, 1, 2, 0, 0, 30, 0, testLoc)
	dir := t.TempDir()
	w := newTestWriter(t, dir, now)
	w.SetMaxAge(64)
	w.SetCompress(true)

	december := time.Date(2026, 12, 20, 23, 59, 59, 0, testLoc)
	inWindow := time.Date(2026, 12, 27, 0, 0, 0, 0, testLoc) // exactly keepDays-1 back
	oldName := createBackup(t, dir, "app", "december", december)
	keepName := createBackup(t, dir, "app", "in window", inWindow)

	w.sweep()
	archive := filepath.Join(dir, "app.2026-12-20-000030.tar.gz")
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("december archive missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, oldName)); !os.IsNotExist(err) {
		t.Fatalf("december backup should be archived, err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, keepName)); err != nil {
		t.Fatalf("in-window backup should stay plain: %v", err)
	}
}

// TestCompressDSTBoundary checks the window on a DST transition day.
func TestCompressDSTBoundary(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Skipf("tzdata unavailable: %v", err)
	}
	now := time.Date(2026, 3, 9, 12, 0, 0, 0, loc) // DST started 2026-03-08
	dir := t.TempDir()
	w := newTestWriter(t, dir, now)
	w.SetMaxAge(64)
	w.SetCompress(true)

	old := time.Date(2026, 3, 1, 23, 30, 0, 0, loc)
	keep := time.Date(2026, 3, 3, 0, 30, 0, 0, loc)
	createBackup(t, dir, "app", "before dst", old)
	keepName := createBackup(t, dir, "app", "in window", keep)

	w.sweep()
	if _, err := os.Stat(filepath.Join(dir, "app.2026-03-01-120000.tar.gz")); err != nil {
		t.Fatalf("pre-DST day should be archived: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, keepName)); err != nil {
		t.Fatalf("in-window backup should stay plain: %v", err)
	}
}

// TestDeleteUnifiedDayBoundary verifies backups and archives share one
// retention rule: a file goes away once its whole day ends before
// now-maxage, so removal is at most one day late and never early.
func TestDeleteUnifiedDayBoundary(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow) // 2026-03-15 10:00, maxage=10 -> deadline 2026-03-05 10:00
	const maxage = 10

	insideDay := time.Date(2026, 3, 5, 0, 0, 0, 0, testLoc)
	outsideDay := time.Date(2026, 3, 4, 0, 0, 0, 0, testLoc)
	inside := []string{
		createBackup(t, dir, "app", "start of deadline day", insideDay),
		createBackup(t, dir, "app", "end of deadline day", insideDay.Add(23*time.Hour+59*time.Minute+59*time.Second)),
		"app.2026-03-05-093000.tar.gz",
	}
	outside := []string{
		createBackup(t, dir, "app", "start of previous day", outsideDay),
		createBackup(t, dir, "app", "end of previous day", outsideDay.Add(23*time.Hour+59*time.Minute+59*time.Second)),
		"app.2026-03-04-235959.tar.gz",
	}
	for _, name := range inside {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("keep"), 0o644); err != nil {
			t.Fatalf("create %s failed: %v", name, err)
		}
	}
	for _, name := range outside {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("drop"), 0o644); err != nil {
			t.Fatalf("create %s failed: %v", name, err)
		}
	}

	w.delete(maxage)

	for _, name := range inside {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s is inside retention and must survive: %v", name, err)
		}
	}
	for _, name := range outside {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Fatalf("%s is past retention and must be removed, err=%v", name, err)
		}
	}
}

// TestCompressTempCleanup verifies a failed archive keeps its sources intact.
func TestCompressTempCleanup(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	day := testNow.AddDate(0, 0, -9)
	name := createBackup(t, dir, "app", "old data", day)
	// Pre-create the temp target as a directory so os.Create fails.
	if err := os.Mkdir(w.tmppath(day), 0o755); err != nil {
		t.Fatalf("mkdir stub failed: %v", err)
	}

	w.sweep()
	if n := countArchives(t, dir); n != 0 {
		t.Fatalf("archive created despite failure: %d", n)
	}
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("source must survive failed archive: %v", err)
	}
}

// TestCompressStaleTempCleanup verifies temp files left by an interrupted
// attempt are dropped once the day is already archived.
func TestCompressStaleTempCleanup(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	day := testNow.AddDate(0, 0, -9)
	archive := filepath.Join(dir, archiveName(day, testNow))
	if err := os.WriteFile(archive, []byte("stub"), 0o644); err != nil {
		t.Fatalf("create archive stub failed: %v", err)
	}
	tmp := w.tmppath(day)
	if err := os.WriteFile(tmp, []byte("partial"), 0o644); err != nil {
		t.Fatalf("create temp stub failed: %v", err)
	}

	w.sweep()
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Fatalf("stale temp file should be removed, err=%v", err)
	}
	if _, err := os.Stat(archive); err != nil {
		t.Fatalf("existing archive must be kept: %v", err)
	}
}

// TestCompressLateBackupKeepsArchive verifies a backup that shows up for an
// already archived day stays plain and never rewrites the archive.
func TestCompressLateBackupKeepsArchive(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	day := testNow.AddDate(0, 0, -9)
	createBackup(t, dir, "app", "first", day)
	w.sweep()

	archive := filepath.Join(dir, archiveName(day, testNow))
	got := readArchive(t, archive)
	if len(got) != 1 {
		t.Fatalf("archive should hold one entry, got %v", got)
	}

	// A stale-mtime restart can produce a new backup for the same old day.
	setNow(w, testNow.Add(5*time.Minute))
	late := createBackup(t, dir, "app", "late", day.Add(time.Hour))
	w.sweep()

	if n := countArchives(t, dir); n != 1 {
		t.Fatalf("archives after late backup = %d, want 1: %v", n, dirNames(t, dir))
	}
	if got := readArchive(t, archive); len(got) != 1 || got[wantKeyFor(day)] != "first" {
		t.Fatalf("archive must not be rewritten: %v", got)
	}
	if _, err := os.Stat(filepath.Join(dir, late)); err != nil {
		t.Fatalf("late backup should stay plain until retention: %v", err)
	}
}

// wantKeyFor returns the tar entry name of the first backup created for day.
func wantKeyFor(day time.Time) string {
	return "app" + day.Format(nameLayout) + ".log"
}

// TestCompressArchivePathOccupied verifies an occupied archive path is treated
// as already archived: nothing is overwritten and no source is dropped.
func TestCompressArchivePathOccupied(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	day := testNow.AddDate(0, 0, -9)
	name := createBackup(t, dir, "app", "must survive", day)
	occupied := filepath.Join(dir, archiveName(day, testNow))
	if err := os.WriteFile(occupied, []byte("existing archive"), 0o644); err != nil {
		t.Fatalf("create occupied archive failed: %v", err)
	}

	w.sweep()
	if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
		t.Fatalf("source must survive when the archive path is occupied: %v", err)
	}
	data, err := os.ReadFile(occupied)
	if err != nil {
		t.Fatalf("occupied archive disappeared: %v", err)
	}
	if string(data) != "existing archive" {
		t.Fatalf("occupied archive was overwritten: %q", data)
	}
}

// TestConcurrentWritesWithSweeps verifies concurrent day-crossing writes and
// sweeps stay race-free and still keep exactly one archive per day.
func TestConcurrentWritesWithSweeps(t *testing.T) {
	dir := t.TempDir()
	w := newTestWriter(t, dir, testNow)
	w.SetMaxAge(64)
	w.SetCompress(true)

	oldDay := testNow.AddDate(0, 0, -9)
	for i := 0; i < 5; i++ {
		createBackup(t, dir, "app", "old payload", oldDay.Add(time.Duration(i)*time.Minute))
	}

	errCh := make(chan error, 512)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			date := "2026-03-15"
			if i%2 == 1 {
				date = "2026-03-14"
			}
			line := []byte("time=" + date + "T00:00:00 concurrent\n")
			for j := 0; j < 40; j++ {
				if _, err := w.Write(line); err != nil {
					errCh <- err
					return
				}
			}
		}(i)
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				w.sweep()
			}
		}()
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatalf("concurrent write failed: %v", err)
	}
	if err := w.flush(); err != nil {
		t.Fatalf("flush failed: %v", err)
	}

	if n := countArchives(t, dir); n != 1 {
		t.Fatalf("archives = %d, want 1: %v", n, dirNames(t, dir))
	}
	got := readArchive(t, filepath.Join(dir, archiveName(oldDay, testNow)))
	if len(got) != 5 {
		t.Fatalf("archive entries = %d, want 5: %v", len(got), got)
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
