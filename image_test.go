package criprof

import (
	"testing"
)

func BenchmarkGetImageFormat(b *testing.B) {
	// Run getImageFormat function b.N times.
	for i := 0; i < b.N; i++ {
		getImageFormat()
	}
}
