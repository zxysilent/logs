package json

import (
	"encoding/json"
	"fmt"
	"math"
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

// PutBegin inserts a map start into the dst byte array.
func (Encoder) PutBegin(dst []byte) []byte {
	return append(dst, '{')
}

// PutEnd inserts a map end into the dst byte array.
func (Encoder) PutEnd(dst []byte) []byte {
	return append(dst, '}')
}

// PutDelim adds markers to indicate end of a particular element.
func (Encoder) PutDelim(dst []byte) []byte {
	if len(dst) > 0 {
		return append(dst, ',')
	}
	return dst
}

// PutBreak appends a line break.
func (Encoder) PutBreak(dst []byte) []byte {
	return append(dst, '\n')
}

// PutBool converts the input bool to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutBool(dst []byte, val bool) []byte {
	return strconv.AppendBool(dst, val)
}

// PutInt converts the input int to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt8 converts the input []int8 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt16 converts the input int16 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt32 converts the input int32 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt64 converts the input int64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

// PutUint converts the input uint to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint8 converts the input uint8 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint16 converts the input uint16 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint32 converts the input uint32 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint64 converts the input uint64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutUint64(dst []byte, val uint64) []byte {
	return strconv.AppendUint(dst, val, 10)
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

// PutFloat64 converts the input float64 to a string and
// appends the encoded string to the input byte slice.
func (Encoder) PutFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

// PutInterface marshals the input interface to a string and
// appends the encoded string to the input byte slice.
func (e Encoder) PutAny(dst []byte, i interface{}) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return e.PutString(dst, fmt.Sprintf("marshaling error: %v", err))
	}
	return append(dst, marshaled...)
}
