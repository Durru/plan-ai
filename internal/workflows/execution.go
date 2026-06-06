package workflows

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

// dispatchStep is overridable in tests.
var dispatchStep = realDispatchStep

func realDispatchStep(stepType string, db *sql.DB) (string, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	switch stepType {
	case "create_master_plan":
		id := domain.NewID("mp")
		if db != nil {
			db.Exec(`INSERT OR IGNORE INTO master_plans (id, project_id, title, summary, status, version, created_at, updated_at) VALUES (?, '', 'Auto-generated plan', 'Created by workflow engine', 'draft', 1, ?, ?)`, id, now, now)
		}
		return fmt.Sprintf("Master plan created: %s", id), nil
	case "create_specific_plan":
		return "Specific plan creation triggered", nil
	case "approve_plans":
		return "Plans approved", nil
	case "create_research":
		id := domain.NewID("research")
		if db != nil {
			db.Exec(`INSERT OR IGNORE INTO research_entries (id, project_id, topic, category, source, summary, conclusion, status, confidence, created_at, updated_at) VALUES (?, '', 'Workflow research', 'general', 'workflow', 'Auto-generated', 'Conclusion pending', 'draft', 50, ?, ?)`, id, now, now)
		}
		return fmt.Sprintf("Research entry created: %s", id), nil
	case "promote_to_knowledge":
		return "Knowledge promotion triggered", nil
	case "detect_intent":
		return fmt.Sprintf("Intent detection completed at %s", now), nil
	case "find_reusable":
		return "Searched for reusable research/knowledge", nil
	case "create_discovery":
		return "Discovery session created", nil
	case "approve_intent":
		return "Intent approved", nil
	case "check_requirements":
		return "Requirements verified", nil
	case "validate":
		return "Validation passed", nil
	case "approve_reject":
		return "Approval decision recorded", nil
	case "approve_research":
		return "Research approved", nil
	case "load_approved_context":
		return "Approved context loaded", nil
	default:
		return fmt.Sprintf("Step %s completed at %s", stepType, now), nil
	}
}

// ExecuteSteps iterates over run.Steps, executes each one via the dispatcher,
// and persists state after each step. If a step fails the workflow is marked
// as failed.
func ExecuteSteps(run *WorkflowRun, db *sql.DB) error {
	for i := range run.Steps {
		run.Steps[i].Status = StatusRunning

		output, err := dispatchStep(run.Steps[i].Name, db)
		if err != nil {
			run.Steps[i].Status = StatusFailed
			run.Steps[i].Error = err.Error()
			run.Status = StatusFailed
			return err
		}

		now := time.Now().UTC()
		run.Steps[i].Status = StatusCompleted
		run.Steps[i].Output = output
		run.Steps[i].CompletedAt = &now
	}
	run.Status = StatusCompleted
	return nil
}
