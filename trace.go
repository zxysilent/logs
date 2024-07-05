package logs

import (
	"unsafe"
)

// go clean -testcache // Delete all cached test results

const (
	traceStr  = "23456789abcdefghijkmnpqrstuvwxyz" //32
	traceMask = 1<<5 - 1                           //11111
	traceSize = 8                                  //6-16
)

func trace() string {
	buf := make([]byte, traceSize)
	for idx, cache := 0, fastRand(); idx < traceSize; {
		buf[idx] = traceStr[cache&traceMask]
		cache >>= 4 //hacker
		idx++
	}
	return unsafe.String(&buf[0], len(buf))
}
func TraceId() string {
	return trace()
}

//go:linkname fastRand runtime.fastrand64
func fastRand() uint64
