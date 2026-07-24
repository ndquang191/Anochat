package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	Env                string
	ClientURL          string
	AllowDatabaseReset bool
	DevAuthEnabled     bool
	Database           DatabaseConfig
	OAuth              OAuthConfig
	Chat               ChatConfig
	User               UserConfig
	Security           SecurityConfig
	Redis              RedisConfig
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
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

type ChatConfig struct {
	MessageRateLimit int
	MaxMessageLength int
}

type UserConfig struct {
	MinAge int
	MaxAge int
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

// Load reads local development values from .env when present, then builds the
// configuration from environment variables. Existing environment variables
// always win over values in .env.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("load .env: %w", err)
	}

	accessTokenExpiry, err := durationEnv("OAUTH_ACCESS_TOKEN_EXPIRY", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTokenExpiry, err := durationEnv("OAUTH_REFRESH_TOKEN_EXPIRY", 30*24*time.Hour)
	if err != nil {
		return nil, err
	}
	messageRateLimit, err := intEnv("CHAT_MESSAGE_RATE_LIMIT", 10)
	if err != nil {
		return nil, err
	}
	maxMessageLength, err := intEnv("CHAT_MAX_MESSAGE_LENGTH", 1000)
	if err != nil {
		return nil, err
	}
	minAge, err := intEnv("USER_MIN_AGE", 10)
	if err != nil {
		return nil, err
	}
	maxAge, err := intEnv("USER_MAX_AGE", 99)
	if err != nil {
		return nil, err
	}
	rateLimit, err := intEnv("SECURITY_RATE_LIMIT", 100)
	if err != nil {
		return nil, err
	}
	redisDB, err := intEnv("REDIS_DB", 0)
	if err != nil {
		return nil, err
	}
	allowDatabaseReset, err := boolEnv("ALLOW_DATABASE_RESET", false)
	if err != nil {
		return nil, err
	}
	devAuthEnabled, err := boolEnv("DEV_AUTH_ENABLED", false)
	if err != nil {
		return nil, err
	}

	clientURL := envOrDefault("SERVER_CLIENT_URL", "http://localhost:3000")
	cfg := &Config{
		Port:               envOrDefault("SERVER_PORT", "8080"),
		Env:                strings.ToLower(envOrDefault("SERVER_ENV", "development")),
		ClientURL:          clientURL,
		AllowDatabaseReset: allowDatabaseReset,
		DevAuthEnabled:     devAuthEnabled,
		Database: DatabaseConfig{
			Host:     envOrDefault("DATABASE_HOST", "localhost"),
			Port:     envOrDefault("DATABASE_PORT", "5432"),
			User:     os.Getenv("DATABASE_USER"),
			Password: os.Getenv("DATABASE_PASSWORD"),
			Name:     envOrDefault("DATABASE_NAME", "anochat"),
			SSLMode:  envOrDefault("DATABASE_SSL_MODE", "disable"),
		},
		OAuth: OAuthConfig{
			GoogleClientID:     os.Getenv("OAUTH_GOOGLE_CLIENT_ID"),
			GoogleClientSecret: os.Getenv("OAUTH_GOOGLE_CLIENT_SECRET"),
			RedirectURL:        envOrDefault("OAUTH_REDIRECT_URL", "http://localhost:8080/auth/callback"),
			JWTSecret:          os.Getenv("OAUTH_JWT_SECRET"),
			AccessTokenExpiry:  accessTokenExpiry,
			RefreshTokenExpiry: refreshTokenExpiry,
		},
		Chat: ChatConfig{
			MessageRateLimit: messageRateLimit,
			MaxMessageLength: maxMessageLength,
		},
		User: UserConfig{
			MinAge: minAge,
			MaxAge: maxAge,
		},
		Security: SecurityConfig{
			CORSOrigins: []string{clientURL},
			RateLimit:   rateLimit,
		},
		Redis: RedisConfig{
			URL:      envOrDefault("REDIS_URL", "localhost:6379"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       redisDB,
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func intEnv(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}
	return parsed, nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be true or false: %w", key, err)
	}
	return parsed, nil
}

func durationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a duration such as 15m or 720h: %w", key, err)
	}
	return parsed, nil
}

func (c *Config) validate() error {
	if c.Env != "development" && c.Env != "test" && c.Env != "production" {
		return fmt.Errorf("SERVER_ENV must be development, test, or production")
	}
	if c.DevAuthEnabled && !c.IsDevelopment() {
		return fmt.Errorf("DEV_AUTH_ENABLED can only be enabled in development")
	}
	if c.Database.User == "" {
		return fmt.Errorf("DATABASE_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DATABASE_PASSWORD is required")
	}
	if c.OAuth.JWTSecret == "" {
		return fmt.Errorf("OAUTH_JWT_SECRET is required")
	}
	if c.Chat.MessageRateLimit <= 0 {
		return fmt.Errorf("CHAT_MESSAGE_RATE_LIMIT must be greater than zero")
	}
	if c.Chat.MaxMessageLength <= 0 {
		return fmt.Errorf("CHAT_MAX_MESSAGE_LENGTH must be greater than zero")
	}
	if c.User.MinAge <= 0 || c.User.MaxAge < c.User.MinAge {
		return fmt.Errorf("USER_MIN_AGE and USER_MAX_AGE define an invalid range")
	}
	if c.Security.RateLimit <= 0 {
		return fmt.Errorf("SECURITY_RATE_LIMIT must be greater than zero")
	}
	if c.Redis.DB < 0 {
		return fmt.Errorf("REDIS_DB must be zero or greater")
	}
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
