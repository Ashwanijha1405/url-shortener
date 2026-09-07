package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
	"github.com/Ashwanijha1405/url-shortener/internal/config"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func New(cfg config.RedisConfig) (*Client, error) {
	if cfg.Addr == "" {
		return nil, fmt.Errorf("redis address is required")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		client: rdb,
	}, nil
}

func NewFromClient(rdb *redis.Client) *Client {
	return &Client{
		client: rdb,
	}
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", cache.ErrCacheMiss
		}
		return "", fmt.Errorf("redis get: %w", err)
	}
	return val, nil
}

func (c *Client) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis del: %w", err)
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func (c *Client) Close() error {
	if c.client == nil {
		return nil
	}
	return c.client.Close()
}
