package criprof

import "testing"

func BenchmarkGetRuntime(b *testing.B) {
	// Run getRuntime function b.N times.
	for i := 0; i < b.N; i++ {
		getRuntime()
	}
}
