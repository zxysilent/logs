package logs

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/zxysilent/logs/internal/buffer"
	"github.com/zxysilent/logs/internal/encoder"
)

var (
	enc = encoder.Encoder{}
)

const trackKey = "logs-track-id"

func TrackCtx(ctx context.Context, trackid ...string) context.Context {
	val := ctx.Value(trackKey)
	if val == nil {
		var id = ""
		if len(trackid) == 0 {
			id = trace()
		} else {
			id = trackid[0]
		}
		ctx = context.WithValue(ctx, trackKey, id)
	}
	return ctx
}

func header(ctx context.Context, caller bool, skip int, sep string, buf *buffer.Buffer, lv logLevel) {
	*buf = enc.PutBeginMarker(*buf)
	*buf = enc.PutTimeFast(enc.PutKey(*buf, timeFieldName), time.Now())
	*buf = enc.PutString(enc.PutKey(*buf, levelFieldName), lv.String())
	if ctx != nil {
		val := ctx.Value(trackKey)
		if val != nil {
			if traceId, ok := val.(string); ok {
				*buf = enc.PutString(enc.PutKey(*buf, traceFieldName), traceId)
			}
		}
	}
	if caller {
		_, file, line, ok := runtime.Caller(skip + 3)
		if !ok {
			file = "###"
			line = 1
		} else {
			slash := strings.LastIndex(file, sep)
			if slash >= 0 {
				file = file[slash:]
			}
		}
		*buf = enc.PutString(enc.PutKey(*buf, callerFieldName), file+":"+strconv.Itoa(line))
	}
}

func print(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log.skip, log.sep, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ',')
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprint(args...))
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)
}

func printf(ctx context.Context, lv logLevel, caller bool, log *Logger, attr *buffer.Buffer, format string, args ...interface{}) {
	buf := buffer.Get()
	header(ctx, caller, log.skip, log.sep, buf, lv)
	if attr != nil && len(*attr) >= 1 {
		*buf = append(*buf, ',')
		*buf = append(*buf, *attr...)
	}
	if len(args) >= 1 {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), fmt.Sprintf(format, args...))
	} else {
		*buf = enc.PutString(enc.PutKey(*buf, msgFieldName), format)
	}
	*buf = enc.PutEndMarker(*buf)
	*buf = enc.PutLineBreak(*buf)
	log.Write(*buf)
	buffer.Put(buf)

}
