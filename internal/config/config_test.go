package config

import "testing"

func TestLoadNormalizesLegacyDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql+asyncpg://user:pass@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SECRET_KEY", "01234567890123456789012345678901")
	t.Setenv("APP_ENV", "production")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgresql://user:pass@localhost:5432/db" {
		t.Fatalf("unexpected database URL: %s", cfg.DatabaseURL)
	}
}

func TestProductionRejectsShortSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("SECRET_KEY", "short")
	t.Setenv("APP_ENV", "production")
	if _, err := Load(); err == nil {
		t.Fatal("expected short production secret to be rejected")
	}
}
