package logs

import (
	"context"
	"unsafe"
)

// go clean -testcache // Delete all cached test results

const (
	traceStr  = "23456789abcdefghijkmnpqrstuvwxyz" //32
	traceMask = 1<<5 - 1                           //11111
	traceSize = 8                                  //6-12
)

func trace() string {
	buf := make([]byte, traceSize)
	for idx, cache := 0, fastrand(); idx < traceSize; {
		buf[idx] = traceStr[cache&traceMask]
		cache >>= 5
		idx++
	}
	return unsafe.String(&buf[0], len(buf))
}

func TraceId() string {
	return trace()
}

func TraceOf(ctx context.Context) string {
	traceId, _ := ctx.Value(traceKey).(string)
	return traceId
}

//go:linkname fastrand runtime.fastrand64
func fastrand() uint64

const traceKey = "zlogs-trace-key"

func TraceCtx(ctx context.Context, tarceid ...string) context.Context {
	val := ctx.Value(traceKey)
	if val == nil {
		var id = ""
		if len(tarceid) == 0 {
			id = trace()
		} else {
			id = tarceid[0]
		}
		ctx = context.WithValue(ctx, traceKey, id)
	}
	return ctx
}
