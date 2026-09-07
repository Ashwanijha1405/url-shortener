package validator

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		// Valid URLs
		{
			name:    "valid https URL",
			url:     "https://github.com",
			wantErr: false,
		},
		{
			name:    "valid http URL with path and query",
			url:     "http://example.com/search?q=test#heading",
			wantErr: false,
		},
		{
			name:    "valid https with custom port",
			url:     "https://api.example.com:8443/v1/resource",
			wantErr: false,
		},

		// Basic Validation Errors
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "unsupported scheme ftp",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "unsupported scheme file",
			url:     "file:///etc/passwd",
			wantErr: true,
		},
		{
			name:    "missing host",
			url:     "https://",
			wantErr: true,
		},
		{
			name:    "invalid URL",
			url:     "://invalid",
			wantErr: true,
		},

		// SSRF IPv4 Loopback & Private Addresses
		{
			name:    "ssrf ipv4 loopback 127.0.0.1",
			url:     "http://127.0.0.1/admin",
			wantErr: true,
		},
		{
			name:    "ssrf ipv4 loopback with port",
			url:     "http://127.0.0.1:8080/metrics",
			wantErr: true,
		},
		{
			name:    "ssrf ipv4 10.0.0.0/8 private",
			url:     "http://10.0.1.5:9000/internal",
			wantErr: true,
		},
		{
			name:    "ssrf ipv4 172.16.0.0/12 private",
			url:     "http://172.16.0.1/status",
			wantErr: true,
		},
		{
			name:    "ssrf ipv4 192.168.0.0/16 private",
			url:     "http://192.168.1.1/router",
			wantErr: true,
		},
		{
			name:    "ssrf aws metadata 169.254.169.254",
			url:     "http://169.254.169.254/latest/meta-data/",
			wantErr: true,
		},
		{
			name:    "ssrf 0.0.0.0 unspecified",
			url:     "http://0.0.0.0:8080",
			wantErr: true,
		},

		// SSRF IPv6 Addresses
		{
			name:    "ssrf ipv6 loopback ::1",
			url:     "http://[::1]/secret",
			wantErr: true,
		},
		{
			name:    "ssrf ipv6 unique local fc00::1",
			url:     "http://[fc00::1]:8080",
			wantErr: true,
		},
		{
			name:    "ssrf ipv6 link local fe80::1",
			url:     "http://[fe80::1]",
			wantErr: true,
		},

		// SSRF Hostnames
		{
			name:    "ssrf localhost hostname",
			url:     "http://localhost:3000",
			wantErr: true,
		},
		{
			name:    "ssrf sub.localhost hostname",
			url:     "http://service.localhost",
			wantErr: true,
		},
		{
			name:    "ssrf internal domain",
			url:     "http://database.internal:5432",
			wantErr: true,
		},
		{
			name:    "ssrf local domain",
			url:     "http://myhost.local",
			wantErr: true,
		},
		{
			name:    "ssrf lan domain",
			url:     "http://gateway.lan",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url)

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"ValidateURL(%q) error = %v, wantErr = %v",
					tt.url,
					err,
					tt.wantErr,
				)
			}
		})
	}
}
