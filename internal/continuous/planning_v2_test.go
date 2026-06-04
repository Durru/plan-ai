package continuous_test

import (
	"testing"

	. "github.com/plan-ai/plan-ai/internal/continuous"
)

func TestBuildTargetedRegenerationNarrowsDatabaseChange(t *testing.T) {
	regen := BuildTargetedRegeneration("project", "Database migration PostgreSQL to MariaDB", "database")
	if !regen.SnapshotRequired || !regen.ApprovalRequired || len(regen.AffectedSections) < 4 {
		t.Fatalf("unexpected regeneration: %#v", regen)
	}
}
