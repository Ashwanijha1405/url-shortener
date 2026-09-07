package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		expectedIP string
	}{
		{
			name:       "remote addr with port",
			remoteAddr: "192.0.2.1:12345",
			expectedIP: "192.0.2.1",
		},
		{
			name:       "x-forwarded-for single IP",
			remoteAddr: "192.0.2.1:12345",
			xff:        "203.0.113.195",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "x-forwarded-for chain",
			remoteAddr: "192.0.2.1:12345",
			xff:        "203.0.113.195, 70.41.3.18, 150.172.238.178",
			expectedIP: "203.0.113.195",
		},
		{
			name:       "x-real-ip header",
			remoteAddr: "192.0.2.1:12345",
			xri:        "198.51.100.1",
			expectedIP: "198.51.100.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			ip := GetClientIP(req)
			if ip != tt.expectedIP {
				t.Fatalf("GetClientIP() = %q, expected %q", ip, tt.expectedIP)
			}
		})
	}
}

func TestIPRateLimiterAllow(t *testing.T) {
	limiter := NewIPRateLimiter(2.0, 2)
	defer limiter.Close()

	ip := "192.0.2.100"

	// Initial burst = 2 tokens
	allowed, _ := limiter.Allow(ip)
	if !allowed {
		t.Fatal("expected 1st request to be allowed")
	}

	allowed, _ = limiter.Allow(ip)
	if !allowed {
		t.Fatal("expected 2nd request to be allowed")
	}

	// 3rd request should be rejected immediately
	allowed, wait := limiter.Allow(ip)
	if allowed {
		t.Fatal("expected 3rd request to be rejected")
	}
	if wait <= 0 {
		t.Fatalf("expected wait duration > 0, got %v", wait)
	}

	// Different IP should still have full burst
	otherIP := "192.0.2.200"
	allowed, _ = limiter.Allow(otherIP)
	if !allowed {
		t.Fatal("expected different IP to be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	limiter := NewIPRateLimiter(1.0, 1)
	defer limiter.Close()

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))

	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	})

	mw := RateLimit(limiter, logger)
	handler := mw(dummyHandler)

	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req1.RemoteAddr = "192.0.2.50:5000"
	rec1 := httptest.NewRecorder()

	handler.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 1st request status 200, got %d", rec1.Code)
	}

	// 2nd request should exceed rate limit (429)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	req2.RemoteAddr = "192.0.2.50:5000"
	rec2 := httptest.NewRecorder()

	handler.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 2nd request status 429, got %d", rec2.Code)
	}

	if retryAfter := rec2.Header().Get("Retry-After"); retryAfter == "" {
		t.Fatal("expected Retry-After header in 429 response")
	}

	if !strings.Contains(rec2.Body.String(), `"rate limit exceeded`) {
		t.Fatalf("expected rate limit error in response body, got %s", rec2.Body.String())
	}

	if !strings.Contains(logBuf.String(), `"rate limit exceeded"`) {
		t.Fatalf("expected rate limit log in logger, got %s", logBuf.String())
	}
}

func TestRateLimitMiddlewareNilLimiter(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := RateLimit(nil, nil)
	handler := mw(dummyHandler)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 with nil limiter, got %d", rec.Code)
	}
}
