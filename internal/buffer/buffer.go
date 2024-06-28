package buffer

import (
	"sync"
)

// buffer adapted from go/src/fmt/print.go
type Buffer []byte

// Having an initial size gives a dramatic speedup.
var pool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 512)
		return (*Buffer)(&b)
	},
}

func Get() *Buffer {
	return pool.Get().(*Buffer)
}

const maxBufferSize = 4 << 10

func Put(b *Buffer) {
	if b == nil {
		return
	}
	// To reduce peak allocation, return only smaller buffers to the pool.
	if cap(*b) <= maxBufferSize {
		*b = (*b)[:0]
		pool.Put(b)
		return
	}
	b = nil
}

func (b *Buffer) Reset() {
	*b = (*b)[:0]
}

func (b *Buffer) Write(p []byte) (int, error) {
	*b = append(*b, p...)
	return len(p), nil
}

func (b *Buffer) WriteString(s string) {
	*b = append(*b, s...)
}

func (b *Buffer) WriteByte(c byte) error {
	*b = append(*b, c)
	return nil
}

func (b *Buffer) String() string {
	return string(*b)
}
