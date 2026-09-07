package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ashwanijha1405/url-shortener/internal/cache"
	"github.com/Ashwanijha1405/url-shortener/internal/cache/redis"
	"github.com/Ashwanijha1405/url-shortener/internal/config"
	"github.com/Ashwanijha1405/url-shortener/internal/database"
	"github.com/Ashwanijha1405/url-shortener/internal/handler"
	"github.com/Ashwanijha1405/url-shortener/internal/middleware"
	"github.com/Ashwanijha1405/url-shortener/internal/repository/postgres"
	"github.com/Ashwanijha1405/url-shortener/internal/service"
)

func initLogger(cfg config.LogConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
	return logger
}

func main() {
	cfg := config.Load()
	logger := initLogger(cfg.Log)

	ctx := context.Background()

	// Connect to PostgreSQL
	db, err := database.Connect(ctx, cfg.DB)
	if err != nil {
		logger.Error("database connection failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Pool.Close()

	logger.Info("connected to PostgreSQL",
		slog.Int("max_conns", int(cfg.DB.MaxConns)),
		slog.Int("min_conns", int(cfg.DB.MinConns)),
	)

	// Initialize Cache Layer
	var cacheClient cache.Cache = cache.NewNoopCache()
	var healthOpts []handler.HealthOption

	if cfg.Redis.Enabled {
		rClient, err := redis.New(cfg.Redis)
		if err != nil {
			logger.Warn("failed to connect to Redis cache, falling back to NoopCache", slog.String("error", err.Error()))
		} else {
			defer rClient.Close()
			cacheClient = rClient
			healthOpts = append(healthOpts, handler.WithCacheChecker(rClient))
			logger.Info("connected to Redis cache",
				slog.String("addr", cfg.Redis.Addr),
				slog.Duration("ttl", cfg.Redis.TTL),
			)
		}
	} else {
		logger.Info("Redis cache disabled; running in uncached mode")
	}

	// Initialize Rate Limiter
	var rateLimiter *middleware.IPRateLimiter
	if cfg.RateLimit.Enabled {
		rateLimiter = middleware.NewIPRateLimiter(cfg.RateLimit.RPS, cfg.RateLimit.Burst)
		defer rateLimiter.Close()
		logger.Info("client IP rate limiting enabled",
			slog.Float64("rps", cfg.RateLimit.RPS),
			slog.Int("burst", cfg.RateLimit.Burst),
		)
	}

	// Initialize repository and service
	repo := postgres.NewRepository(db)
	svc := service.New(
		repo,
		service.WithCache(cacheClient, cfg.Redis.TTL),
	)

	// Initialize HTTP handlers
	urlHandler := handler.NewHandler(svc, handler.WithMaxBodyBytes(cfg.Server.MaxBodyBytes))
	healthHandler := handler.NewHealthHandler(db, healthOpts...)

	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /health/live", healthHandler.Live)
	mux.HandleFunc("GET /health/ready", healthHandler.Ready)

	// Core API endpoints
	if rateLimiter != nil {
		mux.Handle("POST /api/v1/urls", middleware.RateLimit(rateLimiter, logger)(http.HandlerFunc(urlHandler.CreateURL)))
	} else {
		mux.HandleFunc("POST /api/v1/urls", urlHandler.CreateURL)
	}
	mux.HandleFunc("GET /{shortCode}", urlHandler.RedirectURL)

	// Middleware pipeline: RequestID -> Logger -> Recoverer -> Mux
	var httpHandler http.Handler = mux
	httpHandler = middleware.Recoverer(logger)(httpHandler)
	httpHandler = middleware.Logger(logger)(httpHandler)
	httpHandler = middleware.RequestID(httpHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: httpHandler,
	}

	// Start server in a separate goroutine.
	go func() {
		logger.Info("starting HTTP server", slog.String("port", cfg.Server.Port))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for termination signal.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	<-signalChan

	logger.Info("shutdown signal received")

	// Give active requests some time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.Server.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", slog.String("error", err.Error()))
		return
	}

	logger.Info("server stopped gracefully")
}
