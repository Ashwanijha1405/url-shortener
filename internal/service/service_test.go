package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
	"github.com/Ashwanijha1405/url-shortener/internal/repository"
)

type mockRepository struct {
	createFunc         func(ctx context.Context, shortCode string, originalURL string) error
	getByShortCodeFunc func(ctx context.Context, shortCode string) (string, error)
}

func (m *mockRepository) Create(ctx context.Context, shortCode string, originalURL string) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, shortCode, originalURL)
	}
	return nil
}

func (m *mockRepository) GetByShortCode(ctx context.Context, shortCode string) (string, error) {
	if m.getByShortCodeFunc != nil {
		return m.getByShortCodeFunc(ctx, shortCode)
	}
	return "", nil
}

type mockCache struct {
	getFunc func(ctx context.Context, key string) (string, error)
	setFunc func(ctx context.Context, key string, value string, ttl time.Duration) error
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return "", cache.ErrCacheMiss
}

func (m *mockCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	return nil
}

func (m *mockCache) Ping(ctx context.Context) error {
	return nil
}

func (m *mockCache) Close() error {
	return nil
}

func TestCreateShortURLSuccess(t *testing.T) {
	called := false
	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			called = true
			if len(shortCode) != DefaultCodeLength {
				t.Errorf("expected shortCode of length %d, got %d", DefaultCodeLength, len(shortCode))
			}
			if originalURL != "https://golang.org" {
				t.Errorf("expected originalURL 'https://golang.org', got %q", originalURL)
			}
			return nil
		},
	}

	svc := New(repo)
	code, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if code == "" {
		t.Fatal("expected non-empty short code")
	}
	if !called {
		t.Fatal("expected repository Create to be called")
	}
}

func TestCreateShortURLPrewarmsCache(t *testing.T) {
	var prewarmedKey, prewarmedVal string
	c := &mockCache{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			prewarmedKey = key
			prewarmedVal = value
			return nil
		},
	}
	repo := &mockRepository{}

	svc := New(repo, WithCache(c, time.Hour))
	code, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	expectedKey := "url:" + code
	if prewarmedKey != expectedKey {
		t.Fatalf("expected cache key %q, got %q", expectedKey, prewarmedKey)
	}
	if prewarmedVal != "https://golang.org" {
		t.Fatalf("expected cache val 'https://golang.org', got %q", prewarmedVal)
	}
}

func TestCreateShortURLCacheFailOpen(t *testing.T) {
	c := &mockCache{
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			return errors.New("redis connection failure")
		},
	}
	repo := &mockRepository{}

	svc := New(repo, WithCache(c, time.Hour))
	code, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("expected creation to succeed even if cache fails, got %v", err)
	}
	if code == "" {
		t.Fatal("expected non-empty short code")
	}
}

func TestCreateShortURLInvalidInput(t *testing.T) {
	repo := &mockRepository{}
	svc := New(repo)

	testCases := []struct {
		name string
		url  string
	}{
		{"empty url", ""},
		{"invalid scheme", "ftp://example.com"},
		{"missing host", "https://"},
		{"oversized url", "https://example.com/" + strings.Repeat("a", 2050)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.CreateShortURL(context.Background(), tc.url)
			if err == nil {
				t.Fatalf("expected error for url %q, got nil", tc.url)
			}
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestCreateShortURLCollisionRetrySuccess(t *testing.T) {
	attempts := 0
	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			attempts++
			if attempts < 3 {
				return repository.ErrConflict
			}
			return nil
		},
	}

	svc := New(repo, WithMaxRetries(3))
	code, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if code == "" {
		t.Fatal("expected non-empty short code")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestCreateShortURLCollisionExhausted(t *testing.T) {
	attempts := 0
	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			attempts++
			return repository.ErrConflict
		},
	}

	svc := New(repo, WithMaxRetries(3))
	_, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err == nil {
		t.Fatal("expected error when collisions exhausted, got nil")
	}
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestCreateShortURLRepositoryError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			return repoErr
		},
	}

	svc := New(repo)
	_, err := svc.CreateShortURL(context.Background(), "https://golang.org")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repoErr in chain, got %v", err)
	}
}

func TestResolveURLCacheHit(t *testing.T) {
	dbCalled := false
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			dbCalled = true
			return "", nil
		},
	}
	c := &mockCache{
		getFunc: func(ctx context.Context, key string) (string, error) {
			if key == "url:fastCode" {
				return "https://cached-destination.com", nil
			}
			return "", cache.ErrCacheMiss
		},
	}

	svc := New(repo, WithCache(c, time.Hour))
	url, err := svc.ResolveURL(context.Background(), "fastCode")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if url != "https://cached-destination.com" {
		t.Fatalf("expected cached destination, got %q", url)
	}
	if dbCalled {
		t.Fatal("expected database NOT to be queried on cache hit")
	}
}

func TestResolveURLCacheMissPopulatesCache(t *testing.T) {
	var populatedKey, populatedVal string
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "https://from-db.com", nil
		},
	}
	c := &mockCache{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "", cache.ErrCacheMiss
		},
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			populatedKey = key
			populatedVal = value
			return nil
		},
	}

	svc := New(repo, WithCache(c, time.Hour))
	url, err := svc.ResolveURL(context.Background(), "coldCode")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if url != "https://from-db.com" {
		t.Fatalf("expected 'https://from-db.com', got %q", url)
	}
	if populatedKey != "url:coldCode" || populatedVal != "https://from-db.com" {
		t.Fatalf("expected cache to be populated with key 'url:coldCode' and value 'https://from-db.com', got key=%q val=%q", populatedKey, populatedVal)
	}
}

func TestResolveURLCacheFailureFallsBackToDB(t *testing.T) {
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "https://from-db-fallback.com", nil
		},
	}
	c := &mockCache{
		getFunc: func(ctx context.Context, key string) (string, error) {
			return "", errors.New("redis timeout")
		},
	}

	svc := New(repo, WithCache(c, time.Hour))
	url, err := svc.ResolveURL(context.Background(), "anyCode")
	if err != nil {
		t.Fatalf("expected success via DB fallback, got %v", err)
	}
	if url != "https://from-db-fallback.com" {
		t.Fatalf("expected fallback destination, got %q", url)
	}
}

func TestResolveURLEmptyCode(t *testing.T) {
	repo := &mockRepository{}
	svc := New(repo)

	_, err := svc.ResolveURL(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty short code, got nil")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestResolveURLNotFound(t *testing.T) {
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "", repository.ErrNotFound
		},
	}

	svc := New(repo)
	_, err := svc.ResolveURL(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for not found, got nil")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestResolveURLRepositoryError(t *testing.T) {
	repoErr := errors.New("db query timeout")
	repo := &mockRepository{
		getByShortCodeFunc: func(ctx context.Context, shortCode string) (string, error) {
			return "", repoErr
		},
	}

	svc := New(repo)
	_, err := svc.ResolveURL(context.Background(), "someCode")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repoErr in chain, got %v", err)
	}
}

func TestCreateShortURLWithMetadataCacheHitAndMiss(t *testing.T) {
	stored := make(map[string]string)
	repoCallCount := 0

	repo := &mockRepository{
		createFunc: func(ctx context.Context, shortCode string, originalURL string) error {
			repoCallCount++
			return nil
		},
	}

	c := &mockCache{
		getFunc: func(ctx context.Context, key string) (string, error) {
			if val, ok := stored[key]; ok {
				return val, nil
			}
			return "", cache.ErrCacheMiss
		},
		setFunc: func(ctx context.Context, key string, value string, ttl time.Duration) error {
			stored[key] = value
			return nil
		},
	}

	svc := New(repo, WithCache(c, time.Hour))

	// First call: should be cache MISS (persisted in DB + cache pre-warmed)
	res1, err := svc.CreateShortURLWithMetadata(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	if res1.Cache != "miss" {
		t.Fatalf("expected cache miss, got %s", res1.Cache)
	}
	if repoCallCount != 1 {
		t.Fatalf("expected 1 repo call, got %d", repoCallCount)
	}

	// Second call with same URL: should be cache HIT (returned directly from Redis without DB call)
	res2, err := svc.CreateShortURLWithMetadata(context.Background(), "https://golang.org")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if res2.Cache != "hit" {
		t.Fatalf("expected cache hit, got %s", res2.Cache)
	}
	if res2.ShortCode != res1.ShortCode {
		t.Fatalf("expected same shortCode %s, got %s", res1.ShortCode, res2.ShortCode)
	}
	if repoCallCount != 1 {
		t.Fatalf("expected repoCallCount to remain 1 on cache hit, got %d", repoCallCount)
	}
}

