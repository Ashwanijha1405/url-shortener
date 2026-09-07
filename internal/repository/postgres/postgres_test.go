package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/config"
	"github.com/Ashwanijha1405/url-shortener/internal/database"
	"github.com/Ashwanijha1405/url-shortener/internal/repository"
)

func setupTestRepository(t *testing.T) *Repository {
	t.Helper()

	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("skipping postgres integration test: DATABASE_URL not set")
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Pool.Close()
	})

	return NewRepository(db)
}

func TestCreate(t *testing.T) {
	repo := setupTestRepository(t)

	ctx := context.Background()

	shortCode := "crt123"
	originalURL := "https://example.com"

	err := repo.Create(ctx, shortCode, originalURL)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}

	var storedURL string

	err = repo.db.Pool.QueryRow(
		ctx,
		`SELECT original_url
		 FROM urls
		 WHERE short_code = $1`,
		shortCode,
	).Scan(&storedURL)

	if err != nil {
		t.Fatalf("failed to query created URL: %v", err)
	}

	if storedURL != originalURL {
		t.Fatalf(
			"expected URL %q, got %q",
			originalURL,
			storedURL,
		)
	}

	_, err = repo.db.Pool.Exec(
		ctx,
		`DELETE FROM urls WHERE short_code = $1`,
		shortCode,
	)

	if err != nil {
		t.Fatalf("failed to clean up test data: %v", err)
	}
}

func TestGetByShortCode(t *testing.T) {
	repo := setupTestRepository(t)

	ctx := context.Background()

	shortCode := "get123"
	originalURL := "https://github.com"

	err := repo.Create(ctx, shortCode, originalURL)
	if err != nil {
		t.Fatalf("failed to create test URL: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.db.Pool.Exec(
			ctx,
			`DELETE FROM urls WHERE short_code = $1`,
			shortCode,
		)
	})

	got, err := repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		t.Fatalf("GetByShortCode() returned error: %v", err)
	}

	if got != originalURL {
		t.Fatalf(
			"expected URL %q, got %q",
			originalURL,
			got,
		)
	}
}

func TestGetByShortCodeNotFound(t *testing.T) {
	repo := setupTestRepository(t)

	ctx := context.Background()

	_, err := repo.GetByShortCode(ctx, "missing")

	if err == nil {
		t.Fatal("expected error for nonexistent short code")
	}

	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestCreateDuplicateShortCode(t *testing.T) {
	repo := setupTestRepository(t)

	ctx := context.Background()

	shortCode := "dup123"
	originalURL := "https://example.com"

	err := repo.Create(ctx, shortCode, originalURL)
	if err != nil {
		t.Fatalf("failed to create first URL: %v", err)
	}

	t.Cleanup(func() {
		_, _ = repo.db.Pool.Exec(
			ctx,
			`DELETE FROM urls WHERE short_code = $1`,
			shortCode,
		)
	})

	err = repo.Create(ctx, shortCode, "https://github.com")

	if err == nil {
		t.Fatal("expected error when creating duplicate short code")
	}

	if !errors.Is(err, repository.ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}
