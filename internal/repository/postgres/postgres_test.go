package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/database"
)

func setupTestRepository(t *testing.T) *Repository {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(ctx)
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
}
