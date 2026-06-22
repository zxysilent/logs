package logs

import (
	"fmt"
	"time"

	"github.com/zxysilent/logs/internal/textenc"
)

// Str adds a string field.
func (s *fielder) Str(key, val string) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutStringQuote(textenc.PutKey(*s.attr, key), val)
	return s
}

// Stringer adds a field from a fmt.Stringer (nil renders as null).
func (s *fielder) Stringer(key string, val fmt.Stringer) *fielder {
	if s.attr == nil {
		return s
	}
	if val != nil {
		*s.attr = textenc.PutStringQuote(textenc.PutKey(*s.attr, key), val.String())
		return s
	}

	*s.attr = textenc.PutAny(textenc.PutKey(*s.attr, key), nil)
	return s
}

// Bytes adds a byte-slice field.
func (s *fielder) Bytes(key string, val []byte) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBytesQuote(textenc.PutKey(*s.attr, key), val)
	return s
}

// Err adds an error field (nil renders as null).
func (s *fielder) Err(err error) *fielder {
	if s.attr == nil {
		return s
	}
	if err == nil {
		*s.attr = textenc.PutNil(textenc.PutKey(*s.attr, errorFieldName))
	} else {
		*s.attr = textenc.PutStringQuote(textenc.PutKey(*s.attr, errorFieldName), err.Error())
	}
	return s
}

// IfErr adds an error field only if err is non-nil; otherwise the log is skipped.
func (s *fielder) IfErr(err error) *fielder {
	if err == nil {
		s.skip = true
		return s
	}
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutStringQuote(textenc.PutKey(*s.attr, errorFieldName), err.Error())
	return s
}

// If skips the log unless b is true.
func (s *fielder) If(b bool) *fielder {
	s.skip = !b
	return s
}

// Bool adds a bool field.
func (s *fielder) Bool(key string, b bool) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBool(textenc.PutKey(*s.attr, key), b)
	return s
}

// Int adds an int field.
func (s *fielder) Int(key string, i int) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt(textenc.PutKey(*s.attr, key), i)
	return s
}

// Int8 adds an int8 field.
func (s *fielder) Int8(key string, i int8) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt8(textenc.PutKey(*s.attr, key), i)
	return s
}

// Int16 adds an int16 field.
func (s *fielder) Int16(key string, i int16) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt16(textenc.PutKey(*s.attr, key), i)
	return s
}

// Int32 adds an int32 field.
func (s *fielder) Int32(key string, i int32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt32(textenc.PutKey(*s.attr, key), i)
	return s
}

// Int64 adds an int64 field.
func (s *fielder) Int64(key string, i int64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt64(textenc.PutKey(*s.attr, key), i)
	return s
}

// Uint adds a uint field.
func (s *fielder) Uint(key string, i uint) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint(textenc.PutKey(*s.attr, key), i)
	return s
}

// Uint8 adds a uint8 field.
func (s *fielder) Uint8(key string, i uint8) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint8(textenc.PutKey(*s.attr, key), i)
	return s
}

// Uint16 adds a uint16 field.
func (s *fielder) Uint16(key string, i uint16) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint16(textenc.PutKey(*s.attr, key), i)
	return s
}

// Uint32 adds a uint32 field.
func (s *fielder) Uint32(key string, i uint32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint32(textenc.PutKey(*s.attr, key), i)
	return s
}

// Uint64 adds a uint64 field.
func (s *fielder) Uint64(key string, i uint64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint64(textenc.PutKey(*s.attr, key), i)
	return s
}

// Float32 adds a float32 field.
func (s *fielder) Float32(key string, f float32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat32(textenc.PutKey(*s.attr, key), f)
	return s
}

// Float64 adds a float64 field.
func (s *fielder) Float64(key string, f float64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat64(textenc.PutKey(*s.attr, key), f)
	return s
}

// Time adds a time.Time field.
func (s *fielder) Time(key string, t time.Time) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutTime(textenc.PutKey(*s.attr, key), t)
	return s
}

// Dur adds a time.Duration field.
func (s *fielder) Dur(key string, d time.Duration) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutDuration(textenc.PutKey(*s.attr, key), d)
	return s
}

// Any adds an arbitrary value as a JSON-marshaled field.
func (s *fielder) Any(key string, i any) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutAny(textenc.PutKey(*s.attr, key), i)
	return s
}

// Raw adds a raw byte field without quoting or escaping.
func (s *fielder) Raw(key string, b []byte) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = append(textenc.PutKey(*s.attr, key), b...)
	return s
}
