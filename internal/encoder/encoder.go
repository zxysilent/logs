package encoder

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

type Encoder struct{}

// PutKey appends a new key to the output JSON.
func (e Encoder) PutKey(dst []byte, key string) []byte {
	if len(dst) > 0 && dst[len(dst)-1] != '{' {
		dst = append(dst, ',')
	}
	return append(e.PutString(dst, key), ':')
}

// PutNil inserts a 'Nil' object into the dst byte array.
func (Encoder) PutNil(dst []byte) []byte {
	return append(dst, "nil"...)
}

// PutBeginMarker inserts a map start into the dst byte array.
func (Encoder) PutBeginMarker(dst []byte) []byte {
	return append(dst, '{')
}

// PutEndMarker inserts a map end into the dst byte array.
func (Encoder) PutEndMarker(dst []byte) []byte {
	return append(dst, '}')
}

// PutLineBreak appends a line break.
func (Encoder) PutLineBreak(dst []byte) []byte {
	return append(dst, '\n')
}

// PutArrayStart adds markers to indicate the start of an array.
func (Encoder) PutArrayStart(dst []byte) []byte {
	return append(dst, '[')
}

// PutArrayEnd adds markers to indicate the end of an array.
func (Encoder) PutArrayEnd(dst []byte) []byte {
	return append(dst, ']')
}

// PutArrayDelim adds markers to indicate end of a particular array element.
func (Encoder) PutArrayDelim(dst []byte) []byte {
	if len(dst) > 0 {
		return append(dst, ',')
	}
	return dst
}

// PutBool converts the input bool to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutBool(dst []byte, val bool) []byte {
	return strconv.AppendBool(dst, val)
}

// PutBools encodes the input bools to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutBools(dst []byte, vals []bool) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendBool(dst, vals[0])
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendBool(append(dst, ','), val)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutInt converts the input int to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInts encodes the input ints to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutInts(dst []byte, vals []int) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutInt8 converts the input []int8 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInts8 encodes the input int8s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutInts8(dst []byte, vals []int8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutInt16 converts the input int16 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInts16 encodes the input int16s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutInts16(dst []byte, vals []int16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutInt32 converts the input int32 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInts32 encodes the input int32s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutInts32(dst []byte, vals []int32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, int64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), int64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutInt64 converts the input int64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

// PutInts64 encodes the input int64s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutInts64(dst []byte, vals []int64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendInt(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendInt(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutUint converts the input uint to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUints encodes the input uints to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutUints(dst []byte, vals []uint) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutUint8 converts the input uint8 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUints8 encodes the input uint8s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutUints8(dst []byte, vals []uint8) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutUint16 converts the input uint16 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUints16 encodes the input uint16s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutUints16(dst []byte, vals []uint16) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutUint32 converts the input uint32 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUints32 encodes the input uint32s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutUints32(dst []byte, vals []uint32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, uint64(vals[0]), 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), uint64(val), 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutUint64 converts the input uint64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint64(dst []byte, val uint64) []byte {
	return strconv.AppendUint(dst, val, 10)
}

// PutUints64 encodes the input uint64s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutUints64(dst []byte, vals []uint64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = strconv.AppendUint(dst, vals[0], 10)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = strconv.AppendUint(append(dst, ','), val, 10)
		}
	}
	dst = append(dst, ']')
	return dst
}

func appendFloat(dst []byte, val float64, bitSize int) []byte {
	// JSON does not permit NaN or Infinity. A typical JSON encoder would fail
	// with an error, but a logging library wants the data to get through so we
	// make a tradeoff and store those types as string.
	switch {
	case math.IsNaN(val):
		return append(dst, `"NaN"`...)
	case math.IsInf(val, 1):
		return append(dst, `"+Inf"`...)
	case math.IsInf(val, -1):
		return append(dst, `"-Inf"`...)
	}
	return strconv.AppendFloat(dst, val, 'f', -1, bitSize)
}

// PutFloat32 converts the input float32 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutFloat32(dst []byte, val float32) []byte {
	return appendFloat(dst, float64(val), 32)
}

// PutFloats32 encodes the input float32s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutFloats32(dst []byte, vals []float32) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, float64(vals[0]), 32)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), float64(val), 32)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutFloat64 converts the input float64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

// PutFloats64 encodes the input float64s to json and
// appends the encoded string list to the input byte slice.
func (Encoder) PutFloats64(dst []byte, vals []float64) []byte {
	if len(vals) == 0 {
		return append(dst, '[', ']')
	}
	dst = append(dst, '[')
	dst = appendFloat(dst, vals[0], 64)
	if len(vals) > 1 {
		for _, val := range vals[1:] {
			dst = appendFloat(append(dst, ','), val, 64)
		}
	}
	dst = append(dst, ']')
	return dst
}

// PutAny marshals the input interface to a string and
// appends the encoded string to the input byte slice.
func (e Encoder) PutAny(dst []byte, i interface{}) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return e.PutString(dst, fmt.Sprintf("marshaling error: %v", err))
	}
	return append(dst, marshaled...)
}

// PutType appends the parameter type (as a string) to the input byte slice.
func (e Encoder) PutType(dst []byte, i interface{}) []byte {
	if i == nil {
		return e.PutString(dst, "<nil>")
	}
	return e.PutString(dst, reflect.TypeOf(i).String())
}

// PutObjectData takes in an object that is already in a byte array
// and adds it to the dst.
func (Encoder) PutObjectData(dst []byte, o []byte) []byte {
	// Three conditions apply here:
	// 1. new content starts with '{' - which should be dropped   OR
	// 2. new content starts with '{' - which should be replaced with ','
	//    to separate with existing content OR
	// 3. existing content has already other fields
	if o[0] == '{' {
		if len(dst) > 1 {
			dst = append(dst, ',')
		}
		o = o[1:]
	} else if len(dst) > 1 {
		dst = append(dst, ',')
	}
	return append(dst, o...)
}
