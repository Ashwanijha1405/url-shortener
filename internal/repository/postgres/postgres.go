package postgres

import (
	"context"
	"fmt"

	"github.com/Ashwanijha1405/url-shortener/internal/database"
)

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	shortCode string,
	originalURL string,
) error {

	_, err := r.db.Pool.Exec(
		ctx,
		`INSERT INTO urls (short_code, original_url)
		 VALUES ($1, $2)`,
		shortCode,
		originalURL,
	)

	if err != nil {
		return fmt.Errorf("create URL: %w", err)
	}

	return nil
}

func (r *Repository) GetByShortCode(
	ctx context.Context,
	shortCode string,
) (string, error) {

	var originalURL string

	err := r.db.Pool.QueryRow(
		ctx,
		`SELECT original_url
		 FROM urls
		 WHERE short_code = $1`,
		shortCode,
	).Scan(&originalURL)

	if err != nil {
		return "", fmt.Errorf("get URL: %w", err)
	}

	return originalURL, nil
}
