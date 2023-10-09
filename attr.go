package logs

import (
	"fmt"
	"time"
)

func (s *FieldLogger) Str(key, val string) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutString(s.enc.PutKey(*s.attr, key), val)
	return s
}

func (s *FieldLogger) Stringer(key string, val fmt.Stringer) *FieldLogger {
	if s.attr == nil {
		return s
	}
	if val != nil {
		*s.attr = s.enc.PutString(s.enc.PutKey(*s.attr, key), val.String())
		return s
	}

	*s.attr = s.enc.PutAny(s.enc.PutKey(*s.attr, key), nil)
	return s
}

func (s *FieldLogger) Bytes(key string, val []byte) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutBytes(s.enc.PutKey(*s.attr, key), val)
	return s
}

func (s *FieldLogger) Err(err error) *FieldLogger {
	if s.attr == nil {
		return s
	}
	if err == nil {
		*s.attr = s.enc.PutNil(s.enc.PutKey(*s.attr, errorFieldName))
	} else {
		*s.attr = s.enc.PutString(s.enc.PutKey(*s.attr, errorFieldName), err.Error())
	}
	return s
}

func (s *FieldLogger) Bool(key string, b bool) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutBool(s.enc.PutKey(*s.attr, key), b)
	return s
}

func (s *FieldLogger) Int(key string, i int) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutInt(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int8(key string, i int8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutInt8(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int16(key string, i int16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutInt16(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int32(key string, i int32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutInt32(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int64(key string, i int64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutInt64(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint(key string, i uint) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutUint(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint8(key string, i uint8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutUint8(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint16(key string, i uint16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutUint16(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint32(key string, i uint32) *FieldLogger {
	*s.attr = s.enc.PutUint32(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint64(key string, i uint64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutUint64(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Float32(key string, f float32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutFloat32(s.enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Float64(key string, f float64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutFloat64(s.enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Time(key string, t time.Time) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutTime(s.enc.PutKey(*s.attr, key), t)
	// *s.attr = s.enc.PutTime(s.enc.PutKey(*s.attr, key), t, timeFieldFormat)
	return s
}

func (s *FieldLogger) Dur(key string, d time.Duration) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutDuration(s.enc.PutKey(*s.attr, key), d)
	return s
}

func (s *FieldLogger) Any(key string, i interface{}) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = s.enc.PutAny(s.enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Raw(key string, b []byte) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = append(s.enc.PutKey(*s.attr, key), b...)
	return s
}
