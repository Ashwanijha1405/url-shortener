package generator

import "testing"

func BenchmarkGenerate(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Generate(DefaultLength)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

func BenchmarkGenerateCustomLength(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Generate(12)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}
