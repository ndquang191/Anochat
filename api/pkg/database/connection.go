package database

import (
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase initializes the database connection using GORM
func InitDatabase() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		slog.Error("DATABASE_URL environment variable is not set")
		return ErrDatabaseURLMissing
	}

	// Configure GORM logger
	var gormLogger logger.Interface
	if os.Getenv("ENV") == "production" {
		gormLogger = logger.Default.LogMode(logger.Silent)
	} else {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return err
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get underlying sql.DB", "error", err)
		return err
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(30)                 // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                  // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(time.Hour)       // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // Maximum connection idle time

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		slog.Error("Failed to ping database", "error", err)
		return err
	}

	DB = db
	slog.Info("Database connection established successfully")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			sqlDB.Close()
		}
		slog.Info("Database connection closed")
	}
}

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	return DB
}

// HealthCheck checks if database is accessible
func HealthCheck() error {
	if DB == nil {
		return ErrDatabaseNotInitialized
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}