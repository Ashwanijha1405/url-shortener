package cache

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrCacheMiss is returned when a requested key does not exist in the cache.
	ErrCacheMiss = errors.New("cache miss")
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Ping(ctx context.Context) error
	Close() error
}

// NoopCache provides a non-operational cache implementation for setups without cache.
type NoopCache struct{}

func NewNoopCache() *NoopCache {
	return &NoopCache{}
}

func (n *NoopCache) Get(ctx context.Context, key string) (string, error) {
	return "", ErrCacheMiss
}

func (n *NoopCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}

func (n *NoopCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (n *NoopCache) Ping(ctx context.Context) error {
	return nil
}

func (n *NoopCache) Close() error {
	return nil
}
