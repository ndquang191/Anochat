package database

import (
	"strings"
	"testing"
)

func TestEmbeddedMigrationsAreOrderedAndVersioned(t *testing.T) {
	items, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("migration count = %d, want at least 2", len(items))
	}
	for i, item := range items {
		if item.Version <= 0 {
			t.Errorf("migration %q has invalid version %d", item.Name, item.Version)
		}
		if i > 0 && items[i-1].Version >= item.Version {
			t.Errorf("migrations are not strictly ordered: %d then %d", items[i-1].Version, item.Version)
		}
		if strings.TrimSpace(item.SQL) == "" {
			t.Errorf("migration %q is empty", item.Name)
		}
		statements, err := splitSQLStatements(item.SQL)
		if err != nil {
			t.Errorf("migration %q cannot be parsed: %v", item.Name, err)
		} else if len(statements) == 0 {
			t.Errorf("migration %q contains no executable statements", item.Name)
		}
	}
}

func TestActiveRoomMigrationEnforcesMembershipWithTrigger(t *testing.T) {
	items, err := loadMigrations()
	if err != nil {
		t.Fatalf("loadMigrations() error = %v", err)
	}

	var activeRoomSQL string
	for _, item := range items {
		if item.Version == 2 {
			activeRoomSQL = item.SQL
			break
		}
	}
	for _, required := range []string{
		"CREATE TABLE IF NOT EXISTS active_room_members",
		"user_id UUID PRIMARY KEY",
		"CREATE TRIGGER rooms_active_membership_trigger",
		"AFTER INSERT OR UPDATE OF ended_at, user1_id, user2_id ON rooms",
	} {
		if !strings.Contains(activeRoomSQL, required) {
			t.Errorf("active-room migration is missing %q", required)
		}
	}
}

func TestSplitSQLStatementsPreservesDollarQuotedBodies(t *testing.T) {
	source := `
CREATE TABLE example (id UUID);
DO $$
BEGIN
    PERFORM 'value;still-in-body';
END
$$;
CREATE FUNCTION example_fn() RETURNS void AS $function$
BEGIN
    -- semicolon inside a function body
    PERFORM 1;
END;
$function$ LANGUAGE plpgsql;
`

	statements, err := splitSQLStatements(source)
	if err != nil {
		t.Fatalf("splitSQLStatements() error = %v", err)
	}
	if len(statements) != 3 {
		t.Fatalf("statement count = %d, want 3: %#v", len(statements), statements)
	}
	if !strings.Contains(statements[1], "value;still-in-body") {
		t.Errorf("DO body was split incorrectly: %q", statements[1])
	}
	if !strings.Contains(statements[2], "PERFORM 1;") {
		t.Errorf("function body was split incorrectly: %q", statements[2])
	}
}
