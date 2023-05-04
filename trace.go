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
	for idx, cache := 0, fastRand(); idx < 8; idx++ {
		buf[idx] = traceStr[cache&traceMask]
		cache >>= 4

	}
	return *(*string)(unsafe.Pointer(&buf))
	// return unsafe.String(&buf[0], len(buf)) //1.20
}

func TraceId() string {
	return trace()
}

//go:linkname fastRand runtime.fastrand
func fastRand() uint32
