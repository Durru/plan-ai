package context_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	approvedcontext "github.com/Durru/plan-ai/internal/context"
	"github.com/Durru/plan-ai/internal/store"
)

func TestRegistryStoresAndRetrievesApprovedItems(t *testing.T) {
	db := openTestDB(t)
	registry := approvedcontext.NewRegistry(store.NewApprovedContextRepository(db))

	req, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeRequirement, SourceID: "source:1", Content: "The app must save drafts"})
	if err != nil {
		t.Fatalf("store requirement: %v", err)
	}
	if _, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeDecision, SourceID: "source:1", Content: "Use SQLite"}); err != nil {
		t.Fatalf("store decision: %v", err)
	}
	if _, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypeConstraint, SourceID: "source:1", Content: "No network downloads"}); err != nil {
		t.Fatalf("store constraint: %v", err)
	}

	got, err := registry.GetApproved(approvedcontext.TypeRequirement, req.ID)
	if err != nil {
		t.Fatalf("get approved: %v", err)
	}
	if got.State != approvedcontext.StateApproved || got.Content != req.Content {
		t.Fatalf("approved requirement mismatch: %+v", got)
	}

	items, err := registry.ListApproved("project:test", "")
	if err != nil {
		t.Fatalf("list approved: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("items = %d, want 3", len(items))
	}
}

func TestRegistryAvoidsDuplicatesAndFindsApprovedItems(t *testing.T) {
	db := openTestDB(t)
	registry := approvedcontext.NewRegistry(store.NewApprovedContextRepository(db))

	first, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypePreference, SourceID: "source:1", Content: "Prefer concise plans"})
	if err != nil {
		t.Fatalf("store first: %v", err)
	}
	second, err := registry.StoreApproved(approvedcontext.ApprovedItem{ProjectID: "project:test", Type: approvedcontext.TypePreference, SourceID: "source:2", Content: "Prefer concise plans"})
	if err != nil {
		t.Fatalf("store duplicate: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("duplicate created new id: %s vs %s", first.ID, second.ID)
	}

	found, err := registry.FindApproved("project:test", approvedcontext.TypePreference, "concise")
	if err != nil {
		t.Fatalf("find approved: %v", err)
	}
	if len(found) != 1 || found[0].ID != first.ID {
		t.Fatalf("found = %#v", found)
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
