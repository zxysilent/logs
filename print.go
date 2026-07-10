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
	callerSkip = 4 // putCaller → print/printf → (method) → caller
	writerSkip = 7
)

// lastSep returns the last index where any separator matches in file, or -1 if none.
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

// putCaller writes "caller=file:line" into buf using runtime.Callers + FuncForPC.FileLine
// (no allocation: pcs does not escape and FileLine does not allocate).
// skip is the full frame count passed to runtime.Callers (computed by the caller).
func (c *config) putCaller(buf *buffer, skip int) {
	var pcs [1]uintptr
	file := "###"
	line := 0
	if runtime.Callers(skip, pcs[:]) >= 1 {
		// PC-1 points into the CALL instruction (CallersFrames semantics).
		if fn := runtime.FuncForPC(pcs[0] - 1); fn != nil {
			file, line = fn.FileLine(pcs[0] - 1)
			if slash := lastSep(file, c.sep); slash >= 0 {
				file = file[slash:]
			}
		}
	}
	*buf = append(*buf, " caller="...)
	*buf = textenc.PutString(*buf, file)
	*buf = append(*buf, ':')
	*buf = strconv.AppendInt(*buf, int64(line), 10)
}

// putLevel writes " level=XXX" — buf is always non-empty at this point.
func putLevel(dst *buffer, lv Level) {
	switch lv {
	case LevelDebug:
		*dst = append(*dst, " level=DBG"...)
	case LevelInfo:
		*dst = append(*dst, " level=INF"...)
	case LevelWarn:
		*dst = append(*dst, " level=WRN"...)
	case LevelError:
		*dst = append(*dst, " level=ERR"...)
	default:
		*dst = append(*dst, " level=OFF"...)
	}
}

// print writes a log record.
func (c *config) print(trace string, lv Level, caller bool, attr *buffer, args ...any) {
	buf := getb()
	*buf = append(*buf, "time="...)
	*buf = textenc.PutTime(*buf, time.Now())
	putLevel(buf, lv)
	if trace != "" {
		*buf = append(*buf, " trace="...)
		*buf = textenc.PutString(*buf, trace)
	}
	if caller {
		c.putCaller(buf, c.skip+callerSkip)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, *attr...)
	}
	n := len(args)
	if n == 1 {
		*buf = append(*buf, " msg="...)
		switch v := args[0].(type) {
		case string:
			*buf = textenc.PutStringQuote(*buf, v)
		case []byte:
			*buf = textenc.PutBytesQuote(*buf, v)
		case bool:
			*buf = textenc.PutBool(*buf, v)
		case int:
			*buf = textenc.PutInt(*buf, v)
		case int8:
			*buf = textenc.PutInt8(*buf, v)
		case int16:
			*buf = textenc.PutInt16(*buf, v)
		case int32:
			*buf = textenc.PutInt32(*buf, v)
		case int64:
			*buf = textenc.PutInt64(*buf, v)
		case uint:
			*buf = textenc.PutUint(*buf, v)
		case uint8:
			*buf = textenc.PutUint8(*buf, v)
		case uint16:
			*buf = textenc.PutUint16(*buf, v)
		case uint32:
			*buf = textenc.PutUint32(*buf, v)
		case uint64:
			*buf = textenc.PutUint64(*buf, v)
		case float32:
			*buf = textenc.PutFloat32(*buf, v)
		case float64:
			*buf = textenc.PutFloat64(*buf, v)
		case fmt.Stringer:
			*buf = textenc.PutStringQuote(*buf, v.String())
		default:
			*buf = textenc.PutStringQuote(*buf, fmt.Sprint(v))
		}
	} else if n > 1 {
		*buf = append(*buf, " msg="...)
		*buf = textenc.PutStringQuote(*buf, fmt.Sprint(args...))
	}
	*buf = append(*buf, '\n')
	c.out.Write(*buf)
	putb(buf)
}

// printf writes a formatted log record.
func (c *config) printf(trace string, lv Level, caller bool, attr *buffer, format string, args ...any) {
	buf := getb()
	*buf = append(*buf, "time="...)
	*buf = textenc.PutTime(*buf, time.Now())
	putLevel(buf, lv)
	if trace != "" {
		*buf = append(*buf, " trace="...)
		*buf = textenc.PutString(*buf, trace)
	}
	if caller {
		c.putCaller(buf, c.skip+callerSkip)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, *attr...)
	}
	*buf = append(*buf, " msg="...)
	if len(args) >= 1 {
		*buf = textenc.PutStringQuote(*buf, fmt.Sprintf(format, args...))
	} else {
		*buf = textenc.PutStringQuote(*buf, format)
	}
	*buf = append(*buf, '\n')
	c.out.Write(*buf)
	putb(buf)
}

// printb writes a log record with a byte slice message.
func (c *config) printb(trace string, lv Level, caller bool, attr *buffer, msg []byte) {
	buf := getb()
	*buf = append(*buf, "time="...)
	*buf = textenc.PutTime(*buf, time.Now())
	putLevel(buf, lv)
	if trace != "" {
		*buf = append(*buf, " trace="...)
		*buf = textenc.PutString(*buf, trace)
	}
	if caller {
		c.putCaller(buf, c.skip+writerSkip)
	}
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ' ')
		*buf = append(*buf, *attr...)
	}
	if len(msg) >= 1 {
		*buf = append(*buf, " msg="...)
		*buf = textenc.PutBytesQuote(*buf, msg)
	}
	*buf = append(*buf, '\n')
	c.out.Write(*buf)
	putb(buf)
}

const maxBufferSize = 512

// buffer adapted from go/src/fmt/print.go
type buffer []byte

// bpool reuses buffers across log calls. An initial size gives a dramatic speedup.
var bpool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, maxBufferSize)
		return (*buffer)(&b)
	},
}

// getb gets a pooled buffer.
func getb() *buffer {
	return bpool.Get().(*buffer)
}

// putb returns a buffer to the pool, dropping oversized ones to bound peak memory.
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

// fpool reuses fielder values to avoid per-call allocation.
var fpool = sync.Pool{New: func() any { return &fielder{} }}

// getfl gets a pooled fielder.
func getfl() *fielder {
	return fpool.Get().(*fielder)
}

// putfl resets and returns a fielder to the pool.
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
