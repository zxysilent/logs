package file

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// newBareWriter builds a Writer without the ticker goroutine so fuzz targets
// stay cheap. Callers must release w.file themselves.
func newBareWriter(dir string) *Writer {
	w := &Writer{
		fpath:   filepath.Join(dir, "app.log"),
		fdir:    dir,
		fname:   "app",
		fsuffix: ".log",
		loc:     testLoc,
		maxage:  defMaxage,
		maxsize: sizeMiB * defMaxsize,
		console: false,
	}
	setNow(w, testNow)
	return w
}

// release closes the open file of a bare writer, if any.
func release(w *Writer) {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}
}

// FuzzPasename checks the parser never panics, always reports the writer's
// location, and round-trips an accepted name byte for byte.
func FuzzPasename(f *testing.F) {
	for _, seed := range []string{
		"app.2026-01-02-030405.log",
		"app.2026-01-02-100530.tar.gz",
		"app.log",
		"app.tar.gz",
		"app.2026-01-02-030405.log.tmp",
		"appx.2026-01-02-030405.log",
		// time.Parse accepts a fractional second the layout does not describe,
		// so non-canonical names must stay unmanaged.
		"app.2026-01-02-030405,0.log",
		"app.2026-01-02-030405.5.log",
		"",
		"app.",
		"app.9999-99-99-999999.log",
	} {
		f.Add(seed)
	}
	w := &Writer{fname: "app", fsuffix: ".log", loc: testLoc}
	f.Fuzz(func(t *testing.T, name string) {
		got, kind, err := w.pasename(name)
		if err != nil {
			return
		}
		if got.Location() != testLoc {
			t.Fatalf("location = %v, want %v", got.Location(), testLoc)
		}
		var want string
		switch kind {
		case kindBackup:
			want = "app" + got.Format(nameLayout) + ".log"
		case kindArchive:
			want = "app" + got.Format(nameLayout) + archiveSuffix
		default:
			t.Fatalf("unexpected kind %v for %q", kind, name)
		}
		if want != name {
			t.Fatalf("round-trip mismatch: %q parsed as %v, reprints to %q", name, got, want)
		}
	})
}

// FuzzDeleteRetention checks the retention sweep never touches unmanaged files
// and never drops a file whose calendar day is still inside the window.
func FuzzDeleteRetention(f *testing.F) {
	f.Add("app.2026-03-01-000000.log", 14)
	f.Add("keep.txt", 400)
	f.Add("app.2026-03-05-120000.tar.gz", 0)
	f.Add("app.log", -1)
	f.Fuzz(func(t *testing.T, name string, daysAgo int) {
		if name == "" || name == "." || name == ".." ||
			strings.ContainsAny(name, `/\:`) || len(name) > 200 {
			t.Skip()
		}
		if daysAgo < 0 {
			daysAgo = -daysAgo
		}
		daysAgo %= 400
		const maxage = 10

		dir := t.TempDir()
		w := newBareWriter(dir)
		// Control file with a known age: it must live exactly as long as its
		// calendar day stays inside the window.
		controlDay := testNow.AddDate(0, 0, -daysAgo)
		control := filepath.Join(dir, "app"+controlDay.Format(nameLayout)+".log")
		if err := os.WriteFile(control, []byte("control"), 0o644); err != nil {
			t.Fatalf("write control failed: %v", err)
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte("payload"), 0o644); err != nil {
			t.Skipf("unwritable name %q: %v", name, err)
		}

		deadline := testNow.AddDate(0, 0, -maxage)
		w.delete(maxage)

		_, err := os.Stat(control)
		controlGone := os.IsNotExist(err)
		if expired := !w.dayof(controlDay).AddDate(0, 0, 1).After(deadline); expired != controlGone {
			t.Fatalf("control day %v: expired=%v but gone=%v", controlDay, expired, controlGone)
		}

		parsed, _, err := w.pasename(name)
		managed := err == nil
		_, statErr := os.Stat(path)
		gone := os.IsNotExist(statErr)

		if !managed {
			if gone {
				t.Fatalf("unmanaged file %q was deleted", name)
			}
			return
		}
		// Mirrors the production rule: the whole day must be past the deadline.
		expired := !w.dayof(parsed).AddDate(0, 0, 1).After(deadline)
		if expired && !gone {
			t.Fatalf("%q is past retention and should be deleted", name)
		}
		if !expired && gone {
			t.Fatalf("%q is inside retention and must survive", name)
		}
	})
}

// FuzzArchiveNoDataLoss checks a sweep either keeps every backup as a plain
// file or moves its exact bytes into the day archive, never losing content.
func FuzzArchiveNoDataLoss(f *testing.F) {
	f.Add("alpha", "beta", 9)
	f.Add("", "x", 30)
	f.Add("time=2026-03-06T00:00:00 level=INF msg=hi\n", "\x00\xff", 8)
	f.Fuzz(func(t *testing.T, first, second string, daysAgo int) {
		if daysAgo < 0 {
			daysAgo = -daysAgo
		}
		daysAgo = keepDays + 1 + daysAgo%40
		day := testNow.AddDate(0, 0, -daysAgo)

		dir := t.TempDir()
		w := newBareWriter(dir)
		w.compress = true
		first, second = truncate(first), truncate(second)
		name1 := createBackup(t, dir, "app", first, day)
		name2 := createBackup(t, dir, "app", second, day.Add(time.Minute))

		w.archive()

		found := map[string]string{}
		for _, entry := range dirNames(t, dir) {
			if strings.HasSuffix(entry, archiveSuffix) {
				for k, v := range readArchive(t, filepath.Join(dir, entry)) {
					found[k] = v
				}
				continue
			}
			if strings.HasSuffix(entry, ".log") && entry != "app.log" {
				data, err := os.ReadFile(filepath.Join(dir, entry))
				if err != nil {
					t.Fatalf("read %s failed: %v", entry, err)
				}
				found[entry] = string(data)
			}
		}
		if found[name1] != first {
			t.Fatalf("lost %q: got %q", first, found[name1])
		}
		if found[name2] != second {
			t.Fatalf("lost %q: got %q", second, found[name2])
		}
	})
}

// FuzzWriteRotate checks arbitrary payloads never panic the write path and are
// fully accounted for.
func FuzzWriteRotate(f *testing.F) {
	f.Add([]byte("time=2026-03-15T00:00:00 level=INF msg=hi\n"))
	f.Add([]byte("short"))
	f.Add([]byte(""))
	f.Add(make([]byte, 64))
	f.Fuzz(func(t *testing.T, payload []byte) {
		if len(payload) > 4096 {
			payload = payload[:4096]
		}
		dir := t.TempDir()
		w := newBareWriter(dir)
		defer release(w)

		n, err := w.Write(payload)
		if err != nil {
			t.Fatalf("write failed: %v", err)
		}
		if n != len(payload) {
			t.Fatalf("short write: %d of %d", n, len(payload))
		}
		if err := w.flush(); err != nil {
			t.Fatalf("flush failed: %v", err)
		}
		data, err := os.ReadFile(w.fpath)
		if err != nil {
			t.Fatalf("read back failed: %v", err)
		}
		if string(data) != string(payload) {
			t.Fatalf("payload mismatch: wrote %q, read %q", payload, data)
		}
	})
}

// truncate keeps fuzz payloads small enough to stay fast.
func truncate(s string) string {
	const max = 512
	if len(s) > max {
		return s[:max]
	}
	return s
}
