package config

import (
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port string
	Env  string

	// Database configuration
	Database DatabaseConfig

	// OAuth configuration
	OAuth OAuthConfig

	// Chat configuration
	Chat ChatConfig

	// Security configuration
	Security SecurityConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// OAuthConfig holds OAuth2 settings
type OAuthConfig struct {
	GoogleClientID     string
	GoogleClientSecret string
	RedirectURL        string
	JWTSecret          string
	JWTExpiration      time.Duration
}

// ChatConfig holds chat-related settings
type ChatConfig struct {
	QueueTimeout     time.Duration
	MessageRateLimit int
	MaxMessageLength int
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	CORSOrigins []string
	RateLimit   int
}

// Load loads configuration from environment variables
func Load() *Config {
	config := &Config{
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("ENV", "development"),

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
			QueueTimeout:     getEnvAsDuration("QUEUE_TIMEOUT", 30*time.Second),
			MessageRateLimit: getEnvAsInt("MESSAGE_RATE_LIMIT", 10),
			MaxMessageLength: getEnvAsInt("MAX_MESSAGE_LENGTH", 1000),
		},

		Security: SecurityConfig{
			CORSOrigins: getEnvAsSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
			RateLimit:   getEnvAsInt("RATE_LIMIT", 100),
		},
	}

	// Validate required configuration
	config.validate()

	return config
}

// validate checks if required configuration is present
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

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// Helper functions for environment variable parsing

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
		// Simple comma-separated values parsing
		// For more complex cases, consider using a proper CSV parser
		return []string{value} // For now, just return as single item
	}
	return defaultValue
} 