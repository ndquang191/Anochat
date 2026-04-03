package initialize

import (
	"log/slog"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

func Logger(cfg *config.Config) *zap.Logger {
	var zapLogger *zap.Logger
	if cfg.IsProduction() {
		zapLogger, _ = zap.NewProduction()
	} else {
		zapLogger, _ = zap.NewDevelopment()
	}

	slog.SetDefault(slog.New(zapslog.NewHandler(zapLogger.Core())))
	slog.Info("Logger initialized")

	return zapLogger
}
