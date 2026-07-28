package database

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

const migrationAdvisoryLockID int64 = 0x414e4f43484154

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	Version int64
	Name    string
	SQL     string
}

func RunMigrations() error {
	if DB == nil {
		return ErrDatabaseNotInitialized
	}

	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, item := range migrations {
		if err := applyMigration(DB, item); err != nil {
			return err
		}
	}

	slog.Info("Database migrations are up to date", "count", len(migrations))
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	items := make([]migration, 0, len(entries))
	seen := make(map[int64]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		versionText, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		version, err := strconv.ParseInt(versionText, 10, 64)
		if err != nil || version <= 0 {
			return nil, fmt.Errorf("invalid migration version in %q", entry.Name())
		}
		if previous, exists := seen[version]; exists {
			return nil, fmt.Errorf("duplicate migration version %d in %q and %q", version, previous, entry.Name())
		}
		body, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		seen[version] = entry.Name()
		items = append(items, migration{Version: version, Name: entry.Name(), SQL: string(body)})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Version < items[j].Version })
	return items, nil
}

func applyMigration(db *gorm.DB, item migration) error {
	startedAt := time.Now()
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", migrationAdvisoryLockID).Error; err != nil {
			return fmt.Errorf("lock migration %d: %w", item.Version, err)
		}

		var applied bool
		if err := tx.Raw(
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = ?)",
			item.Version,
		).Scan(&applied).Error; err != nil {
			return fmt.Errorf("check migration %d: %w", item.Version, err)
		}
		if applied {
			return nil
		}

		statements, err := splitSQLStatements(item.SQL)
		if err != nil {
			return fmt.Errorf("parse migration %d (%s): %w", item.Version, item.Name, err)
		}
		for statementIndex, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return fmt.Errorf(
					"apply migration %d (%s), statement %d: %w",
					item.Version,
					item.Name,
					statementIndex+1,
					err,
				)
			}
		}
		if err := tx.Exec(
			"INSERT INTO schema_migrations (version, name) VALUES (?, ?)",
			item.Version,
			item.Name,
		).Error; err != nil {
			return fmt.Errorf("record migration %d: %w", item.Version, err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	slog.Info("Database migration checked", "version", item.Version, "name", item.Name, "duration", time.Since(startedAt))
	return nil
}

func splitSQLStatements(source string) ([]string, error) {
	var statements []string
	var current strings.Builder
	var dollarTag string
	var quote byte
	inLineComment := false
	blockCommentDepth := 0

	for index := 0; index < len(source); {
		if inLineComment {
			current.WriteByte(source[index])
			if source[index] == '\n' {
				inLineComment = false
			}
			index++
			continue
		}
		if blockCommentDepth > 0 {
			switch {
			case strings.HasPrefix(source[index:], "/*"):
				blockCommentDepth++
				current.WriteString("/*")
				index += 2
			case strings.HasPrefix(source[index:], "*/"):
				blockCommentDepth--
				current.WriteString("*/")
				index += 2
			default:
				current.WriteByte(source[index])
				index++
			}
			continue
		}
		if dollarTag != "" {
			if strings.HasPrefix(source[index:], dollarTag) {
				current.WriteString(dollarTag)
				index += len(dollarTag)
				dollarTag = ""
			} else {
				current.WriteByte(source[index])
				index++
			}
			continue
		}
		if quote != 0 {
			current.WriteByte(source[index])
			if source[index] == quote {
				if index+1 < len(source) && source[index+1] == quote {
					current.WriteByte(source[index+1])
					index += 2
					continue
				}
				quote = 0
			}
			index++
			continue
		}

		switch {
		case strings.HasPrefix(source[index:], "--"):
			inLineComment = true
			current.WriteString("--")
			index += 2
		case strings.HasPrefix(source[index:], "/*"):
			blockCommentDepth = 1
			current.WriteString("/*")
			index += 2
		case source[index] == '\'' || source[index] == '"':
			quote = source[index]
			current.WriteByte(source[index])
			index++
		case source[index] == '$':
			tagEnd := index + 1
			for tagEnd < len(source) &&
				((source[tagEnd] >= 'a' && source[tagEnd] <= 'z') ||
					(source[tagEnd] >= 'A' && source[tagEnd] <= 'Z') ||
					(source[tagEnd] >= '0' && source[tagEnd] <= '9') ||
					source[tagEnd] == '_') {
				tagEnd++
			}
			if tagEnd < len(source) && source[tagEnd] == '$' {
				dollarTag = source[index : tagEnd+1]
				current.WriteString(dollarTag)
				index = tagEnd + 1
			} else {
				current.WriteByte(source[index])
				index++
			}
		case source[index] == ';':
			if statement := strings.TrimSpace(current.String()); statement != "" {
				statements = append(statements, statement)
			}
			current.Reset()
			index++
		default:
			current.WriteByte(source[index])
			index++
		}
	}

	if quote != 0 || dollarTag != "" || blockCommentDepth > 0 {
		return nil, fmt.Errorf("unterminated SQL quote or comment")
	}
	if statement := strings.TrimSpace(current.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements, nil
}
