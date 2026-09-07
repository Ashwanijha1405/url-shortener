package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Ashwanijha1405/url-shortener/internal/database"
	"github.com/Ashwanijha1405/url-shortener/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const pgErrUniqueViolation = "23505"

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
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation {
			return fmt.Errorf("%w: %v", repository.ErrConflict, err)
		}
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
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrNotFound
		}
		return "", fmt.Errorf("get URL: %w", err)
	}

	return originalURL, nil
}
