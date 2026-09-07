package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/repository"
)

type chaosCache struct {
	mu           sync.RWMutex
	store        map[string]string
	failureRate  float64
	failureError error
}

func newChaosCache(failureRate float64, failureError error) *chaosCache {
	return &chaosCache{
		store:        make(map[string]string),
		failureRate:  failureRate,
		failureError: failureError,
	}
}

func (c *chaosCache) Get(ctx context.Context, key string) (string, error) {
	if rand.Float64() < c.failureRate {
		return "", c.failureError
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	if val, ok := c.store[key]; ok {
		return val, nil
	}
	return "", ErrNotFound
}

func (c *chaosCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if rand.Float64() < c.failureRate {
		return c.failureError
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = value
	return nil
}

func (c *chaosCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, key)
	return nil
}

func (c *chaosCache) Ping(ctx context.Context) error {
	if c.failureRate >= 1.0 {
		return c.failureError
	}
	return nil
}

func (c *chaosCache) Close() error {
	return nil
}

func TestChaos_ConcurrentRedisTotalOutageFallback(t *testing.T) {
	outageCache := newChaosCache(1.0, errors.New("connection refused: dial tcp 127.0.0.1:6379: connect: connection refused"))

	dbRecords := map[string]string{
		"codeA": "https://example.com/a",
		"codeB": "https://example.com/b",
		"codeC": "https://example.com/c",
		"codeD": "https://example.com/d",
	}

	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			if url, ok := dbRecords[shortCode]; ok {
				return url, nil
			}
			return "", repository.ErrNotFound
		},
	}

	svc := New(repo, WithCache(outageCache, 24*time.Hour))

	const numWorkers = 100
	var wg sync.WaitGroup
	var successfulFallbacks int64

	keys := []string{"codeA", "codeB", "codeC", "codeD"}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		key := keys[i%len(keys)]
		expectedURL := dbRecords[key]

		go func(k, expected string) {
			defer wg.Done()
			url, err := svc.ResolveURL(context.Background(), k)
			if err != nil {
				t.Errorf("expected seamless DB fallback for key %q, got error: %v", k, err)
				return
			}
			if url != expected {
				t.Errorf("expected URL %q, got %q", expected, url)
				return
			}
			atomic.AddInt64(&successfulFallbacks, 1)
		}(key, expectedURL)
	}

	wg.Wait()

	if successfulFallbacks != int64(numWorkers) {
		t.Fatalf("expected %d successful fallbacks, got %d", numWorkers, successfulFallbacks)
	}
}

func TestChaos_ConcurrentCollisionBurstRetries(t *testing.T) {
	var attemptsMap sync.Map
	var totalSuccesses int64

	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			// For each unique originalURL, simulate 2 collisions before succeeding on 3rd attempt
			val, _ := attemptsMap.LoadOrStore(originalURL, new(int64))
			countPtr := val.(*int64)
			attempt := atomic.AddInt64(countPtr, 1)
			if attempt < 3 {
				return repository.ErrConflict
			}
			return nil
		},
	}

	svc := New(repo, WithMaxRetries(3))

	const concurrentCreators = 50
	var wg sync.WaitGroup

	for i := 0; i < concurrentCreators; i++ {
		wg.Add(1)
		url := fmt.Sprintf("https://example.com/page/%d", i)

		go func(target string) {
			defer wg.Done()
			code, err := svc.CreateShortURL(context.Background(), target)
			if err != nil {
				t.Errorf("CreateShortURL failed under collision burst: %v", err)
				return
			}
			if code == "" {
				t.Errorf("expected non-empty short code")
				return
			}
			atomic.AddInt64(&totalSuccesses, 1)
		}(url)
	}

	wg.Wait()

	if totalSuccesses != int64(concurrentCreators) {
		t.Fatalf("expected %d successful creations under collision burst, got %d", concurrentCreators, totalSuccesses)
	}
}

func TestChaos_ConcurrentCacheStampede(t *testing.T) {
	c := newFastMockCache()
	targetCode := "hotKey123"
	targetURL := "https://example.com/super-viral-link"

	var dbQueryCount int64
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			if shortCode == targetCode {
				atomic.AddInt64(&dbQueryCount, 1)
				return targetURL, nil
			}
			return "", repository.ErrNotFound
		},
	}

	svc := New(repo, WithCache(c, 24*time.Hour))

	const readers = 80
	var wg sync.WaitGroup
	var successCount int64

	for i := 0; i < readers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			url, err := svc.ResolveURL(context.Background(), targetCode)
			if err != nil {
				t.Errorf("ResolveURL failed during stampede: %v", err)
				return
			}
			if url != targetURL {
				t.Errorf("expected URL %q, got %q", targetURL, url)
				return
			}
			atomic.AddInt64(&successCount, 1)
		}()
	}

	wg.Wait()

	if successCount != int64(readers) {
		t.Fatalf("expected %d successful reads, got %d", readers, successCount)
	}

	cachedVal, err := c.Get(context.Background(), "url:"+targetCode)
	if err != nil || cachedVal != targetURL {
		t.Fatalf("expected primed cache with %q, got val=%q, err=%v", targetURL, cachedVal, err)
	}
}
