package repository

import (
	"context"
	"errors"
)

var (
	// ErrNotFound is returned when a requested record does not exist in the database.
	ErrNotFound = errors.New("record not found")

	// ErrConflict is returned when attempting to insert a duplicate record violating unique constraints.
	ErrConflict = errors.New("record already exists")
)

type URLRepository interface {
	Create(ctx context.Context, shortCode string, originalURL string) error
	GetByShortCode(ctx context.Context, shortCode string) (string, error)
}
