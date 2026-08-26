package validator

import "testing"

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid https URL",
			url:     "https://github.com",
			wantErr: false,
		},
		{
			name:    "valid http URL",
			url:     "http://example.com",
			wantErr: false,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			url:     "ftp://example.com",
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
