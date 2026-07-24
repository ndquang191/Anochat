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
		&model.RoomSession{},
		&model.Message{},
		&model.BannedWord{},
		&model.Report{},
		&model.ReportMessage{},
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

	// Existing suspended accounts predate the counters. Their full history
	// cannot be reconstructed, but the current suspension is at least one ban.
	if err := DB.Exec(`
		UPDATE users
		SET ban_count = 1,
		    banned_at = COALESCE(banned_at, created_at)
		WHERE is_active = false
		  AND is_deleted = false
		  AND ban_count = 0
	`).Error; err != nil {
		slog.Error("Failed to backfill ban counters", "error", err)
		return err
	}

	if err := createIndexes(); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			slog.Error("Failed to create indexes", "error", err)
			return err
		}
	}

	// Preserve currently active rooms in the lifecycle report when upgrading.
	if err := DB.Exec(`
		INSERT INTO room_sessions (room_id, status, matched_at)
		SELECT id, 'matched', created_at
		FROM rooms
		ON CONFLICT (room_id) DO NOTHING
	`).Error; err != nil {
		slog.Error("Failed to backfill room sessions", "error", err)
		return err
	}

	slog.Info("All migrations completed successfully")
	return nil
}

func createIndexes() error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_users_is_deleted ON users(is_deleted)",
		"CREATE INDEX IF NOT EXISTS idx_users_review_requested ON users(review_requested) WHERE is_active = false",
		"CREATE INDEX IF NOT EXISTS idx_users_banned_admin_cursor ON users(review_requested DESC, review_requested_at DESC, banned_at DESC, created_at DESC, id DESC) WHERE is_active = false AND is_deleted = false",
		"CREATE INDEX IF NOT EXISTS idx_profiles_is_male ON profiles(is_male)",
		"CREATE INDEX IF NOT EXISTS idx_profiles_age ON profiles(age)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_user1_id ON rooms(user1_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_user2_id ON rooms(user2_id)",
		"CREATE INDEX IF NOT EXISTS idx_rooms_created_at ON rooms(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_room_sessions_matched_at ON room_sessions(matched_at)",
		"CREATE INDEX IF NOT EXISTS idx_room_sessions_status ON room_sessions(status)",
		"CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)",
		"CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages(room_id, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_messages_room_created_id ON messages(room_id, created_at DESC, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_reports_reported_user_id ON reports(reported_user_id)",
		"CREATE INDEX IF NOT EXISTS idx_reports_status ON reports(status)",
		"CREATE INDEX IF NOT EXISTS idx_reports_created_at ON reports(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_reports_admin_grouping ON reports(status, reported_user_id, created_at DESC, id DESC)",
		"CREATE INDEX IF NOT EXISTS idx_report_messages_report_id ON report_messages(report_id)",
		"CREATE INDEX IF NOT EXISTS idx_report_messages_created_at ON report_messages(created_at)",
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
