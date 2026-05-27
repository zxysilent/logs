package textenc

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestPutType(t *testing.T) {
	w := map[string]func(any) []byte{
		"PutInt":     func(v any) []byte { return PutInt([]byte{}, v.(int)) },
		"PutInt8":    func(v any) []byte { return PutInt8([]byte{}, v.(int8)) },
		"PutInt16":   func(v any) []byte { return PutInt16([]byte{}, v.(int16)) },
		"PutInt32":   func(v any) []byte { return PutInt32([]byte{}, v.(int32)) },
		"PutInt64":   func(v any) []byte { return PutInt64([]byte{}, v.(int64)) },
		"PutUint":    func(v any) []byte { return PutUint([]byte{}, v.(uint)) },
		"PutUint8":   func(v any) []byte { return PutUint8([]byte{}, v.(uint8)) },
		"PutUint16":  func(v any) []byte { return PutUint16([]byte{}, v.(uint16)) },
		"PutUint32":  func(v any) []byte { return PutUint32([]byte{}, v.(uint32)) },
		"PutUint64":  func(v any) []byte { return PutUint64([]byte{}, v.(uint64)) },
		"PutFloat32": func(v any) []byte { return PutFloat32([]byte{}, v.(float32)) },
		"PutFloat64": func(v any) []byte { return PutFloat64([]byte{}, v.(float64)) },
	}
	tests := []struct {
		name  string
		fn    string
		input any
		want  []byte
	}{
		{"PutInt8(math.MaxInt8)", "PutInt8", int8(math.MaxInt8), []byte("127")},
		{"PutInt16(math.MaxInt16)", "PutInt16", int16(math.MaxInt16), []byte("32767")},
		{"PutInt32(math.MaxInt32)", "PutInt32", int32(math.MaxInt32), []byte("2147483647")},
		{"PutInt64(math.MaxInt64)", "PutInt64", int64(math.MaxInt64), []byte("9223372036854775807")},

		{"PutUint8(math.MaxUint8)", "PutUint8", uint8(math.MaxUint8), []byte("255")},
		{"PutUint16(math.MaxUint16)", "PutUint16", uint16(math.MaxUint16), []byte("65535")},
		{"PutUint32(math.MaxUint32)", "PutUint32", uint32(math.MaxUint32), []byte("4294967295")},
		{"PutUint64(math.MaxUint64)", "PutUint64", uint64(math.MaxUint64), []byte("18446744073709551615")},

		{"PutFloat32(-Inf)", "PutFloat32", float32(math.Inf(-1)), []byte(`"-Inf"`)},
		{"PutFloat32(+Inf)", "PutFloat32", float32(math.Inf(1)), []byte(`"+Inf"`)},
		{"PutFloat32(NaN)", "PutFloat32", float32(math.NaN()), []byte(`"NaN"`)},
		{"PutFloat32(0)", "PutFloat32", float32(0), []byte(`0`)},
		{"PutFloat32(-1.1)", "PutFloat32", float32(-1.1), []byte(`-1.1`)},
		{"PutFloat32(1e20)", "PutFloat32", float32(1e20), []byte(`100000000000000000000`)},
		{"PutFloat32(1e21)", "PutFloat32", float32(1e21), []byte(`1000000000000000000000`)},

		{"PutFloat64(-Inf)", "PutFloat64", float64(math.Inf(-1)), []byte(`"-Inf"`)},
		{"PutFloat64(+Inf)", "PutFloat64", float64(math.Inf(1)), []byte(`"+Inf"`)},
		{"PutFloat64(NaN)", "PutFloat64", float64(math.NaN()), []byte(`"NaN"`)},
		{"PutFloat64(0)", "PutFloat64", float64(0), []byte(`0`)},
		{"PutFloat64(-1.1)", "PutFloat64", float64(-1.1), []byte(`-1.1`)},
		{"PutFloat64(1e20)", "PutFloat64", float64(1e20), []byte(`100000000000000000000`)},
		{"PutFloat64(1e21)", "PutFloat64", float64(1e21), []byte(`1000000000000000000000`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w[tt.fn](tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

type stringerValue string

func (v stringerValue) String() string { return string(v) }

func TestEncoderHelpers(t *testing.T) {
	if got := string(PutKey(nil, "key")); got != "key=" {
		t.Fatalf("PutKey mismatch: %q", got)
	}
	if got := string(PutKey(nil, "sp ce")); got != `"sp ce"=` {
		t.Fatalf("PutKey quoted mismatch: %q", got)
	}
	if got := string(PutNil(nil)); got != "nil" {
		t.Fatalf("PutNil mismatch: %q", got)
	}
	if got := string(PutBegin([]byte("a"))); got != "a" {
		t.Fatalf("PutBegin mismatch: %q", got)
	}
	if got := string(PutEnd([]byte("a"))); got != "a" {
		t.Fatalf("PutEnd mismatch: %q", got)
	}
	if got := string(PutDelim(nil)); got != "" {
		t.Fatalf("PutDelim empty mismatch: %q", got)
	}
	if got := string(PutDelim([]byte("a"))); got != "a " {
		t.Fatalf("PutDelim mismatch: %q", got)
	}
	if got := string(PutBreak(nil)); got != "\n" {
		t.Fatalf("PutBreak mismatch: %q", got)
	}
	if got := string(PutBool(nil, true)); got != "true" {
		t.Fatalf("PutBool mismatch: %q", got)
	}
	if got := string(PutInt(nil, -3)); got != "-3" {
		t.Fatalf("PutInt mismatch: %q", got)
	}
	if got := string(PutUint(nil, 3)); got != "3" {
		t.Fatalf("PutUint mismatch: %q", got)
	}
	if got := string(PutNil(PutKey(nil, "k"))); !strings.Contains(got, "nil") {
		t.Fatalf("PutKey+PutNil mismatch: %q", got)
	}
	if got := string(PutStringer(nil, stringerValue("value with space"))); got != `"value with space"` {
		t.Fatalf("PutStringer mismatch: %q", got)
	}
	if got := string(PutStringer(nil, nil)); got != "nil" {
		t.Fatalf("PutStringer nil mismatch: %q", got)
	}
	if got := string(PutAny(nil, struct{ A int }{A: 1})); got != `{"A":1}` {
		t.Fatalf("PutAny struct mismatch: %q", got)
	}
	if got := string(PutAny(nil, func() {})); !strings.Contains(got, "marshaling error") {
		t.Fatalf("PutAny error path mismatch: %q", got)
	}
	if got := string(PutTime(nil, time.Date(2024, 1, 2, 3, 4, 5, 123456789, time.UTC))); got != `2024-01-02T03:04:05.123` {
		t.Fatalf("PutTime mismatch: %q", got)
	}
	if got := string(PutDuration(nil, 2*time.Hour+3*time.Minute+4*time.Second)); got != `2h3m4s` {
		t.Fatalf("PutDuration mismatch: %q", got)
	}
}

// ------------------------------------------------------------------
// Boundary tests
// ------------------------------------------------------------------

func TestPutTimeBoundary(t *testing.T) {
	tests := []struct {
		name string
		tm   time.Time
		want string
	}{
		{"epoch", time.Unix(0, 0).UTC(), "1970-01-01T00:00:00.000"},
		{"year9999", time.Date(9999, 12, 31, 23, 59, 59, 999000000, time.UTC), "9999-12-31T23:59:59.999"},
		{"year1", time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), "0001-01-01T00:00:00.000"},
		{"leapDay", time.Date(2024, 2, 29, 12, 0, 0, 0, time.UTC), "2024-02-29T12:00:00.000"},
		{"midnight", time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC), "2024-06-15T00:00:00.000"},
		{"msEdge", time.Date(2024, 1, 1, 0, 0, 0, 999000000, time.UTC), "2024-01-01T00:00:00.999"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(PutTime(nil, tt.tm)); got != tt.want {
				t.Errorf("PutTime = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPutKeyRawVsPutKey(t *testing.T) {
	// Non-special keys: PutKey and PutKeyRaw should produce same output.
	tests := []string{"key", "key_name", "k", "a.b", "a-b", "a1b2c3"}
	for _, key := range tests {
		gotRaw := string(PutKeyRaw(nil, key))
		gotKey := string(PutKey(nil, key))
		if gotRaw != gotKey {
			t.Errorf("PutKeyRaw(%q)=%q != PutKey(%q)=%q", key, gotRaw, key, gotKey)
		}
	}
	// Key with space: PutKeyRaw does NOT quote
	if got := string(PutKeyRaw(nil, "a b")); got != `a b=` {
		t.Errorf("PutKeyRaw with space: %q", got)
	}
}

// ------------------------------------------------------------------
// Fuzz tests
// ------------------------------------------------------------------

func FuzzPutKey(f *testing.F) {
	f.Add("normal")
	f.Add("key with space")
	f.Add(string(rune(0)))
	f.Fuzz(func(t *testing.T, key string) {
		_ = PutKey(nil, key)
		_ = PutKeyRaw(nil, key)
	})
}

func FuzzPutString(f *testing.F) {
	f.Add("hello")
	f.Add("multibyte\u276d")
	f.Add(string([]byte{0x00, 0x1f}))
	f.Fuzz(func(t *testing.T, s string) {
		_ = PutString(nil, s)
	})
}

func FuzzPutBytes(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte{0, 1, 0x1f, 0x7f, 0xFF})
	f.Fuzz(func(t *testing.T, data []byte) {
		_ = PutBytes(nil, data)
	})
}

func FuzzPutTime(f *testing.F) {
	f.Add(int64(0), int64(0))
	f.Add(int64(1<<60), int64(999999999))
	f.Fuzz(func(t *testing.T, sec int64, nsec int64) {
		if sec < -1<<60 || sec > 1<<60 {
			return
		}
		if nsec < 0 || nsec >= 1e9 {
			return
		}
		tm := time.Unix(sec, nsec)
		_ = PutTime(nil, tm)
	})
}

func FuzzPutAny(f *testing.F) {
	f.Add("hello")
	f.Fuzz(func(t *testing.T, val string) {
		_ = PutAny(nil, val)
		_ = PutAny(nil, 42)
		_ = PutAny(nil, []string{"a", "b"})
	})
}
