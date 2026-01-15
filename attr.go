package logs

import (
	"fmt"
	"time"

	"github.com/zxysilent/logs/internal/textenc"
)

func (s *fieldLogger) Str(key, val string) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutString(textenc.PutKey(*s.attr, key), val)
	return s
}

func (s *fieldLogger) Stringer(key string, val fmt.Stringer) *fieldLogger {
	if s.attr == nil {
		return s
	}
	if val != nil {
		*s.attr = textenc.PutString(textenc.PutKey(*s.attr, key), val.String())
		return s
	}

	*s.attr = textenc.PutAny(textenc.PutKey(*s.attr, key), nil)
	return s
}

func (s *fieldLogger) Bytes(key string, val []byte) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBytes(textenc.PutKey(*s.attr, key), val)
	return s
}

func (s *fieldLogger) Err(err error) *fieldLogger {
	if s.attr == nil {
		return s
	}
	if err == nil {
		*s.attr = textenc.PutNil(textenc.PutKey(*s.attr, errorFieldName))
	} else {
		*s.attr = textenc.PutString(textenc.PutKey(*s.attr, errorFieldName), err.Error())
	}
	return s
}

func (s *fieldLogger) IfErr(err error) *fieldLogger {
	if err == nil {
		s.skip = true
		return s
	}
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutString(textenc.PutKey(*s.attr, errorFieldName), err.Error())
	return s
}

func (s *fieldLogger) If(b bool) *fieldLogger {
	s.skip = !b
	return s
}

func (s *fieldLogger) Bool(key string, b bool) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBool(textenc.PutKey(*s.attr, key), b)
	return s
}

func (s *fieldLogger) Int(key string, i int) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Int8(key string, i int8) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt8(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Int16(key string, i int16) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt16(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Int32(key string, i int32) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt32(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Int64(key string, i int64) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt64(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Uint(key string, i uint) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Uint8(key string, i uint8) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint8(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Uint16(key string, i uint16) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint16(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Uint32(key string, i uint32) *fieldLogger {
	*s.attr = textenc.PutUint32(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Uint64(key string, i uint64) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint64(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Float32(key string, f float32) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat32(textenc.PutKey(*s.attr, key), f)
	return s
}

func (s *fieldLogger) Float64(key string, f float64) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat64(textenc.PutKey(*s.attr, key), f)
	return s
}

func (s *fieldLogger) Time(key string, t time.Time) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutTime(textenc.PutKey(*s.attr, key), t)
	// *s.attr = textenc.PutTime(textenc.PutKey(*s.attr, key), t, timeFieldFormat)
	return s
}

func (s *fieldLogger) Dur(key string, d time.Duration) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutDuration(textenc.PutKey(*s.attr, key), d)
	return s
}

func (s *fieldLogger) Any(key string, i any) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutAny(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fieldLogger) Raw(key string, b []byte) *fieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = append(textenc.PutKey(*s.attr, key), b...)
	return s
}
