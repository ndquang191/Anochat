package config

import (
	"testing"
	"time"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"SERVER_PORT", "SERVER_ENV", "SERVER_CLIENT_URL",
		"DATABASE_HOST", "DATABASE_PORT", "DATABASE_NAME", "DATABASE_SSL_MODE",
		"REDIS_URL", "REDIS_PASSWORD", "REDIS_DB",
		"OAUTH_GOOGLE_CLIENT_ID", "OAUTH_GOOGLE_CLIENT_SECRET", "OAUTH_REDIRECT_URL",
		"OAUTH_ACCESS_TOKEN_EXPIRY", "OAUTH_REFRESH_TOKEN_EXPIRY",
		"CHAT_MESSAGE_RATE_LIMIT", "CHAT_MAX_MESSAGE_LENGTH",
		"USER_MIN_AGE", "USER_MAX_AGE", "SECURITY_RATE_LIMIT", "ALLOW_DATABASE_RESET",
		"DEV_AUTH_ENABLED",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("DATABASE_USER", "postgres")
	t.Setenv("DATABASE_PASSWORD", "postgres")
	t.Setenv("OAUTH_JWT_SECRET", "test-secret")
}

func TestLoadUsesDefaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("Env = %q, want development", cfg.Env)
	}
	if cfg.AllowDatabaseReset {
		t.Error("AllowDatabaseReset = true, want false")
	}
	if cfg.Database.Host != "localhost" || cfg.Database.Port != "5432" {
		t.Errorf("Database address = %s:%s, want localhost:5432", cfg.Database.Host, cfg.Database.Port)
	}
	if cfg.OAuth.AccessTokenExpiry != 15*time.Minute {
		t.Errorf("AccessTokenExpiry = %s, want 15m", cfg.OAuth.AccessTokenExpiry)
	}
	if cfg.OAuth.RefreshTokenExpiry != 30*24*time.Hour {
		t.Errorf("RefreshTokenExpiry = %s, want 720h", cfg.OAuth.RefreshTokenExpiry)
	}
}

func TestLoadUsesEnvironmentOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_ENV", "production")
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("DATABASE_HOST", "db.internal")
	t.Setenv("REDIS_DB", "2")
	t.Setenv("CHAT_MAX_MESSAGE_LENGTH", "2048")
	t.Setenv("ALLOW_DATABASE_RESET", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "production" || cfg.Port != "9000" {
		t.Errorf("server = %s:%s, want production:9000", cfg.Env, cfg.Port)
	}
	if cfg.Database.Host != "db.internal" {
		t.Errorf("Database.Host = %q, want db.internal", cfg.Database.Host)
	}
	if cfg.Redis.DB != 2 {
		t.Errorf("Redis.DB = %d, want 2", cfg.Redis.DB)
	}
	if cfg.Chat.MaxMessageLength != 2048 {
		t.Errorf("MaxMessageLength = %d, want 2048", cfg.Chat.MaxMessageLength)
	}
	if !cfg.AllowDatabaseReset {
		t.Error("AllowDatabaseReset = false, want true")
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("OAUTH_ACCESS_TOKEN_EXPIRY", "later")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid duration error")
	}
}

func TestLoadRejectsInvalidBoolean(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("ALLOW_DATABASE_RESET", "sometimes")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid boolean error")
	}
}

func TestLoadRejectsDevAuthOutsideDevelopment(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("SERVER_ENV", "production")
	t.Setenv("DEV_AUTH_ENABLED", "true")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want dev auth environment error")
	}
}
