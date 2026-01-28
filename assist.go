package logs

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zxysilent/logs/internal/textenc"
)

func header(trace string, caller bool, log *Logger, buf *buffer, lv logLevel) {
	*buf = textenc.PutBegin(*buf)
	*buf = textenc.PutTime(textenc.PutKey(*buf, timeFieldName), time.Now())
	*buf = textenc.PutString(textenc.PutKey(*buf, levelFieldName), lv.String())
	if trace != "" {
		*buf = textenc.PutString(textenc.PutKey(*buf, traceFieldName), trace)
	}
	if caller {
		_, file, line, ok := runtime.Caller(log.skip + 3)
		if !ok {
			file = "###"
			line = 0
		} else {
			slash := strings.LastIndex(file, log.sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = textenc.PutString(textenc.PutKey(*buf, callerFieldName), file+":"+strconv.Itoa(line))
	}
}

func print(trace string, lv logLevel, caller bool, log *Logger, attr *buffer, args ...any) {
	buf := getb()
	defer putb(buf)
	header(trace, caller, log, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = textenc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = textenc.PutString(textenc.PutKey(*buf, msgFieldName), fmt.Sprint(args...))
	}
	*buf = textenc.PutEnd(*buf)
	*buf = textenc.PutBreak(*buf)
	log.Write(*buf)
}

func printf(trace string, lv logLevel, caller bool, log *Logger, attr *buffer, format string, args ...any) {
	buf := getb()
	defer putb(buf)
	header(trace, caller, log, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = textenc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = textenc.PutString(textenc.PutKey(*buf, msgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = textenc.PutString(textenc.PutKey(*buf, msgFieldName), format)
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

const maxBufferSize = 2 << 8

func putb(b *buffer) {
	if b == nil {
		return
	}
	// To reduce peak allocation, return only smaller buffers to the pool.
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		bpool.Put(b)
		return
	}
	b = nil
}

func (b *buffer) Reset() {
	*b = (*b)[:0]
}

func (b *buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(p), nil
}

// func (b *buffer) WriteString(s string) {
// 	*b = append(*b, s...)
// }

// func (b *buffer) WriteByte(c byte) error {
// 	*b = append(*b, c)
// 	return nil
// }

// func (b *buffer) String() string {
// 	return string(*b)
// }

var fpool = sync.Pool{New: func() any { return &fieldLogger{} }}

func getfl() *fieldLogger {
	return fpool.Get().(*fieldLogger)
}

func putfl(fl *fieldLogger) {
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
