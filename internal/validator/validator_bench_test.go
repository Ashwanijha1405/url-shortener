package validator

import "testing"

func BenchmarkValidateURL_Valid(b *testing.B) {
	url := "https://github.com/Ashwanijha1405/url-shortener/tree/main"
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ValidateURL(url)
	}
}

func BenchmarkValidateURL_SSRFBlocked(b *testing.B) {
	url := "http://169.254.169.254/latest/meta-data/"
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ValidateURL(url)
	}
}
