package initialize

import (
	"log/slog"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"
	"gorm.io/gorm"
)

func Database(cfg *config.Config) (*gorm.DB, error) {
	if err := database.InitDatabase(cfg); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		return nil, err
	}

	if err := database.RunMigrations(); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		database.CloseDatabase()
		return nil, err
	}

	return database.GetDB(), nil
}
