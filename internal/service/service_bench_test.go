package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
)

type fastMockCache struct {
	mu   sync.RWMutex
	data map[string]string
}

func newFastMockCache() *fastMockCache {
	return &fastMockCache{
		data: make(map[string]string),
	}
}

func (f *fastMockCache) Get(ctx context.Context, key string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if val, ok := f.data[key]; ok {
		return val, nil
	}
	return "", cache.ErrCacheMiss
}

func (f *fastMockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fastMockCache) Delete(ctx context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

func (f *fastMockCache) Ping(ctx context.Context) error {
	return nil
}

func (f *fastMockCache) Close() error {
	return nil
}

func BenchmarkResolveURL_CacheHit(b *testing.B) {
	c := newFastMockCache()
	c.data["url:fastHit"] = "https://golang.org/doc/"

	repo := &mockRepository{}
	svc := New(repo, WithCache(c, 24*time.Hour))
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.ResolveURL(ctx, "fastHit")
		if err != nil {
			b.Fatalf("ResolveURL failed: %v", err)
		}
	}
}

func BenchmarkResolveURL_CacheMiss_DBFallback(b *testing.B) {
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "https://golang.org/pkg/", nil
		},
	}
	svc := New(repo) // using NoopCache
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.ResolveURL(ctx, "missCode")
		if err != nil {
			b.Fatalf("ResolveURL failed: %v", err)
		}
	}
}

func BenchmarkCreateShortURL(b *testing.B) {
	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			return nil
		},
	}
	c := newFastMockCache()
	svc := New(repo, WithCache(c, 24*time.Hour))
	ctx := context.Background()
	targetURL := "https://golang.org/project"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := svc.CreateShortURL(ctx, targetURL)
		if err != nil {
			b.Fatalf("CreateShortURL failed: %v", err)
		}
	}
}
