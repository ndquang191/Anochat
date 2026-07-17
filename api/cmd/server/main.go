package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndquang191/Anochat/api/internal/initialize"
	"github.com/ndquang191/Anochat/api/pkg/cache"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	zapLogger := initialize.Logger(cfg)
	defer zapLogger.Sync()

	db, err := initialize.Database(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.CloseDatabase()

	redisClient, err := initialize.Redis(cfg)
	if err != nil {
		slog.Error("Failed to initialize Redis", "error", err)
		os.Exit(1)
	}
	defer cache.CloseRedis()

	metrics.Register()

	server := initialize.Router(cfg, db, redisClient)

	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		if err := server.Router.Run(":" + cfg.Port); err != nil {
			slog.Error("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")
	server.QueueService.Stop()
}
