package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockReadinessChecker struct {
	pingFunc func(ctx context.Context) error
}

func (m *mockReadinessChecker) Ping(ctx context.Context) error {
	if m.pingFunc != nil {
		return m.pingFunc(ctx)
	}
	return nil
}

func TestLegacyHealth(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	Health(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `{"status":"ok"}`) {
		t.Fatalf("expected status ok in body, got %s", rec.Body.String())
	}
}

func TestHealthLive(t *testing.T) {
	h := NewHealthHandler(nil)
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	h.Live(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `{"status":"ok"}`) {
		t.Fatalf("expected status ok in body, got %s", rec.Body.String())
	}
}

func TestHealthReadyHealthyWithDBAndCache(t *testing.T) {
	dbChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	cacheChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	h := NewHealthHandler(dbChecker, WithCacheChecker(cacheChecker))
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected status ok in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"database":"healthy"`) {
		t.Fatalf("expected database healthy in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cache":"healthy"`) {
		t.Fatalf("expected cache healthy in body, got %s", rec.Body.String())
	}
}

func TestHealthReadyDBDegraded(t *testing.T) {
	dbChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return errors.New("connection pool exhausted")
		},
	}
	cacheChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	h := NewHealthHandler(dbChecker, WithCacheChecker(cacheChecker))
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("expected status unavailable in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"database":"unhealthy: connection pool exhausted"`) {
		t.Fatalf("expected database unhealthy in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cache":"healthy"`) {
		t.Fatalf("expected cache healthy in body, got %s", rec.Body.String())
	}
}

func TestHealthReadyCacheDegraded(t *testing.T) {
	dbChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return nil
		},
	}
	cacheChecker := &mockReadinessChecker{
		pingFunc: func(ctx context.Context) error {
			return errors.New("redis connection refused")
		},
	}
	h := NewHealthHandler(dbChecker, WithCacheChecker(cacheChecker))
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	h.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"unavailable"`) {
		t.Fatalf("expected status unavailable in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"cache":"unhealthy: redis connection refused"`) {
		t.Fatalf("expected cache unhealthy in body, got %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"database":"healthy"`) {
		t.Fatalf("expected database healthy in body, got %s", rec.Body.String())
	}
}
