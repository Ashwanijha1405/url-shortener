package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					reqID := GetRequestID(r.Context())
					stack := string(debug.Stack())

					logger.ErrorContext(
						r.Context(),
						"panic recovered in http handler",
						slog.String("request_id", reqID),
						slog.String("panic", fmt.Sprintf("%v", rvr)),
						slog.String("stack", stack),
						slog.String("method", r.Method),
						slog.String("path", r.URL.Path),
					)

					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
