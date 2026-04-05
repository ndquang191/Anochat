package initialize

import (
	"log/slog"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

func Logger(cfg *config.Config) *zap.Logger {
	var (
		zapLogger *zap.Logger
		err       error
	)
	if cfg.IsProduction() {
		zapLogger, err = zap.NewProduction()
	} else {
		zapLogger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	slog.SetDefault(slog.New(zapslog.NewHandler(zapLogger.Core())))
	slog.Info("Logger initialized")

	return zapLogger
}
