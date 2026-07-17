package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ndquang191/Anochat/api/pkg/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	if !cfg.IsDevelopment() {
		slog.Error("Database reset is only allowed in development", "environment", cfg.Env)
		os.Exit(1)
	}
	if !cfg.AllowDatabaseReset {
		slog.Error("Database reset is disabled; set ALLOW_DATABASE_RESET=true to continue")
		os.Exit(1)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`
			TRUNCATE TABLE messages, reports, rooms, profiles, users
			RESTART IDENTITY CASCADE
		`).Error
	})
	if err != nil {
		slog.Error("Failed to reset database", "error", err)
		os.Exit(1)
	}

	slog.Info("Development database reset completed", "database", cfg.Database.Name)
}
