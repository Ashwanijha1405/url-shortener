package repository

import "context"

type URLRepository interface {
	Create(ctx context.Context, shortCode string, originalURL string) error
	GetByShortCode(ctx context.Context, shortCode string) (string, error)
}
