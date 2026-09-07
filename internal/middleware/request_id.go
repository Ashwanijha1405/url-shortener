package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type contextKey string

const RequestIDHeader = "X-Request-ID"
const RequestIDContextKey contextKey = "request_id"

func generateRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDHeader)
		if reqID == "" {
			reqID = generateRequestID()
		}

		w.Header().Set(RequestIDHeader, reqID)
		ctx := context.WithValue(r.Context(), RequestIDContextKey, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDContextKey).(string); ok {
		return val
	}
	return ""
}
