package logs

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zxysilent/logs/internal/buffer"
)

const tarceKey = "logs-tarce-id"

func TraceCtx(ctx context.Context, tarceid ...string) context.Context {
	val := ctx.Value(tarceKey)
	if val == nil {
		var id = ""
		if len(tarceid) == 0 {
			id = trace()
		} else {
			id = tarceid[0]
		}
		ctx = context.WithValue(ctx, tarceKey, id)
	}
	return ctx
}

func header(ctx context.Context, caller bool, log *Logger, buf *buffer.Buffer, lv logLevel) {
	*buf = log.enc.PutBegin(*buf)
	*buf = log.enc.PutTime(log.enc.PutKey(*buf, timeFieldName), time.Now())
	*buf = log.enc.PutString(log.enc.PutKey(*buf, levelFieldName), lv.String())
	if ctx != nil {
		val := ctx.Value(tarceKey)
		if val != nil {
			if traceId, ok := val.(string); ok {
				*buf = log.enc.PutString(log.enc.PutKey(*buf, traceFieldName), traceId)
			}
		}
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

func print(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log, buf, lv)
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

func printf(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, format string, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log, buf, lv)
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
