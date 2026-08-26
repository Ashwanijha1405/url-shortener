package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Ashwanijha1405/url-shortener/internal/database"
	"github.com/Ashwanijha1405/url-shortener/internal/repository/postgres"
)

func TestURLShortenerIntegration(t *testing.T) {
	ctx := context.Background()

	// Connect to the real PostgreSQL database.
	db, err := database.Connect(ctx)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Use the real PostgreSQL repository.
	repo := postgres.NewRepository(db)

	// Use the real HTTP handler.
	h := NewHandler(repo)

	// ------------------------------------------------
	// Step 1: Create a short URL
	// ------------------------------------------------

	requestBody := `{"url":"https://github.com"}`

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/urls",
		strings.NewReader(requestBody),
	)

	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()

	h.CreateURL(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf(
			"expected create status %d, got %d",
			http.StatusCreated,
			createRec.Code,
		)
	}

	var response CreateURLResponse

	if err := json.NewDecoder(createRec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	if response.ShortCode == "" {
		t.Fatal("expected short code, got empty string")
	}

	t.Logf("created short code: %s", response.ShortCode)

	// ------------------------------------------------
	// Cleanup
	// ------------------------------------------------

	t.Cleanup(func() {
		_, err := db.Pool.Exec(
			context.Background(),
			"DELETE FROM urls WHERE short_code = $1",
			response.ShortCode,
		)

		if err != nil {
			t.Logf("failed to clean up test URL: %v", err)
		}

		db.Pool.Close()
	})

	// ------------------------------------------------
	// Step 2: Redirect using the generated short code
	// ------------------------------------------------

	redirectReq := httptest.NewRequest(
		http.MethodGet,
		"/"+response.ShortCode,
		nil,
	)

	redirectReq.SetPathValue(
		"shortCode",
		response.ShortCode,
	)

	redirectRec := httptest.NewRecorder()

	h.RedirectURL(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf(
			"expected redirect status %d, got %d",
			http.StatusFound,
			redirectRec.Code,
		)
	}

	location := redirectRec.Header().Get("Location")

	if location != "https://github.com" {
		t.Fatalf(
			"expected redirect location %q, got %q",
			"https://github.com",
			location,
		)
	}
}
