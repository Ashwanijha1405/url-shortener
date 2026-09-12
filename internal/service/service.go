package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
	"github.com/Ashwanijha1405/url-shortener/internal/generator"
	"github.com/Ashwanijha1405/url-shortener/internal/repository"
	"github.com/Ashwanijha1405/url-shortener/internal/validator"
)

const (
	DefaultMaxRetries = 3
	DefaultCodeLength = generator.DefaultLength
	DefaultCacheTTL   = 24 * time.Hour
)

var (
	ErrNotFound         = repository.ErrNotFound
	ErrConflict         = errors.New("short code collision limit exceeded")
	ErrInvalidInput     = errors.New("invalid input")
	ErrGenerationFailed = errors.New("failed to generate short code")
)

type CreateResult struct {
	ShortCode string
	Cache     string
}

type URLService interface {
	CreateShortURL(ctx context.Context, originalURL string) (string, error)
	CreateShortURLWithMetadata(ctx context.Context, originalURL string) (CreateResult, error)
	ResolveURL(ctx context.Context, shortCode string) (string, error)
}

type ServiceOption func(*Service)

func WithMaxRetries(retries int) ServiceOption {
	return func(s *Service) {
		if retries > 0 {
			s.maxRetries = retries
		}
	}
}

func WithCodeLength(length int) ServiceOption {
	return func(s *Service) {
		if length > 0 {
			s.codeLength = length
		}
	}
}

func WithCache(c cache.Cache, ttl time.Duration) ServiceOption {
	return func(s *Service) {
		if c != nil {
			s.cache = c
		}
		if ttl > 0 {
			s.cacheTTL = ttl
		}
	}
}

type Service struct {
	repo       repository.URLRepository
	cache      cache.Cache
	cacheTTL   time.Duration
	maxRetries int
	codeLength int
}

func New(repo repository.URLRepository, opts ...ServiceOption) *Service {
	s := &Service{
		repo:       repo,
		cache:      cache.NewNoopCache(),
		cacheTTL:   DefaultCacheTTL,
		maxRetries: DefaultMaxRetries,
		codeLength: DefaultCodeLength,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) cacheKey(shortCode string) string {
	return "url:" + shortCode
}

func (s *Service) reverseCacheKey(originalURL string) string {
	return "orig:" + originalURL
}

func (s *Service) CreateShortURLWithMetadata(ctx context.Context, originalURL string) (CreateResult, error) {
	if err := validator.ValidateURL(originalURL); err != nil {
		return CreateResult{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// 1. Check cache first for existing shortened URL (deduplication fast path)
	revKey := s.reverseCacheKey(originalURL)
	if cachedCode, err := s.cache.Get(ctx, revKey); err == nil && cachedCode != "" {
		return CreateResult{
			ShortCode: cachedCode,
			Cache:     "hit",
		}, nil
	}

	// 2. Cache MISS: generate short code and persist to database
	for attempt := 0; attempt < s.maxRetries; attempt++ {
		shortCode, err := generator.Generate(s.codeLength)
		if err != nil {
			return CreateResult{}, fmt.Errorf("%w: %v", ErrGenerationFailed, err)
		}

		err = s.repo.Create(ctx, shortCode, originalURL)
		if err == nil {
			// Cache reverse mapping for deduplication / hit path
			if cacheErr := s.cache.Set(ctx, revKey, shortCode, s.cacheTTL); cacheErr != nil {
				slog.WarnContext(ctx, "failed to cache reverse mapping on creation",
					slog.String("short_code", shortCode),
					slog.String("error", cacheErr.Error()),
				)
			}
			// Pre-warm cache for redirect lookups (fail-open: ignore or log error)
			if cacheErr := s.cache.Set(ctx, s.cacheKey(shortCode), originalURL, s.cacheTTL); cacheErr != nil {
				slog.WarnContext(ctx, "failed to pre-warm cache on creation",
					slog.String("short_code", shortCode),
					slog.String("error", cacheErr.Error()),
				)
			}
			return CreateResult{
				ShortCode: shortCode,
				Cache:     "miss",
			}, nil
		}

		if errors.Is(err, repository.ErrConflict) {
			continue
		}

		return CreateResult{}, fmt.Errorf("create short url in repo: %w", err)
	}

	return CreateResult{}, ErrConflict
}

func (s *Service) CreateShortURL(ctx context.Context, originalURL string) (string, error) {
	res, err := s.CreateShortURLWithMetadata(ctx, originalURL)
	return res.ShortCode, err
}

func (s *Service) ResolveURL(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("%w: short code cannot be empty", ErrInvalidInput)
	}

	key := s.cacheKey(shortCode)

	// 1. Check cache first (fast path)
	cachedURL, err := s.cache.Get(ctx, key)
	if err == nil && cachedURL != "" {
		return cachedURL, nil
	}

	if err != nil && !errors.Is(err, cache.ErrCacheMiss) {
		slog.WarnContext(ctx, "cache get failed, falling back to database",
			slog.String("short_code", shortCode),
			slog.String("error", err.Error()),
		)
	}

	// 2. Fallback to repository / database
	originalURL, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("resolve url in repo: %w", err)
	}

	// 3. Populate cache on cache miss (fail-open)
	if setErr := s.cache.Set(ctx, key, originalURL, s.cacheTTL); setErr != nil {
		slog.WarnContext(ctx, "failed to populate cache after miss",
			slog.String("short_code", shortCode),
			slog.String("error", setErr.Error()),
		)
	}

	return originalURL, nil
}
