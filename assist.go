package logs

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zxysilent/logs/internal/buffer"
)

func header(trace string, caller bool, log *Logger, buf *buffer.Buffer, lv logLevel) {
	*buf = log.enc.PutBegin(*buf)
	*buf = log.enc.PutTime(log.enc.PutKey(*buf, timeFieldName), time.Now())
	*buf = log.enc.PutString(log.enc.PutKey(*buf, levelFieldName), lv.String())
	if trace != "" {
		*buf = log.enc.PutString(log.enc.PutKey(*buf, traceFieldName), trace)
	}
	if caller {
		_, file, line, ok := runtime.Caller(log.skip + 3)
		if !ok {
			file = "###"
			line = 1
		} else {
			slash := strings.LastIndex(file, log.sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = log.enc.PutString(log.enc.PutKey(*buf, callerFieldName), file+":"+strconv.Itoa(line))
	}
}

func print(trace string, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, args ...any) {
	buf := buffer.Get()
	header(trace, caller, log, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = log.enc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = log.enc.PutString(log.enc.PutKey(*buf, msgFieldName), fmt.Sprint(args...))
	}
	*buf = log.enc.PutEnd(*buf)
	*buf = log.enc.PutBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)
}

func printf(trace string, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, format string, args ...any) {
	buf := buffer.Get()
	header(trace, caller, log, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = log.enc.PutDelim(*buf)
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = log.enc.PutString(log.enc.PutKey(*buf, msgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = log.enc.PutString(log.enc.PutKey(*buf, msgFieldName), format)
	}
	*buf = log.enc.PutEnd(*buf)
	*buf = log.enc.PutBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)

}
