package change

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS entity_links (
		id TEXT PRIMARY KEY,
		project_id TEXT NOT NULL,
		source_type TEXT NOT NULL,
		source_id TEXT NOT NULL,
		target_type TEXT NOT NULL,
		target_id TEXT NOT NULL,
		link_type TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create entity_links: %v", err)
	}
	return db
}

func insertEntityLink(t *testing.T, db *sql.DB, id, projectID, sourceType, sourceID, targetType, targetID, linkType string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO entity_links (id, project_id, source_type, source_id, target_type, target_id, link_type, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, '2025-01-01T00:00:00Z')`,
		id, projectID, sourceType, sourceID, targetType, targetID, linkType)
	if err != nil {
		t.Fatalf("failed to insert entity_link %s: %v", id, err)
	}
}

func TestEntityAnalyzer_ReturnsRealEntities(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertEntityLink(t, db, "link1", "proj1", "decision", "dec_1", "plan", "plan_1", "affects")
	insertEntityLink(t, db, "link2", "proj1", "plan", "plan_1", "task", "task_1", "implements")

	analyzer := NewAnalyzer(db)
	results, err := analyzer.AnalyzeEntityLinks("proj1", "decision", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one affected entity group")
	}

	foundPlan := false
	foundTask := false
	for _, g := range results {
		for _, id := range g.EntityIDs {
			if g.EntityType == "plan" && id == "plan_1" {
				foundPlan = true
			}
			if g.EntityType == "task" && id == "task_1" {
				foundTask = true
			}
		}
	}
	if !foundPlan {
		t.Error("expected plan_1 to be affected")
	}
	if !foundTask {
		t.Error("expected task_1 to be affected")
	}
}

func TestEntityAnalyzer_HandlesEmptyGraph(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	analyzer := NewAnalyzer(db)
	results, err := analyzer.AnalyzeEntityLinks("proj1", "decision", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected empty results for empty graph, got %d groups", len(results))
	}
}

func TestImpactGraph_AnalyzeChangePropagation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	insertEntityLink(t, db, "link1", "proj1", "decision", "dec_1", "plan", "plan_1", "affects")
	insertEntityLink(t, db, "link2", "proj1", "plan", "plan_1", "task", "task_1", "implements")

	analyzer := NewAnalyzer(db)
	results, err := analyzer.AnalyzeEntityLinks("proj1", "decision", "dec_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) < 2 {
		t.Errorf("expected at least 2 entity types to be affected (plan + task), got %d", len(results))
	}

	totalIDs := 0
	for _, g := range results {
		totalIDs += len(g.EntityIDs)
	}
	if totalIDs < 2 {
		t.Errorf("expected at least 2 entity IDs to be affected, got %d", totalIDs)
	}

	// Verify the entity IDs specifically
	seen := make(map[string]bool)
	for _, g := range results {
		for _, id := range g.EntityIDs {
			seen[g.EntityType+":"+id] = true
		}
	}
	if !seen["plan:plan_1"] {
		t.Error("expected plan:plan_1 in results")
	}
	if !seen["task:task_1"] {
		t.Error("expected task:task_1 in results")
	}
}
