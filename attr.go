package logs

import (
	"encoding/json"
	"fmt"
	"time"
	"unsafe"
)

func isNilValue(i interface{}) bool {
	return (*[2]uintptr)(unsafe.Pointer(&i))[1] == 0
}
func appendJSON(dst []byte, j []byte) []byte {
	return append(dst, j...)
}

func (s *FieldLogger) Str(key, val string) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutString(enc.PutKey(*s.attr, key), val)
	return s
}

func (s *FieldLogger) Strs(key string, vals []string) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutStrings(enc.PutKey(*s.attr, key), vals)
	return s
}

func (s *FieldLogger) Stringer(key string, val fmt.Stringer) *FieldLogger {
	if s.attr == nil {
		return s
	}
	if val != nil {
		*s.attr = enc.PutString(enc.PutKey(*s.attr, key), val.String())
		return s
	}

	*s.attr = enc.PutAny(enc.PutKey(*s.attr, key), nil)
	return s
}

func (s *FieldLogger) Bytes(key string, val []byte) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutBytes(enc.PutKey(*s.attr, key), val)
	return s
}

func (s *FieldLogger) Hex(key string, val []byte) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutHex(enc.PutKey(*s.attr, key), val)
	return s
}

func (s *FieldLogger) Err(err error) *FieldLogger {
	if s.attr == nil {
		return s
	}
	if err == nil {
		*s.attr = enc.PutNil(enc.PutKey(*s.attr, ErrorFieldName))

	} else {
		*s.attr = enc.PutString(enc.PutKey(*s.attr, ErrorFieldName), err.Error())
	}
	return s
}

func (s *FieldLogger) Bool(key string, b bool) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutBool(enc.PutKey(*s.attr, key), b)
	return s
}

func (s *FieldLogger) Bools(key string, b []bool) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutBools(enc.PutKey(*s.attr, key), b)
	return s
}

func (s *FieldLogger) Int(key string, i int) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInt(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Ints(key string, i []int) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInts(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int8(key string, i int8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInt8(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Ints8(key string, i []int8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInts8(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int16(key string, i int16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInt16(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Ints16(key string, i []int16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInts16(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int32(key string, i int32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInt32(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Ints32(key string, i []int32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInts32(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Int64(key string, i int64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInt64(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Ints64(key string, i []int64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutInts64(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint(key string, i uint) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUint(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uints(key string, i []uint) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUints(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint8(key string, i uint8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUint8(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uints8(key string, i []uint8) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUints8(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint16(key string, i uint16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUint16(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uints16(key string, i []uint16) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUints16(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint32(key string, i uint32) *FieldLogger {
	*s.attr = enc.PutUint32(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uints32(key string, i []uint32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUints32(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uint64(key string, i uint64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUint64(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Uints64(key string, i []uint64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutUints64(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) Float32(key string, f float32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutFloat32(enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Floats32(key string, f []float32) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutFloats32(enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Float64(key string, f float64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutFloat64(enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Floats64(key string, f []float64) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutFloats64(enc.PutKey(*s.attr, key), f)
	return s
}

func (s *FieldLogger) Time(key string, t time.Time) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutTime(enc.PutKey(*s.attr, key), t, TimeFieldFormat)
	return s
}

func (s *FieldLogger) Times(key string, t []time.Time) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutTimes(enc.PutKey(*s.attr, key), t, TimeFieldFormat)
	return s
}

func (s *FieldLogger) Dur(key string, d time.Duration) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutDuration(enc.PutKey(*s.attr, key), d, DurationFieldUnit, DurationFieldInteger)
	return s
}

func (s *FieldLogger) Durs(key string, d []time.Duration) *FieldLogger {
	if s.attr == nil {
		return s
	}
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutDurations(enc.PutKey(*s.attr, key), d, DurationFieldUnit, DurationFieldInteger)
	return s
}

func (s *FieldLogger) Any(key string, i interface{}) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutAny(enc.PutKey(*s.attr, key), i)
	return s
}

func (s *FieldLogger) RawJSON(key string, b []byte) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = appendJSON(enc.PutKey(*s.attr, key), b)
	return s
}

func (s *FieldLogger) Auto(key string, val interface{}) *FieldLogger {
	if s.attr == nil {
		return s
	}
	*s.attr = enc.PutKey(*s.attr, key)
	switch val := val.(type) {
	case string:
		*s.attr = enc.PutString(*s.attr, val)
	case []byte:
		*s.attr = enc.PutBytes(*s.attr, val)
	case error:
		if val == nil {
			*s.attr = enc.PutNil(*s.attr)
		} else {
			*s.attr = enc.PutString(*s.attr, val.Error())
		}
	case []error:
		*s.attr = enc.PutArrayStart(*s.attr)
		for i, err := range val {
			if err == nil {
				*s.attr = enc.PutNil(*s.attr)
			} else {
				*s.attr = enc.PutString(*s.attr, err.Error())
			}

			if i < (len(val) - 1) {
				enc.PutArrayDelim(*s.attr)
			}
		}
		*s.attr = enc.PutArrayEnd(*s.attr)
	case bool:
		*s.attr = enc.PutBool(*s.attr, val)
	case int:
		*s.attr = enc.PutInt(*s.attr, val)
	case int8:
		*s.attr = enc.PutInt8(*s.attr, val)
	case int16:
		*s.attr = enc.PutInt16(*s.attr, val)
	case int32:
		*s.attr = enc.PutInt32(*s.attr, val)
	case int64:
		*s.attr = enc.PutInt64(*s.attr, val)
	case uint:
		*s.attr = enc.PutUint(*s.attr, val)
	case uint8:
		*s.attr = enc.PutUint8(*s.attr, val)
	case uint16:
		*s.attr = enc.PutUint16(*s.attr, val)
	case uint32:
		*s.attr = enc.PutUint32(*s.attr, val)
	case uint64:
		*s.attr = enc.PutUint64(*s.attr, val)
	case float32:
		*s.attr = enc.PutFloat32(*s.attr, val)
	case float64:
		*s.attr = enc.PutFloat64(*s.attr, val)
	case time.Time:
		*s.attr = enc.PutTime(*s.attr, val, TimeFieldFormat)
	case time.Duration:
		*s.attr = enc.PutDuration(*s.attr, val, DurationFieldUnit, DurationFieldInteger)
	case *string:
		if val != nil {
			*s.attr = enc.PutString(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *bool:
		if val != nil {
			*s.attr = enc.PutBool(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *int:
		if val != nil {
			*s.attr = enc.PutInt(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *int8:
		if val != nil {
			*s.attr = enc.PutInt8(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *int16:
		if val != nil {
			*s.attr = enc.PutInt16(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *int32:
		if val != nil {
			*s.attr = enc.PutInt32(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *int64:
		if val != nil {
			*s.attr = enc.PutInt64(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *uint:
		if val != nil {
			*s.attr = enc.PutUint(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *uint8:
		if val != nil {
			*s.attr = enc.PutUint8(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *uint16:
		if val != nil {
			*s.attr = enc.PutUint16(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *uint32:
		if val != nil {
			*s.attr = enc.PutUint32(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *uint64:
		if val != nil {
			*s.attr = enc.PutUint64(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *float32:
		if val != nil {
			*s.attr = enc.PutFloat32(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *float64:
		if val != nil {
			*s.attr = enc.PutFloat64(*s.attr, *val)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *time.Time:
		if val != nil {
			*s.attr = enc.PutTime(*s.attr, *val, TimeFieldFormat)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case *time.Duration:
		if val != nil {
			*s.attr = enc.PutDuration(*s.attr, *val, DurationFieldUnit, DurationFieldInteger)
		} else {
			*s.attr = enc.PutNil(*s.attr)
		}
	case []string:
		*s.attr = enc.PutStrings(*s.attr, val)
	case []bool:
		*s.attr = enc.PutBools(*s.attr, val)
	case []int:
		*s.attr = enc.PutInts(*s.attr, val)
	case []int8:
		*s.attr = enc.PutInts8(*s.attr, val)
	case []int16:
		*s.attr = enc.PutInts16(*s.attr, val)
	case []int32:
		*s.attr = enc.PutInts32(*s.attr, val)
	case []int64:
		*s.attr = enc.PutInts64(*s.attr, val)
	case []uint:
		*s.attr = enc.PutUints(*s.attr, val)
	// case []uint8:
	// 	*s.attr = enc.PutUints8(*s.attr, val)
	case []uint16:
		*s.attr = enc.PutUints16(*s.attr, val)
	case []uint32:
		*s.attr = enc.PutUints32(*s.attr, val)
	case []uint64:
		*s.attr = enc.PutUints64(*s.attr, val)
	case []float32:
		*s.attr = enc.PutFloats32(*s.attr, val)
	case []float64:
		*s.attr = enc.PutFloats64(*s.attr, val)
	case []time.Time:
		*s.attr = enc.PutTimes(*s.attr, val, TimeFieldFormat)
	case []time.Duration:
		*s.attr = enc.PutDurations(*s.attr, val, DurationFieldUnit, DurationFieldInteger)
	case nil:
		*s.attr = enc.PutNil(*s.attr)
	case json.RawMessage:
		*s.attr = appendJSON(*s.attr, val)
	default:
		*s.attr = enc.PutAny(*s.attr, val)
	}
	return s
}
