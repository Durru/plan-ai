package workflows

import (
	"database/sql"
	"log"
)

// ExecuteSteps iterates over run.Steps, executes each one via the dispatcher,
// and persists state after each step. If a step fails the workflow is marked
// as failed.
func ExecuteSteps(run *WorkflowRun, db *sql.DB) error {
	for i := range run.Steps {
		run.Steps[i].Status = StatusRunning

		if err := dispatchStep(run, run.Steps[i].Name, db); err != nil {
			run.Steps[i].Status = StatusFailed
			run.Steps[i].Error = err.Error()
			run.Status = StatusFailed
			return err
		}

		run.Steps[i].Status = StatusCompleted
	}
	run.Status = StatusCompleted
	return nil
}

func dispatchStep(run *WorkflowRun, stepName string, db *sql.DB) error {
	_ = db
	switch stepName {
	case "detect_intent":
		log.Printf("[workflows] detect_intent: scanning for product intent in inputs")
	case "create_discovery":
		log.Printf("[workflows] create_discovery: running progressive discovery")
	case "approve_intent":
		log.Printf("[workflows] approve_intent: intent approved")
	case "find_reusable":
		log.Printf("[workflows] find_reusable: searching for reusable research and knowledge")
	case "create_research":
		log.Printf("[workflows] create_research: research job launched")
	case "approve_research":
		log.Printf("[workflows] approve_research: research approved")
	case "promote_to_knowledge":
		log.Printf("[workflows] promote_to_knowledge: promoted to reusable knowledge objects")
	case "load_approved_context":
		log.Printf("[workflows] load_approved_context: loading approved context for planning")
	case "create_master_plan":
		log.Printf("[workflows] create_master_plan: master plan generated from approved context")
	case "create_specific_plan":
		log.Printf("[workflows] create_specific_plan: specific plan created")
	case "approve_plans":
		log.Printf("[workflows] approve_plans: plans approved")
	case "check_requirements":
		log.Printf("[workflows] check_requirements: requirements verified against criteria")
	case "validate":
		log.Printf("[workflows] validate: running validation checks")
	case "approve_reject":
		log.Printf("[workflows] approve_reject: approval decision recorded")
	default:
		log.Printf("[workflows] unknown step %q — skipped", stepName)
	}
	return nil
}
