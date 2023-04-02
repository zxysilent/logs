package logs

import "testing"

func TestTrace(t *testing.T) {
	t.Log(trace())
}
func BenchmarkTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		trace()
	}
}
