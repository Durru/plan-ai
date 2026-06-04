package store

import (
	"path/filepath"
	"testing"

	"github.com/plan-ai/plan-ai/internal/intent"
)

func TestIntentProfileRepositoryPersistsAndApprovesProfile(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "project.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	if err := RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := NewIntentProfileRepository(db)
	saved, err := repo.SaveProfile(intent.Profile{
		ProjectID:     "project",
		Source:        "quiero un SaaS CRM",
		PrimaryIntent: intent.Intent{Name: "CRM", Confidence: 90, State: intent.SignalCandidate},
		SecondaryGoals: []intent.Goal{
			{Name: "SaaS", State: intent.SignalCandidate},
		},
		Status: intent.StatusDraft,
	})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	latest, err := repo.LatestProfile("project")
	if err != nil {
		t.Fatalf("latest: %v", err)
	}
	if latest.ID != saved.ID || latest.PrimaryIntent.Name != "CRM" {
		t.Fatalf("latest = %#v, saved = %#v", latest, saved)
	}

	approved, err := repo.ApproveProfile(saved.ID)
	if err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !approved.Approved || approved.Status != intent.StatusApproved {
		t.Fatalf("approved = %#v", approved)
	}
}
