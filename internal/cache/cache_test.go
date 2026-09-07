package cache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNoopCache(t *testing.T) {
	c := NewNoopCache()
	ctx := context.Background()

	_, err := c.Get(ctx, "any-key")
	if !errors.Is(err, ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	if err := c.Set(ctx, "any-key", "val", time.Hour); err != nil {
		t.Fatalf("expected nil from NoopCache Set, got %v", err)
	}

	if err := c.Delete(ctx, "any-key"); err != nil {
		t.Fatalf("expected nil from NoopCache Delete, got %v", err)
	}

	if err := c.Ping(ctx); err != nil {
		t.Fatalf("expected nil from NoopCache Ping, got %v", err)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("expected nil from NoopCache Close, got %v", err)
	}
}
