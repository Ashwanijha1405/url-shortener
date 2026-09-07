package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Redis     RedisConfig
	RateLimit RateLimitConfig
	Log       LogConfig
}

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
}

type DBConfig struct {
	DatabaseURL     string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
	Enabled  bool
}

type RateLimitConfig struct {
	RPS     float64
	Burst   int
	Enabled bool
}

type LogConfig struct {
	Level  string
	Format string
}

func Load() *Config {
	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "")

	maxConns := int32(getEnvAsInt("DB_MAX_CONNS", 10))
	minConns := int32(getEnvAsInt("DB_MIN_CONNS", 2))
	maxConnLifetime := getEnvAsDuration("DB_MAX_CONN_LIFETIME", time.Hour)
	shutdownTimeout := getEnvAsDuration("SERVER_SHUTDOWN_TIMEOUT", 5*time.Second)
	maxBodyBytes := int64(getEnvAsInt("SERVER_MAX_BODY_BYTES", 10240))

	redisAddr := getEnv("REDIS_ADDR", "")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnvAsInt("REDIS_DB", 0)
	redisTTL := getEnvAsDuration("REDIS_TTL", 24*time.Hour)
	redisEnabledStr := getEnv("REDIS_ENABLED", "")
	redisEnabled := false
	if redisEnabledStr != "" {
		redisEnabled = strings.ToLower(redisEnabledStr) == "true" || redisEnabledStr == "1"
	} else if redisAddr != "" {
		redisEnabled = true
	}

	rateLimitRPS := getEnvAsFloat("RATE_LIMIT_RPS", 5.0)
	rateLimitBurst := getEnvAsInt("RATE_LIMIT_BURST", 10)
	rateLimitEnabledStr := getEnv("RATE_LIMIT_ENABLED", "true")
	rateLimitEnabled := strings.ToLower(rateLimitEnabledStr) == "true" || rateLimitEnabledStr == "1"

	logLevel := strings.ToUpper(getEnv("LOG_LEVEL", "INFO"))
	logFormat := strings.ToLower(getEnv("LOG_FORMAT", "json"))

	return &Config{
		Server: ServerConfig{
			Port:            port,
			ShutdownTimeout: shutdownTimeout,
			MaxBodyBytes:    maxBodyBytes,
		},
		DB: DBConfig{
			DatabaseURL:     dbURL,
			MaxConns:        maxConns,
			MinConns:        minConns,
			MaxConnLifetime: maxConnLifetime,
		},
		Redis: RedisConfig{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
			TTL:      redisTTL,
			Enabled:  redisEnabled,
		},
		RateLimit: RateLimitConfig{
			RPS:     rateLimitRPS,
			Burst:   rateLimitBurst,
			Enabled: rateLimitEnabled,
		},
		Log: LogConfig{
			Level:  logLevel,
			Format: logFormat,
		},
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvAsFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}
