package vision_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Durru/plan-ai/internal/ingestion"
	"github.com/Durru/plan-ai/internal/store"
	"github.com/Durru/plan-ai/internal/vision"
)

func TestServiceExtractsObjectiveAndConstraints(t *testing.T) {
	db := openTestDB(t)
	service := vision.NewService(store.NewVisionDraftRepository(db))
	sources := []ingestion.IngestedSource{{
		ProjectID:         "project:test",
		SourceType:        ingestion.SourceTypePrompt,
		NormalizedContent: "Build a planning assistant for product teams. It must use SQLite. Success: users approve a plan faster.",
	}}

	draft, err := service.CreateDraft("project:test", sources)
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if draft.Summary == "" || draft.ExpectedOutcome == "" {
		t.Fatalf("objective/outcome missing: %+v", draft)
	}
	if len(draft.Constraints) == 0 || draft.Constraints[0] == "" {
		t.Fatalf("constraints missing: %+v", draft.Constraints)
	}
	if draft.Approved {
		t.Fatalf("draft must not be approved without user approval")
	}
}

func TestServiceDetectsMissingInformation(t *testing.T) {
	db := openTestDB(t)
	service := vision.NewService(store.NewVisionDraftRepository(db))
	draft, err := service.CreateDraft("project:test", []ingestion.IngestedSource{{ProjectID: "project:test", NormalizedContent: "Build a tool."}})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	if len(draft.MissingInformation) == 0 {
		t.Fatalf("missing information should be populated for sparse input: %+v", draft)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}
