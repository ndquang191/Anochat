package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port      string
	Env       string
	ClientURL string
	Database  DatabaseConfig
	OAuth     OAuthConfig
	Chat      ChatConfig
	Security  SecurityConfig
	Redis     RedisConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
	JWTSecret          string
	JWTExpiration      time.Duration
}

type ChatConfig struct {
	QueueTimeout         time.Duration
	QueueHeartbeatTTL    time.Duration
	QueueCleanupInterval time.Duration
	MessageRateLimit     int
	MaxMessageLength     int
}

type SecurityConfig struct {
	CORSOrigins []string
	RateLimit   int
}

type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

func Load() *Config {
	config := &Config{
		Port:      getEnv("PORT", "8080"),
		Env:       getEnv("ENV", "development"),
		ClientURL: getEnv("CLIENT_URL", "http://localhost:3000"),

		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", ""),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", ""),
			SSLMode:  getEnv("DB_SSLMODE", "require"),
		},

		OAuth: OAuthConfig{
			GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:        getEnv("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/callback"),
			JWTSecret:          getEnv("JWT_SECRET", ""),
			JWTExpiration:      getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},

		Chat: ChatConfig{
			QueueTimeout:         getEnvAsDuration("QUEUE_TIMEOUT", 30*time.Second),
			QueueHeartbeatTTL:    getEnvAsDuration("QUEUE_HEARTBEAT_TTL", 10*time.Second),
			QueueCleanupInterval: getEnvAsDuration("QUEUE_CLEANUP_INTERVAL", 30*time.Second),
			MessageRateLimit:     getEnvAsInt("MESSAGE_RATE_LIMIT", 10),
			MaxMessageLength:     getEnvAsInt("MAX_MESSAGE_LENGTH", 1000),
		},

		Security: SecurityConfig{
			CORSOrigins: getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
			RateLimit:   getEnvAsInt("RATE_LIMIT", 100),
		},

		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
	}

	config.validate()

	return config
}

func (c *Config) validate() {
	if c.Database.Host == "" {
		slog.Error("DB_HOST is required")
		os.Exit(1)
	}

	if c.Database.User == "" {
		slog.Error("DB_USER is required")
		os.Exit(1)
	}

	if c.Database.Password == "" {
		slog.Error("DB_PASSWORD is required")
		os.Exit(1)
	}

	if c.Database.Name == "" {
		slog.Error("DB_NAME is required")
		os.Exit(1)
	}

	if c.OAuth.GoogleClientID == "" {
		slog.Warn("GOOGLE_CLIENT_ID not set - OAuth will be disabled")
	}

	if c.OAuth.JWTSecret == "" {
		slog.Error("JWT_SECRET is required")
		os.Exit(1)
	}
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		slog.Warn("Invalid integer value for environment variable", "key", key, "value", value)
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		slog.Warn("Invalid duration value for environment variable", "key", key, "value", value)
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return []string{value}
	}
	return defaultValue
}
