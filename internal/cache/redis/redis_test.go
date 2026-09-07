package redis

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
	"github.com/Ashwanijha1405/url-shortener/internal/config"
)

func setupTestRedis(t *testing.T) *Client {
	t.Helper()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		t.Skip("skipping redis integration test: REDIS_ADDR not set")
	}

	cfg := config.RedisConfig{
		Addr: redisAddr,
	}

	client, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to connect to redis: %v", err)
	}

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func TestRedisLifecycle(t *testing.T) {
	client := setupTestRedis(t)
	ctx := context.Background()

	key := "test:url:abc1234"
	val := "https://example.com"

	// 1. Get nonexistent key -> ErrCacheMiss
	_, err := client.Get(ctx, key)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}

	// 2. Set key with TTL
	if err := client.Set(ctx, key, val, 10*time.Second); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// 3. Get key -> value
	got, err := client.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != val {
		t.Fatalf("expected %q, got %q", val, got)
	}

	// 4. Delete key
	if err := client.Delete(ctx, key); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// 5. Verify deleted
	_, err = client.Get(ctx, key)
	if !errors.Is(err, cache.ErrCacheMiss) {
		t.Fatalf("expected ErrCacheMiss after delete, got %v", err)
	}
}

func TestRedisNewEmptyAddr(t *testing.T) {
	cfg := config.RedisConfig{
		Addr: "",
	}
	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error when Addr is empty, got nil")
	}
}
