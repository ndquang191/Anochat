package config

import (
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Port      string
	Env       string
	ClientURL string
	Database  DatabaseConfig
	OAuth     OAuthConfig
	Chat      ChatConfig
	User      UserConfig
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

func Load() *Config {
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	// Base config (committed)
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("Failed to read config.yaml", "error", err)
			os.Exit(1)
		}
		slog.Warn("config.yaml not found")
	} else {
		slog.Info("Loaded config", "file", v.ConfigFileUsed())
	}

	// Local overrides (gitignored — secrets + local env values)
	v.SetConfigName("config.local")
	if err := v.MergeInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("Failed to read config.local.yaml", "error", err)
			os.Exit(1)
		}
		// No local file is fine in Docker/production
	} else {
		slog.Info("Merged local config", "file", v.ConfigFileUsed())
	}

	// Docker / production: allow env var overrides for a handful of fields
	// that differ between environments (hostnames, secrets injected at runtime).
	// Env var names mirror the YAML path with "." → "_" and uppercased.
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	cfg := &Config{
		Port:      v.GetString("server.port"),
		Env:       v.GetString("server.env"),
		ClientURL: v.GetString("server.client_url"),

		Database: DatabaseConfig{
			Host:     v.GetString("database.host"),
			Port:     v.GetString("database.port"),
			User:     v.GetString("database.user"),
			Password: v.GetString("database.password"),
			Name:     v.GetString("database.name"),
			SSLMode:  v.GetString("database.ssl_mode"),
		},

		OAuth: OAuthConfig{
			GoogleClientID:     v.GetString("oauth.google_client_id"),
			GoogleClientSecret: v.GetString("oauth.google_client_secret"),
			RedirectURL:        v.GetString("oauth.redirect_url"),
			JWTSecret:          v.GetString("oauth.jwt_secret"),
			AccessTokenExpiry:  v.GetDuration("oauth.access_token_expiry"),
			RefreshTokenExpiry: v.GetDuration("oauth.refresh_token_expiry"),
		},

		Chat: ChatConfig{
			MessageRateLimit: v.GetInt("chat.message_rate_limit"),
			MaxMessageLength: v.GetInt("chat.max_message_length"),
		},

		User: UserConfig{
			MinAge: v.GetInt("user.min_age"),
			MaxAge: v.GetInt("user.max_age"),
		},

		Security: SecurityConfig{
			CORSOrigins: []string{v.GetString("server.client_url")},
			RateLimit:   v.GetInt("security.rate_limit"),
		},

		Redis: RedisConfig{
			URL:      v.GetString("redis.url"),
			Password: v.GetString("redis.password"),
			DB:       v.GetInt("redis.db"),
		},
	}

	cfg.validate()
	return cfg
}

func (c *Config) validate() {
	if c.Database.Host == "" {
		slog.Error("database.host is required")
		os.Exit(1)
	}
	if c.Database.User == "" {
		slog.Error("database.user is required")
		os.Exit(1)
	}
	if c.Database.Password == "" {
		slog.Error("database.password is required")
		os.Exit(1)
	}
	if c.Database.Name == "" {
		slog.Error("database.name is required")
		os.Exit(1)
	}
	if c.OAuth.JWTSecret == "" {
		slog.Error("oauth.jwt_secret is required")
		os.Exit(1)
	}
	if c.OAuth.GoogleClientID == "" {
		slog.Warn("oauth.google_client_id not set — OAuth will be disabled")
	}
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}
