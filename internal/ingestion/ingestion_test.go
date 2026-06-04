package ingestion_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plan-ai/plan-ai/internal/ingestion"
	"github.com/plan-ai/plan-ai/internal/store"
)

func TestServiceCreatesRawInputAndIngestedSource(t *testing.T) {
	db := openTestDB(t)
	service := ingestion.NewService(store.NewIngestionRepository(db))

	raw, source, err := service.Ingest(ingestion.InputRequest{
		ProjectID:  "project:test",
		SourceType: ingestion.SourceTypePrompt,
		Content:    " Build a dashboard for admins.\r\n\r\n- Must be fast ",
	})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if raw.ID == "" || raw.ProjectID != "project:test" || raw.RawContent == "" {
		t.Fatalf("raw input not populated: %+v", raw)
	}
	if source.ID == "" || source.RawInputID != raw.ID || source.NormalizedContent != "Build a dashboard for admins.\n\n- Must be fast" {
		t.Fatalf("source not normalized/persisted: %+v", source)
	}
}

func TestClassifierDetectsBasicInputClasses(t *testing.T) {
	tests := []struct {
		name string
		text string
		want ingestion.Classification
	}{
		{"vision", "Vision: build a planning app", ingestion.ClassificationVision},
		{"requirement", "The system must save drafts", ingestion.ClassificationRequirement},
		{"constraint", "Constraint: use SQLite only", ingestion.ClassificationConstraint},
		{"preference", "I prefer a minimal UI", ingestion.ClassificationPreference},
		{"decision", "Decision: use Go", ingestion.ClassificationDecision},
		{"reference", "Reference: https://example.com", ingestion.ClassificationReference},
		{"unknown", "hello there", ingestion.ClassificationUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ingestion.Classify(tt.text); got != tt.want {
				t.Fatalf("Classify() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNormalizeDetectsBlocksAndLists(t *testing.T) {
	n := ingestion.Normalize("Title\r\n\r\n  - first\r\n  - second  ")
	if n.Content != "Title\n\n- first\n- second" {
		t.Fatalf("content = %q", n.Content)
	}
	if len(n.Blocks) != 2 {
		t.Fatalf("blocks = %d, want 2", len(n.Blocks))
	}
	if len(n.ListItems) != 2 || n.ListItems[0] != "first" || n.ListItems[1] != "second" {
		t.Fatalf("list items = %#v", n.ListItems)
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

func TestExtractorFindsReferences(t *testing.T) {
	refs := ingestion.ExtractReferences("See https://example.com and image ./screen.png")
	joined := strings.Join(refs, " ")
	if !strings.Contains(joined, "https://example.com") || !strings.Contains(joined, "./screen.png") {
		t.Fatalf("refs = %#v", refs)
	}
}
