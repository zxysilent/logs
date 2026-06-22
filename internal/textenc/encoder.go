package textenc

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

// PutKey appends a new key to the logfmt output.
func PutKey(dst []byte, key string) []byte {
	if len(dst) > 0 {
		dst = append(dst, ' ')
	}
	return append(quoteString(dst, key, true), '=')
}

// PutKeyRaw appends a new key without quoting check.
// Use for internal keys that are known to be valid (no spaces/tabs).
func PutKeyRaw(dst []byte, key string) []byte {
	if len(dst) > 0 {
		dst = append(dst, ' ')
	}
	dst = append(dst, key...)
	return append(dst, '=')
}

// PutNil inserts a 'Nil' object into the dst byte array.
func PutNil(dst []byte) []byte {
	return append(dst, "nil"...)
}

// PutBegin marks the start of a record (no-op in logfmt).
func PutBegin(dst []byte) []byte {
	return dst
}

// PutEnd marks the end of a record (no-op in logfmt).
func PutEnd(dst []byte) []byte {
	return dst
}

// PutDelim appends a separator between elements.
func PutDelim(dst []byte) []byte {
	if len(dst) > 0 {
		return append(dst, ' ')
	}
	return dst
}

// PutBreak appends a line break.
func PutBreak(dst []byte) []byte {
	return append(dst, '\n')
}

// PutBool converts the input bool to a string and
// appends the encoded value to the input byte slice.
func PutBool(dst []byte, val bool) []byte {
	return strconv.AppendBool(dst, val)
}

// PutInt converts the input int to a string and
// appends the encoded value to the input byte slice.
func PutInt(dst []byte, val int) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt8 converts the input int8 to a string and
// appends the encoded value to the input byte slice.
func PutInt8(dst []byte, val int8) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt16 converts the input int16 to a string and
// appends the encoded value to the input byte slice.
func PutInt16(dst []byte, val int16) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt32 converts the input int32 to a string and
// appends the encoded value to the input byte slice.
func PutInt32(dst []byte, val int32) []byte {
	return strconv.AppendInt(dst, int64(val), 10)
}

// PutInt64 converts the input int64 to a string and
// appends the encoded value to the input byte slice.
func PutInt64(dst []byte, val int64) []byte {
	return strconv.AppendInt(dst, val, 10)
}

// PutUint converts the input uint to a string and
// appends the encoded value to the input byte slice.
func PutUint(dst []byte, val uint) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint8 converts the input uint8 to a string and
// appends the encoded value to the input byte slice.
func PutUint8(dst []byte, val uint8) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint16 converts the input uint16 to a string and
// appends the encoded value to the input byte slice.
func PutUint16(dst []byte, val uint16) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint32 converts the input uint32 to a string and
// appends the encoded value to the input byte slice.
func PutUint32(dst []byte, val uint32) []byte {
	return strconv.AppendUint(dst, uint64(val), 10)
}

// PutUint64 converts the input uint64 to a string and
// appends the encoded value to the input byte slice.
func PutUint64(dst []byte, val uint64) []byte {
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
// appends the encoded value to the input byte slice.
func PutFloat32(dst []byte, val float32) []byte {
	return appendFloat(dst, float64(val), 32)
}

// PutFloat64 converts the input float64 to a string and
// appends the encoded value to the input byte slice.
func PutFloat64(dst []byte, val float64) []byte {
	return appendFloat(dst, val, 64)
}

// PutAny marshals the input value to JSON and appends the result to dst.
//
// 注意：输出为原始 JSON 片段，不做 logfmt 的空格引号处理。若值序列化后
// 含空格/制表符（如 map 键含空格、struct 字段含空格），产出的 key={...}
// 片段会带裸空格，可能影响按空格切分字段的 logfmt 解析。
// 取舍：保留 JSON 原貌优先于 logfmt 严格性；调用方应避免对含空格的复杂值
// 使用 Any，必要时改用带引号的字符串字段。
func PutAny(dst []byte, i any) []byte {
	marshaled, err := json.Marshal(i)
	if err != nil {
		return PutStringQuote(dst, fmt.Sprintf("marshaling error: %v", err))
	}
	return append(dst, marshaled...)
}

// Thank for github.com/rs/zerolog
