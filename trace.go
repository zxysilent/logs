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

// trace generates a random base-32 id of traceSize bytes.
func trace() string {
	buf := make([]byte, traceSize)
	for idx, cache := 0, fastrand(); idx < traceSize; {
		buf[idx] = traceStr[cache&traceMask]
		cache >>= 5
		idx++
	}
	return unsafe.String(&buf[0], len(buf))
}

// TraceId generates a new random trace id.
func TraceId() string {
	return trace()
}

// TraceOf returns the trace id stored in ctx, or empty string if none.
func TraceOf(ctx context.Context) string {
	traceId, _ := ctx.Value(traceKey).(string)
	return traceId
}

// fastrand is linked to the runtime's fast PRNG.
//
//go:linkname fastrand runtime.fastrand64
func fastrand() uint64

// ctxKey is the private context key type for storing the trace id.
type ctxKey struct{}

// traceKey is the context key used to store and retrieve the trace id.
var traceKey = ctxKey{}

// TraceCtx processes traceid and returns a new context.
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
