package logs

import (
	"unsafe"
)

// go clean -testcache // Delete all cached test results

const (
	traceStr  = "0123456789abcdef" //16
	traceMask = 1<<4 - 1           //1111
)

// xxxxxxxx
func trace() string {
	buf := make([]byte, 8)
	for idx, cache, remain := 0, fastRand(), 8; idx < 8; {
		buf[idx] = traceStr[cache&traceMask]
		cache >>= 4
		remain--
		idx++
	}
	return *(*string)(unsafe.Pointer(&buf))
}

func TraceId() string {
	return trace()
}

//go:linkname fastRand runtime.fastrand
func fastRand() uint32
