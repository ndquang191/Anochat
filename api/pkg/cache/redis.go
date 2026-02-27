package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func InitRedis(cfg *config.Config) error {
	Client = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URL,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		Client = nil
		return err
	}

	slog.Info("Redis connection established", "addr", cfg.Redis.URL)
	return nil
}

func CloseRedis() {
	if Client != nil {
		if err := Client.Close(); err != nil {
			slog.Error("Failed to close Redis connection", "error", err)
		} else {
			slog.Info("Redis connection closed")
		}
	}
}

func HealthCheck() error {
	if Client == nil {
		return redis.ErrClosed
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return Client.Ping(ctx).Err()
}
