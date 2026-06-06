package capabilities_test

import (
	"database/sql"
	"testing"

	"github.com/plan-ai/plan-ai/internal/capabilities"
	_ "modernc.org/sqlite"
)

func openSeedTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS capabilities_v2 (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			schema_info TEXT NOT NULL DEFAULT '{}',
			version TEXT NOT NULL DEFAULT '1.0',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		);`)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

func TestCapabilities_SeedFromDB(t *testing.T) {
	db := openSeedTestDB(t)
	defer db.Close()

	if err := capabilities.SeedDefaults(db); err != nil {
		t.Fatalf("SeedDefaults failed: %v", err)
	}

	// Verify 11+ capabilities are inserted
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM capabilities_v2").Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count < 11 {
		t.Fatalf("expected >= 11 capabilities, got %d", count)
	}

	// SeedDefaults must be idempotent
	if err := capabilities.SeedDefaults(db); err != nil {
		t.Fatalf("second SeedDefaults failed: %v", err)
	}
	var count2 int
	if err := db.QueryRow("SELECT COUNT(*) FROM capabilities_v2").Scan(&count2); err != nil {
		t.Fatalf("count query after re-seed: %v", err)
	}
	if count2 != count {
		t.Fatalf("idempotency violated: %d -> %d", count, count2)
	}

	// Verify List() returns them
	r := capabilities.NewRegistry(db)
	list := r.ListCapabilities()
	if len(list) != count {
		t.Fatalf("ListCapabilities returned %d, want %d", len(list), count)
	}
}

func TestCapabilities_DisabledNotListed(t *testing.T) {
	db := openSeedTestDB(t)
	defer db.Close()

	if err := capabilities.SeedDefaults(db); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}

	// Insert a disabled capability directly
	_, err := db.Exec(
		`INSERT INTO capabilities_v2 (id, name, description, schema_info, version, enabled, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"cap-disabled", "disabled_skill", "A disabled capability", "{}", "1.0", 0, "2024-01-01T00:00:00Z",
	)
	if err != nil {
		t.Fatalf("insert disabled: %v", err)
	}

	r := capabilities.NewRegistry(db)
	list := r.ListCapabilities()

	// Disabled capability should not appear in List
	for _, c := range list {
		if c.Name == "disabled_skill" {
			t.Errorf("disabled capability %q should not be listed", c.Name)
		}
	}
}
