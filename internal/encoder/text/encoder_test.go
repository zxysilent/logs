package text

import (
	"math"
	"reflect"
	"testing"
)

func TestPutType(t *testing.T) {
	var enc = Encoder{}
	w := map[string]func(interface{}) []byte{
		"PutInt":     func(v interface{}) []byte { return enc.PutInt([]byte{}, v.(int)) },
		"PutInt8":    func(v interface{}) []byte { return enc.PutInt8([]byte{}, v.(int8)) },
		"PutInt16":   func(v interface{}) []byte { return enc.PutInt16([]byte{}, v.(int16)) },
		"PutInt32":   func(v interface{}) []byte { return enc.PutInt32([]byte{}, v.(int32)) },
		"PutInt64":   func(v interface{}) []byte { return enc.PutInt64([]byte{}, v.(int64)) },
		"PutUint":    func(v interface{}) []byte { return enc.PutUint([]byte{}, v.(uint)) },
		"PutUint8":   func(v interface{}) []byte { return enc.PutUint8([]byte{}, v.(uint8)) },
		"PutUint16":  func(v interface{}) []byte { return enc.PutUint16([]byte{}, v.(uint16)) },
		"PutUint32":  func(v interface{}) []byte { return enc.PutUint32([]byte{}, v.(uint32)) },
		"PutUint64":  func(v interface{}) []byte { return enc.PutUint64([]byte{}, v.(uint64)) },
		"PutFloat32": func(v interface{}) []byte { return enc.PutFloat32([]byte{}, v.(float32)) },
		"PutFloat64": func(v interface{}) []byte { return enc.PutFloat64([]byte{}, v.(float64)) },
	}
	tests := []struct {
		name  string
		fn    string
		input interface{}
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
