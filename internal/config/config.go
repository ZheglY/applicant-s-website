package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment               string
	HTTPAddr                  string
	DatabaseURL               string
	RedisURL                  string
	RedisKeyPrefix            string
	SecretKey                 string
	AccessTokenTTL            time.Duration
	CookieSecure              bool
	WebDir                    string
	ShutdownTimeout           time.Duration
	DirectionsCacheTTL        time.Duration
	NewsCacheTTL              time.Duration
	SummaryCacheTTL           time.Duration
	LoginRateLimit            int
	LoginRateWindow           time.Duration
	RegistrationRateLimit     int
	RegistrationRateWindow    time.Duration
	RateLimitFailOpen         bool
	DefaultAdmissionsLogin    string
	DefaultAdmissionsPassword string
	DefaultAnalystLogin       string
	DefaultAnalystPassword    string
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg := Config{
		Environment:               env("APP_ENV", "development"),
		HTTPAddr:                  env("HTTP_ADDR", ":8000"),
		DatabaseURL:               normalizeDatabaseURL(env("DATABASE_URL", "postgresql://admin:123@localhost:5433/unik")),
		RedisURL:                  env("REDIS_URL", "redis://localhost:6379/0"),
		RedisKeyPrefix:            env("REDIS_KEY_PREFIX", "unik:v1:"),
		SecretKey:                 env("SECRET_KEY", "dev-secret-change-me"),
		CookieSecure:              envBool("COOKIE_SECURE", false),
		WebDir:                    env("WEB_DIR", "app"),
		LoginRateLimit:            envInt("LOGIN_RATE_LIMIT", 10),
		RegistrationRateLimit:     envInt("REGISTRATION_RATE_LIMIT", 5),
		RateLimitFailOpen:         envBool("RATE_LIMIT_FAIL_OPEN", true),
		DefaultAdmissionsLogin:    env("DEFAULT_ADMISSIONS_LOGIN", "admin@unik.edu"),
		DefaultAdmissionsPassword: env("DEFAULT_ADMISSIONS_PASSWORD", "admin"),
		DefaultAnalystLogin:       env("DEFAULT_ANALYST_LOGIN", "prepod@unik.edu"),
		DefaultAnalystPassword:    env("DEFAULT_ANALYST_PASSWORD", "123456"),
	}

	var err error
	if cfg.AccessTokenTTL, err = envDuration("ACCESS_TOKEN_TTL", 24*time.Hour); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = envDuration("SHUTDOWN_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.DirectionsCacheTTL, err = envDuration("DIRECTIONS_CACHE_TTL", 15*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.NewsCacheTTL, err = envDuration("NEWS_CACHE_TTL", 2*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.SummaryCacheTTL, err = envDuration("SUMMARY_CACHE_TTL", time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.LoginRateWindow, err = envDuration("LOGIN_RATE_WINDOW", 5*time.Minute); err != nil {
		return Config{}, err
	}
	if cfg.RegistrationRateWindow, err = envDuration("REGISTRATION_RATE_WINDOW", 10*time.Minute); err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.RedisURL == "" {
		return Config{}, errors.New("REDIS_URL is required")
	}
	if cfg.LoginRateLimit <= 0 || cfg.RegistrationRateLimit <= 0 {
		return Config{}, errors.New("rate limits must be positive")
	}
	if len(cfg.SecretKey) < 32 && cfg.Environment == "production" {
		return Config{}, errors.New("SECRET_KEY must contain at least 32 characters in production")
	}
	return cfg, nil
}

// loadDotEnv provides a small, dependency-free .env reader. Existing process
// environment variables always take precedence.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(strings.TrimPrefix(key, "export "))
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if key != "" {
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
	}
	return scanner.Err()
}

func env(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return parsed, nil
}

func normalizeDatabaseURL(value string) string {
	return strings.Replace(value, "postgresql+asyncpg://", "postgresql://", 1)
}
