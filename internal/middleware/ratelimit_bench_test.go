package middleware

import (
	"testing"
)

func BenchmarkRateLimiter_Allow(b *testing.B) {
	limiter := NewIPRateLimiter(1000000.0, 1000000)
	defer limiter.Close()

	ip := "192.0.2.1"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = limiter.Allow(ip)
	}
}

func BenchmarkRateLimiter_AllowParallel(b *testing.B) {
	limiter := NewIPRateLimiter(1000000.0, 1000000)
	defer limiter.Close()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ip := "192.0.2.5"
		for pb.Next() {
			_, _ = limiter.Allow(ip)
		}
	})
}
