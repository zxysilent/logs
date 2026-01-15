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

// TraceCtx 处理traceid并返回新的context
// 1. ctx存在traceid，traceid参数不存在/为空 → 复用原有traceid
// 2. ctx不存在traceid → 使用新值（传入的traceid或生成新的）
// 3. ctx存在traceid且traceid参数存在 → 追加（原有值.新traceid）
func TraceCtx(ctx context.Context, traceid ...string) context.Context {
	ntraceid := ""
	if len(traceid) > 0 {
		ntraceid = traceid[0]
	}

	otraceid, _ := ctx.Value(traceKey).(string)

	ftraceid := ""
	if otraceid == "" {
		if ntraceid == "" {
			ntraceid = trace()
		}
		ftraceid = ntraceid
	} else {
		if ntraceid == "" {
			ftraceid = otraceid
		} else {
			ftraceid = otraceid + "." + ntraceid
		}
	}
	return context.WithValue(ctx, traceKey, ftraceid)
}
