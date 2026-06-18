package logs

import (
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/zxysilent/logs/internal/textenc"
)

const (
	callerBaseSkip = 2
	writeExtraSkip = 2
)

// lastSep 返回 file 中任一分隔符最靠后的匹配下标，无匹配返回 -1。
// 采用从右向左单次扫描 + 首字节剪枝：仅当当前字节等于某个分隔符首字节时才做整串比较，
// 命中即返回（最靠后）。相比对每个分隔符各做一次全串查找，在多个多字符分隔符
// （如 "/internal"、"/src"）场景下更快且零分配。
func lastSep(file string, seps []string) int {
	for i := len(file) - 1; i >= 0; i-- {
		c := file[i]
		for _, sep := range seps {
			if len(sep) == 0 || sep[0] != c {
				continue
			}
			if i+len(sep) <= len(file) && file[i:i+len(sep)] == sep {
				return i
			}
		}
	}
	return -1
}

func print(trace string, lv logLevel, caller bool, log *Logger, attr *buffer, args ...any) {
	buf := getb()
	defer putb(buf)
	*buf = textenc.PutBegin(*buf)
	*buf = textenc.PutTime(textenc.PutKeyRaw(*buf, timeFieldName), time.Now())
	*buf = textenc.PutString(textenc.PutKeyRaw(*buf, levelFieldName), lv.String())
	if trace != "" {
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, traceFieldName), trace)
	}
	if caller {
		_, file, line, ok := runtime.Caller(log.skip + callerBaseSkip)
		if !ok {
			file = "###"
			line = 0
		} else {
			slash := lastSep(file, log.sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, callerFieldName), file)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = textenc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	n := len(args)
	if n == 1 {
		key := textenc.PutKeyRaw(*buf, mesgFieldName)
		switch v := args[0].(type) {
		case string:
			*buf = textenc.PutStringQuote(key, v)
		case []byte:
			*buf = textenc.PutBytesQuote(key, v)
		case bool:
			*buf = textenc.PutBool(key, v)
		case int:
			*buf = textenc.PutInt(key, v)
		case int8:
			*buf = textenc.PutInt8(key, v)
		case int16:
			*buf = textenc.PutInt16(key, v)
		case int32:
			*buf = textenc.PutInt32(key, v)
		case int64:
			*buf = textenc.PutInt64(key, v)
		case uint:
			*buf = textenc.PutUint(key, v)
		case uint8:
			*buf = textenc.PutUint8(key, v)
		case uint16:
			*buf = textenc.PutUint16(key, v)
		case uint32:
			*buf = textenc.PutUint32(key, v)
		case uint64:
			*buf = textenc.PutUint64(key, v)
		case float32:
			*buf = textenc.PutFloat32(key, v)
		case float64:
			*buf = textenc.PutFloat64(key, v)
		case fmt.Stringer:
			*buf = textenc.PutStringQuote(key, v.String())
		default:
			*buf = textenc.PutStringQuote(key, fmt.Sprint(v))
		}
	} else if n > 1 {
		*buf = textenc.PutStringQuote(textenc.PutKeyRaw(*buf, mesgFieldName), fmt.Sprint(args...))
	}
	*buf = textenc.PutEnd(*buf)
	*buf = textenc.PutBreak(*buf)
	log.Write(*buf)
}

func printf(trace string, lv logLevel, caller bool, log *Logger, attr *buffer, format string, args ...any) {
	buf := getb()
	defer putb(buf)
	*buf = textenc.PutBegin(*buf)
	*buf = textenc.PutTime(textenc.PutKeyRaw(*buf, timeFieldName), time.Now())
	*buf = textenc.PutString(textenc.PutKeyRaw(*buf, levelFieldName), lv.String())
	if trace != "" {
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, traceFieldName), trace)
	}
	if caller {
		_, file, line, ok := runtime.Caller(log.skip + callerBaseSkip)
		if !ok {
			file = "###"
			line = 0
		} else {
			slash := lastSep(file, log.sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, callerFieldName), file)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = textenc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = textenc.PutStringQuote(textenc.PutKeyRaw(*buf, mesgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = textenc.PutStringQuote(textenc.PutKeyRaw(*buf, mesgFieldName), format)
	}
	*buf = textenc.PutEnd(*buf)
	*buf = textenc.PutBreak(*buf)
	log.Write(*buf)
}

func printb(trace string, lv logLevel, caller bool, log *Logger, attr *buffer, msg []byte) {
	buf := getb()
	defer putb(buf)
	*buf = textenc.PutBegin(*buf)
	*buf = textenc.PutTime(textenc.PutKeyRaw(*buf, timeFieldName), time.Now())
	*buf = textenc.PutString(textenc.PutKeyRaw(*buf, levelFieldName), lv.String())
	if trace != "" {
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, traceFieldName), trace)
	}
	if caller {
		_, file, line, ok := runtime.Caller(log.skip + callerBaseSkip + writeExtraSkip)
		if !ok {
			file = "###"
			line = 0
		} else {
			slash := lastSep(file, log.sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = textenc.PutString(textenc.PutKeyRaw(*buf, callerFieldName), file)
		*buf = append(*buf, ':')
		*buf = strconv.AppendInt(*buf, int64(line), 10)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = textenc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(msg) >= 1 {
		*buf = textenc.PutBytesQuote(textenc.PutKeyRaw(*buf, mesgFieldName), msg)
	}
	*buf = textenc.PutEnd(*buf)
	*buf = textenc.PutBreak(*buf)
	log.Write(*buf)
}

// buffer adapted from go/src/fmt/print.go
type buffer []byte

// Having an initial size gives a dramatic speedup.
var bpool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return (*buffer)(&b)
	},
}

func getb() *buffer {
	return bpool.Get().(*buffer)
}

const maxBufferSize = 512

func putb(b *buffer) {
	if b == nil {
		return
	}
	// To reduce peak allocation, return only smaller buffers to the pool.
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bpool.Put(b)
	}
}

var fpool = sync.Pool{New: func() any { return &fielder{} }}

func getfl() *fielder {
	return fpool.Get().(*fielder)
}

func putfl(fl *fielder) {
	if fl == nil {
		return
	}
	putb(fl.attr)
	fl.attr = nil
	fl.trace = ""
	fl.caller = false
	fl.skip = false
	fpool.Put(fl)
}
