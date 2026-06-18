package logs

import (
	"fmt"
	"time"

	"github.com/zxysilent/logs/internal/textenc"
)

func (s *fielder) Str(key, val string) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutStringQuote(textenc.PutKey(*s.attr, key), val)
	return s
}

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

func (s *fielder) Bytes(key string, val []byte) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBytesQuote(textenc.PutKey(*s.attr, key), val)
	return s
}

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

func (s *fielder) If(b bool) *fielder {
	s.skip = !b
	return s
}

func (s *fielder) Bool(key string, b bool) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutBool(textenc.PutKey(*s.attr, key), b)
	return s
}

func (s *fielder) Int(key string, i int) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Int8(key string, i int8) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt8(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Int16(key string, i int16) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt16(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Int32(key string, i int32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt32(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Int64(key string, i int64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutInt64(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Uint(key string, i uint) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Uint8(key string, i uint8) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint8(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Uint16(key string, i uint16) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint16(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Uint32(key string, i uint32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint32(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Uint64(key string, i uint64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutUint64(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Float32(key string, f float32) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat32(textenc.PutKey(*s.attr, key), f)
	return s
}

func (s *fielder) Float64(key string, f float64) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutFloat64(textenc.PutKey(*s.attr, key), f)
	return s
}

func (s *fielder) Time(key string, t time.Time) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutTime(textenc.PutKey(*s.attr, key), t)
	return s
}

func (s *fielder) Dur(key string, d time.Duration) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutDuration(textenc.PutKey(*s.attr, key), d)
	return s
}

func (s *fielder) Any(key string, i any) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = textenc.PutAny(textenc.PutKey(*s.attr, key), i)
	return s
}

func (s *fielder) Raw(key string, b []byte) *fielder {
	if s.attr == nil {
		return s
	}
	*s.attr = append(textenc.PutKey(*s.attr, key), b...)
	return s
}
