package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_MAX_CONNS", "")
	t.Setenv("DB_MIN_CONNS", "")
	t.Setenv("DB_MAX_CONN_LIFETIME", "")
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "")
	t.Setenv("SERVER_MAX_BODY_BYTES", "")
	t.Setenv("REDIS_ADDR", "")
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_DB", "")
	t.Setenv("REDIS_TTL", "")
	t.Setenv("REDIS_ENABLED", "")
	t.Setenv("RATE_LIMIT_RPS", "")
	t.Setenv("RATE_LIMIT_BURST", "")
	t.Setenv("RATE_LIMIT_ENABLED", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("LOG_FORMAT", "")

	cfg := Load()

	if cfg.Server.Port != "8080" {
		t.Errorf("expected default Port '8080', got %q", cfg.Server.Port)
	}
	if cfg.Server.ShutdownTimeout != 5*time.Second {
		t.Errorf("expected default ShutdownTimeout 5s, got %v", cfg.Server.ShutdownTimeout)
	}
	if cfg.Server.MaxBodyBytes != 10240 {
		t.Errorf("expected default MaxBodyBytes 10240, got %d", cfg.Server.MaxBodyBytes)
	}
	if cfg.DB.DatabaseURL != "" {
		t.Errorf("expected empty default DatabaseURL, got %q", cfg.DB.DatabaseURL)
	}
	if cfg.DB.MaxConns != 10 {
		t.Errorf("expected default MaxConns 10, got %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 2 {
		t.Errorf("expected default MinConns 2, got %d", cfg.DB.MinConns)
	}
	if cfg.DB.MaxConnLifetime != time.Hour {
		t.Errorf("expected default MaxConnLifetime 1h, got %v", cfg.DB.MaxConnLifetime)
	}
	if cfg.Redis.Addr != "" {
		t.Errorf("expected empty default Redis.Addr, got %q", cfg.Redis.Addr)
	}
	if cfg.Redis.TTL != 24*time.Hour {
		t.Errorf("expected default Redis.TTL 24h, got %v", cfg.Redis.TTL)
	}
	if cfg.Redis.Enabled != false {
		t.Errorf("expected default Redis.Enabled false, got %v", cfg.Redis.Enabled)
	}
	if cfg.RateLimit.RPS != 5.0 {
		t.Errorf("expected default RateLimit.RPS 5.0, got %f", cfg.RateLimit.RPS)
	}
	if cfg.RateLimit.Burst != 10 {
		t.Errorf("expected default RateLimit.Burst 10, got %d", cfg.RateLimit.Burst)
	}
	if cfg.RateLimit.Enabled != true {
		t.Errorf("expected default RateLimit.Enabled true, got %v", cfg.RateLimit.Enabled)
	}
	if cfg.Log.Level != "INFO" {
		t.Errorf("expected default Log.Level 'INFO', got %q", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("expected default Log.Format 'json', got %q", cfg.Log.Format)
	}
}

func TestLoadCustomEnv(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("DB_MAX_CONNS", "25")
	t.Setenv("DB_MIN_CONNS", "5")
	t.Setenv("DB_MAX_CONN_LIFETIME", "30m")
	t.Setenv("SERVER_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("SERVER_MAX_BODY_BYTES", "20480")
	t.Setenv("REDIS_ADDR", "localhost:6379")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DB", "1")
	t.Setenv("REDIS_TTL", "12h")
	t.Setenv("REDIS_ENABLED", "true")
	t.Setenv("RATE_LIMIT_RPS", "20.5")
	t.Setenv("RATE_LIMIT_BURST", "50")
	t.Setenv("RATE_LIMIT_ENABLED", "false")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "TEXT")

	cfg := Load()

	if cfg.Server.Port != "9090" {
		t.Errorf("expected Port '9090', got %q", cfg.Server.Port)
	}
	if cfg.Server.ShutdownTimeout != 10*time.Second {
		t.Errorf("expected ShutdownTimeout 10s, got %v", cfg.Server.ShutdownTimeout)
	}
	if cfg.Server.MaxBodyBytes != 20480 {
		t.Errorf("expected MaxBodyBytes 20480, got %d", cfg.Server.MaxBodyBytes)
	}
	if cfg.DB.DatabaseURL != "postgres://user:pass@localhost:5432/testdb" {
		t.Errorf("expected DatabaseURL 'postgres://user:pass@localhost:5432/testdb', got %q", cfg.DB.DatabaseURL)
	}
	if cfg.DB.MaxConns != 25 {
		t.Errorf("expected MaxConns 25, got %d", cfg.DB.MaxConns)
	}
	if cfg.DB.MinConns != 5 {
		t.Errorf("expected MinConns 5, got %d", cfg.DB.MinConns)
	}
	if cfg.DB.MaxConnLifetime != 30*time.Minute {
		t.Errorf("expected MaxConnLifetime 30m, got %v", cfg.DB.MaxConnLifetime)
	}
	if cfg.Redis.Addr != "localhost:6379" {
		t.Errorf("expected Redis.Addr 'localhost:6379', got %q", cfg.Redis.Addr)
	}
	if cfg.Redis.Password != "secret" {
		t.Errorf("expected Redis.Password 'secret', got %q", cfg.Redis.Password)
	}
	if cfg.Redis.DB != 1 {
		t.Errorf("expected Redis.DB 1, got %d", cfg.Redis.DB)
	}
	if cfg.Redis.TTL != 12*time.Hour {
		t.Errorf("expected Redis.TTL 12h, got %v", cfg.Redis.TTL)
	}
	if cfg.Redis.Enabled != true {
		t.Errorf("expected Redis.Enabled true, got %v", cfg.Redis.Enabled)
	}
	if cfg.RateLimit.RPS != 20.5 {
		t.Errorf("expected RateLimit.RPS 20.5, got %f", cfg.RateLimit.RPS)
	}
	if cfg.RateLimit.Burst != 50 {
		t.Errorf("expected RateLimit.Burst 50, got %d", cfg.RateLimit.Burst)
	}
	if cfg.RateLimit.Enabled != false {
		t.Errorf("expected RateLimit.Enabled false, got %v", cfg.RateLimit.Enabled)
	}
	if cfg.Log.Level != "DEBUG" {
		t.Errorf("expected Log.Level 'DEBUG', got %q", cfg.Log.Level)
	}
	if cfg.Log.Format != "text" {
		t.Errorf("expected Log.Format 'text', got %q", cfg.Log.Format)
	}
}
