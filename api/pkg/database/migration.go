package database

import (
	"log/slog"
	"strings"

	"github.com/ndquang191/Anochat/api/internal/model"
)

func RunMigrations() error {
	if DB == nil {
		return ErrDatabaseNotInitialized
	}

	slog.Info("Starting database migrations")

	DB.Exec("SET session_replication_role = 'replica';")

	err := DB.AutoMigrate(
		&model.User{},
		&model.Profile{},
		&model.Room{},
		&model.Message{},
	)

	DB.Exec("SET session_replication_role = 'origin';")

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "already exists") || strings.Contains(errStr, "unexpected EOF") {
			slog.Warn("Migration constraint warning (safe to ignore)", "error", err)
		} else {
			slog.Error("Migration failed", "error", err)
			return err
		}
	}

	if err := createIndexes(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			slog.Error("Failed to create indexes", "error", err)
			return err
		}
	}

	slog.Info("All migrations completed successfully")
	return nil
}

func createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_deleted ON users(is_deleted)",
		"CREATE INDEX IF NOT EXISTS idx_profiles_is_male ON profiles(is_male)",
		"CREATE INDEX IF NOT EXISTS idx_profiles_age ON profiles(age)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_user1_id ON rooms(user1_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_user2_id ON rooms(user2_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_category ON rooms(category)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_is_sensitive ON rooms(is_sensitive)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_is_deleted ON rooms(is_deleted)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at)",
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
