package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"github.com/ndquang191/Anochat/api/pkg/database"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}
	if err := database.InitDatabase(cfg); err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer database.CloseDatabase()

	if err := database.RunMigrations(); err != nil {
		return err
	}
	return nil
}
