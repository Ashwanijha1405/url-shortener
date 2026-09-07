package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type clientBucket struct {
	tokens   float64
	lastSeen time.Time
}

type IPRateLimiter struct {
	mu       sync.Mutex
	clients  map[string]*clientBucket
	rate     float64 // tokens added per second
	burst    float64 // maximum bucket capacity
	stopChan chan struct{}
}

func NewIPRateLimiter(rate float64, burst int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		clients:  make(map[string]*clientBucket),
		rate:     rate,
		burst:    float64(burst),
		stopChan: make(chan struct{}),
	}

	go limiter.cleanupLoop(time.Minute, 3*time.Minute)

	return limiter
}

func (limiter *IPRateLimiter) cleanupLoop(interval, maxIdle time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			limiter.mu.Lock()
			now := time.Now()
			for ip, client := range limiter.clients {
				if now.Sub(client.lastSeen) > maxIdle {
					delete(limiter.clients, ip)
				}
			}
			limiter.mu.Unlock()
		case <-limiter.stopChan:
			return
		}
	}
}

func (limiter *IPRateLimiter) Close() {
	select {
	case <-limiter.stopChan:
		// already closed
	default:
		close(limiter.stopChan)
	}
}

func (limiter *IPRateLimiter) Allow(ip string) (bool, time.Duration) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	client, exists := limiter.clients[ip]
	if !exists {
		limiter.clients[ip] = &clientBucket{
			tokens:   limiter.burst - 1.0,
			lastSeen: now,
		}
		return true, 0
	}

	elapsed := now.Sub(client.lastSeen).Seconds()
	client.lastSeen = now

	// Refill tokens based on elapsed time capped at burst
	client.tokens = math.Min(limiter.burst, client.tokens+elapsed*limiter.rate)

	if client.tokens >= 1.0 {
		client.tokens -= 1.0
		return true, 0
	}

	// Calculate wait time until at least 1 token is available
	missing := 1.0 - client.tokens
	waitSec := missing / limiter.rate
	return false, time.Duration(waitSec * float64(time.Second))
}

func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (first IP in comma-separated chain)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if ip != "" {
			return ip
		}
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}

	return r.RemoteAddr
}

func RateLimit(limiter *IPRateLimiter, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}

			ip := GetClientIP(r)
			allowed, waitDuration := limiter.Allow(ip)

			if !allowed {
				retryAfterSec := int(math.Ceil(waitDuration.Seconds()))
				if retryAfterSec < 1 {
					retryAfterSec = 1
				}

				reqID := GetRequestID(r.Context())

				if logger != nil {
					logger.WarnContext(r.Context(), "rate limit exceeded",
						slog.String("request_id", reqID),
						slog.String("client_ip", ip),
						slog.String("path", r.URL.Path),
						slog.Int("retry_after_seconds", retryAfterSec),
					)
				}

				w.Header().Set("Retry-After", fmt.Sprintf("%d", retryAfterSec))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"error": "rate limit exceeded, please try again later",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
