package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
	bytesWritten int64
}

func newResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	return &responseWriterInterceptor{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriterInterceptor) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriterInterceptor) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriterInterceptor(w)

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			reqID := GetRequestID(r.Context())

			attributes := []slog.Attr{
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.statusCode),
				slog.Int64("duration_ms", duration.Milliseconds()),
				slog.Int64("bytes", wrapped.bytesWritten),
				slog.String("remote_ip", r.RemoteAddr),
			}

			logLevel := slog.LevelInfo
			if wrapped.statusCode >= 500 {
				logLevel = slog.LevelError
			} else if wrapped.statusCode >= 400 {
				logLevel = slog.LevelWarn
			}

			logger.LogAttrs(r.Context(), logLevel, "http request handled", attributes...)
		})
	}
}
