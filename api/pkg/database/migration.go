package database

import (
	"github.com/ndquang191/Anochat/api/internal/model"
	"log/slog"
)

// RunMigrations executes all database migrations using GORM AutoMigrate
func RunMigrations() error {
	if DB == nil {
		return ErrDatabaseNotInitialized
	}

	slog.Info("Starting database migrations")

	// Run auto migration for all models
	err := DB.AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Room{},
		&model.Message{},
	)

	if err != nil {
		slog.Error("Migration failed", "error", err)
		return err
	}

	// Create additional indexes for better performance
	if err := createIndexes(); err != nil {
		slog.Error("Failed to create indexes", "error", err)
		return err
	}

	slog.Info("All migrations completed successfully")
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes() error {
	indexes := []string{
		// Users indexes
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_deleted ON users(is_deleted)",

		// Profiles indexes
		"CREATE INDEX IF NOT EXISTS idx_profiles_is_male ON profiles(is_male)",
		"CREATE INDEX IF NOT EXISTS idx_profiles_age ON profiles(age)",

		// Rooms indexes
		"CREATE INDEX IF NOT EXISTS idx_rooms_user1_id ON rooms(user1_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_user2_id ON rooms(user2_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_category ON rooms(category)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_is_sensitive ON rooms(is_sensitive)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_is_deleted ON rooms(is_deleted)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at)",

		// Messages indexes
		"CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages(room_id, created_at)",
	}

	for _, indexSQL := range indexes {
		if err := DB.Exec(indexSQL).Error; err != nil {
			slog.Error("Failed to create index", "sql", indexSQL, "error", err)
			return err
		}
	}

	slog.Info("Database indexes created successfully")
	return nil
}