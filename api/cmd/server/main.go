package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ndquang191/Anochat/api/internal/initialize"
	"github.com/ndquang191/Anochat/api/pkg/cache"
	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"
	"github.com/ndquang191/Anochat/api/pkg/metrics"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Server stopped with an error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	zapLogger := initialize.Logger(cfg)
	defer zapLogger.Sync()

	db, err := initialize.Database(cfg)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer database.CloseDatabase()

	redisClient, err := initialize.Redis(cfg)
	if err != nil {
		return fmt.Errorf("initialize Redis: %w", err)
	}
	defer cache.CloseRedis()

	metrics.Register()

	server := initialize.Router(cfg, db, redisClient)
	appCtx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	server.Start(appCtx)
	slog.Info("WebSocket hub and queue reconciler started")

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           server.Router,
		ReadHeaderTimeout: 10 * time.Second,
	}
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("Starting server", "port", cfg.Port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	var serveErr error
	select {
	case sig := <-quit:
		slog.Info("Shutdown signal received", "signal", sig.String())
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server stopped unexpectedly", "error", err)
			serveErr = err
		}
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("HTTP server graceful shutdown failed", "error", err)
	}
	stopBackground()
	if err := server.Wait(shutdownCtx); err != nil {
		slog.Error("Background services did not stop cleanly", "error", err)
	}
	slog.Info("Server shutdown complete")
	return serveErr
}
