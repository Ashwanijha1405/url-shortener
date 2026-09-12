package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ashwanijha1405/url-shortener/internal/service"
)

type mockURLService struct {
	createShortURLFunc func(ctx context.Context, originalURL string) (string, error)
	resolveURLFunc     func(ctx context.Context, shortCode string) (string, error)
}

func (m *mockURLService) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	if m.createShortURLFunc != nil {
		return m.createShortURLFunc(ctx, originalURL)
	}
	return "", nil
}

func (m *mockURLService) CreateShortURLWithMetadata(ctx context.Context, originalURL string) (service.CreateResult, error) {
	code, err := m.CreateShortURL(ctx, originalURL)
	return service.CreateResult{
		ShortCode: code,
		Cache:     "miss",
	}, err
}

func (m *mockURLService) ResolveURL(ctx context.Context, shortCode string) (string, error) {
	if m.resolveURLFunc != nil {
		return m.resolveURLFunc(ctx, shortCode)
	}
	return "", nil
}

func TestCreateURL(t *testing.T) {
	svc := &mockURLService{
		createShortURLFunc: func(ctx context.Context, originalURL string) (string, error) {
			if originalURL != "https://github.com" {
				t.Fatalf("unexpected URL: %s", originalURL)
			}
			return "abc1234", nil
		},
	}

	h := NewHandler(svc)

	body := `{"url":"https://github.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if !strings.Contains(rec.Body.String(), `"short_code":"abc1234"`) {
		t.Fatalf("expected short_code in response, got: %s", rec.Body.String())
	}
}

func TestCreateURLInvalidRequest(t *testing.T) {
	svc := &mockURLService{
		createShortURLFunc: func(ctx context.Context, originalURL string) (string, error) {
			return "", service.ErrInvalidInput
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":""}`))
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateURLInvalidJSON(t *testing.T) {
	svc := &mockURLService{}
	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`invalid-json`))
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateURLOversizedBody(t *testing.T) {
	svc := &mockURLService{}
	h := NewHandler(svc, WithMaxBodyBytes(50))

	largeBody := `{"url":"https://example.com/` + strings.Repeat("x", 100) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(largeBody))
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for oversized body, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateURLConflict(t *testing.T) {
	svc := &mockURLService{
		createShortURLFunc: func(ctx context.Context, originalURL string) (string, error) {
			return "", service.ErrConflict
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://github.com"}`))
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestCreateURLInternalError(t *testing.T) {
	svc := &mockURLService{
		createShortURLFunc: func(ctx context.Context, originalURL string) (string, error) {
			return "", errors.New("db error")
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", strings.NewReader(`{"url":"https://github.com"}`))
	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestRedirectURL(t *testing.T) {
	svc := &mockURLService{
		resolveURLFunc: func(ctx context.Context, shortCode string) (string, error) {
			if shortCode != "abc1234" {
				t.Fatalf("unexpected short code: %s", shortCode)
			}
			return "https://github.com", nil
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/abc1234", nil)
	req.SetPathValue("shortCode", "abc1234")
	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "https://github.com" {
		t.Fatalf("expected Location https://github.com, got %s", location)
	}
}

func TestRedirectURLNotFound(t *testing.T) {
	svc := &mockURLService{
		resolveURLFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "", service.ErrNotFound
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/doesnotexist", nil)
	req.SetPathValue("shortCode", "doesnotexist")
	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestRedirectURLInternalErrorNotMasked(t *testing.T) {
	svc := &mockURLService{
		resolveURLFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "", errors.New("database connection down")
		},
	}

	h := NewHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/anyCode", nil)
	req.SetPathValue("shortCode", "anyCode")
	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500 for database error (unmasked), got %d", rec.Code)
	}
}
