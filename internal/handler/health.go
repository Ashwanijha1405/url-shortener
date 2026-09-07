package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type ReadinessChecker interface {
	Ping(ctx context.Context) error
}

type HealthHandler struct {
	db    ReadinessChecker
	cache ReadinessChecker
}

type HealthOption func(*HealthHandler)

func WithCacheChecker(c ReadinessChecker) HealthOption {
	return func(h *HealthHandler) {
		h.cache = c
	}
}

func NewHealthHandler(db ReadinessChecker, opts ...HealthOption) *HealthHandler {
	h := &HealthHandler{
		db: db,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// Health is the legacy /health endpoint maintained for backward compatibility.
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// Live handles GET /health/live (process liveness check).
func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// Ready handles GET /health/ready (dependency readiness check for database and cache).
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	if h.db != nil {
		if err := h.db.Ping(ctx); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["database"] = "healthy"
		}
	}

	if h.cache != nil {
		if err := h.cache.Ping(ctx); err != nil {
			checks["cache"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["cache"] = "healthy"
		}
	}

	if !allHealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		checks["status"] = "unavailable"
		_ = json.NewEncoder(w).Encode(checks)
		return
	}

	checks["status"] = "ok"
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(checks)
}
