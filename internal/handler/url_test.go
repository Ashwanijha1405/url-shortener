package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockURLRepository struct {
	createFunc         func(ctx context.Context, shortCode string, originalURL string) error
	getByShortCodeFunc func(ctx context.Context, shortCode string) (string, error)
}

func (m *mockURLRepository) Create(
	ctx context.Context,
	shortCode string,
	originalURL string,
) error {
	return m.createFunc(ctx, shortCode, originalURL)
}

func (m *mockURLRepository) GetByShortCode(
	ctx context.Context,
	shortCode string,
) (string, error) {
	return m.getByShortCodeFunc(ctx, shortCode)
}

func TestCreateURL(t *testing.T) {
	repo := &mockURLRepository{
		createFunc: func(
			ctx context.Context,
			shortCode string,
			originalURL string,
		) error {
			if originalURL != "https://github.com" {
				t.Fatalf("unexpected URL: %s", originalURL)
			}

			if shortCode == "" {
				t.Fatal("expected short code to be generated")
			}

			return nil
		},
		getByShortCodeFunc: func(
			ctx context.Context,
			shortCode string,
		) (string, error) {
			return "", nil
		},
	}

	h := NewHandler(repo)

	body := `{"url":"https://github.com"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if !strings.Contains(rec.Body.String(), "short_code") {
		t.Fatalf("expected short_code in response, got: %s", rec.Body.String())
	}
}

func TestCreateURLInvalidRequest(t *testing.T) {
	repo := &mockURLRepository{}

	h := NewHandler(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`{"url":""}`),
	)

	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateURLInvalidJSON(t *testing.T) {
	repo := &mockURLRepository{}

	h := NewHandler(repo)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(`invalid-json`),
	)

	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestCreateURLRepositoryError(t *testing.T) {
	repo := &mockURLRepository{
		createFunc: func(
			ctx context.Context,
			shortCode string,
			originalURL string,
		) error {
			return errors.New("database error")
		},
	}

	h := NewHandler(repo)

	body := `{"url":"https://github.com"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(body),
	)

	rec := httptest.NewRecorder()

	h.CreateURL(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			rec.Code,
		)
	}
}

func TestRedirectURL(t *testing.T) {
	repo := &mockURLRepository{
		getByShortCodeFunc: func(
			ctx context.Context,
			shortCode string,
		) (string, error) {

			if shortCode != "abc123" {
				t.Fatalf("unexpected short code: %s", shortCode)
			}

			return "https://github.com", nil
		},
	}

	h := NewHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/abc123",
		nil,
	)

	req.SetPathValue("shortCode", "abc123")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusFound,
			rec.Code,
		)
	}

	location := rec.Header().Get("Location")

	if location != "https://github.com" {
		t.Fatalf(
			"expected Location https://github.com, got %s",
			location,
		)
	}
}

func TestRedirectURLNotFound(t *testing.T) {
	repo := &mockURLRepository{
		getByShortCodeFunc: func(
			ctx context.Context,
			shortCode string,
		) (string, error) {
			return "", errors.New("not found")
		},
	}

	h := NewHandler(repo)

	req := httptest.NewRequest(
		http.MethodGet,
		"/doesnotexist",
		nil,
	)

	req.SetPathValue("shortCode", "doesnotexist")

	rec := httptest.NewRecorder()

	h.RedirectURL(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}
