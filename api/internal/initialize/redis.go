package initialize

import (
	"log/slog"

	"github.com/ndquang191/Anochat/api/pkg/cache"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/redis/go-redis/v9"
)

func Redis(cfg *config.Config) (*redis.Client, error) {
	if err := cache.InitRedis(cfg); err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		return nil, err
	}

	return cache.Client, nil
}
