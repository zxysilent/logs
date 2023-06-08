package encoder

import (
	"time"

	"github.com/zxysilent/logs/internal/encoder/json"
	"github.com/zxysilent/logs/internal/encoder/text"
)

// Thank for github.com/rs/zerolog

var _ Encoder = (*json.Encoder)(nil)
var _ Encoder = (*text.Encoder)(nil)

type Encoder interface {
	PutBegin(dst []byte) []byte
	PutEnd(dst []byte) []byte
	PutBreak(dst []byte) []byte
	PutDelim(dst []byte) []byte
	PutBool(dst []byte, val bool) []byte
	PutBytes(dst, s []byte) []byte
	PutDuration(dst []byte, d time.Duration) []byte
	PutFloat32(dst []byte, val float32) []byte
	PutFloat64(dst []byte, val float64) []byte
	PutInt(dst []byte, val int) []byte
	PutInt16(dst []byte, val int16) []byte
	PutInt32(dst []byte, val int32) []byte
	PutInt64(dst []byte, val int64) []byte
	PutInt8(dst []byte, val int8) []byte
	PutAny(dst []byte, i interface{}) []byte
	PutKey(dst []byte, key string) []byte
	PutNil(dst []byte) []byte
	PutString(dst []byte, s string) []byte
	PutTime(dst []byte, t time.Time) []byte
	PutUint(dst []byte, val uint) []byte
	PutUint16(dst []byte, val uint16) []byte
	PutUint32(dst []byte, val uint32) []byte
	PutUint64(dst []byte, val uint64) []byte
	PutUint8(dst []byte, val uint8) []byte
}
