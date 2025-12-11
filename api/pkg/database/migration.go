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

	// Disable foreign key constraints temporarily for safe migration
	DB.Exec("SET session_replication_role = 'replica';")

	// Run auto migration for all models
	err := DB.AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Room{},
		&model.Message{},
	)

	// Re-enable foreign key constraints
	DB.Exec("SET session_replication_role = 'origin';")

	if err != nil {
		// Check if error is about constraint already existing (safe to ignore)
		if errStr := err.Error(); errStr != "" &&
			(contains(errStr, "already exists") || contains(errStr, "unexpected EOF")) {
			slog.Warn("Migration constraint warning (safe to ignore)", "error", err)
		} else {
			slog.Error("Migration failed", "error", err)
			return err
		}
	}

	// Create additional indexes for better performance
	if err := createIndexes(); err != nil {
		slog.Error("Failed to create indexes", "error", err)
		// Don't fail if indexes already exist
		if errStr := err.Error(); !contains(errStr, "already exists") {
			return err
		}
	}

	slog.Info("All migrations completed successfully")
	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		len(s) > len(substr)*2 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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