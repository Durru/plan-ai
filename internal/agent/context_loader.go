package agent

import (
	"database/sql"

	approvedcontext "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/store"
)

// StoreContextLoader loads context from the store using repositories.
type StoreContextLoader struct {
	db *sql.DB
}

// NewContextLoader creates a StoreContextLoader.
func NewContextLoader(db *sql.DB) *StoreContextLoader {
	return &StoreContextLoader{db: db}
}

// Load loads context data for the requested keys.
func (l *StoreContextLoader) Load(projectID string, keys []string) (ContextPayload, error) {
	var ctx ContextPayload
	ctx.ProjectID = projectID

	for _, key := range keys {
		switch key {
		case "approved":
			l.loadApproved(&ctx, projectID)
		case "visions":
			l.loadVisions(&ctx, projectID)
		case "decisions":
			l.loadDecisions(&ctx, projectID)
		case "master_plans", "plans":
			l.loadMasterPlans(&ctx, projectID)
		case "knowledge":
			l.loadKnowledge(&ctx, projectID)
		case "research":
			l.loadResearch(&ctx, projectID)
		case "phases":
			l.loadPhases(&ctx, projectID)
		case "tasks":
			l.loadTasks(&ctx, projectID)
		case "validations":
			l.loadValidations(&ctx, projectID)
		case "status":
			// no heavy loading for status
		}
	}

	return ctx, nil
}

func (l *StoreContextLoader) loadApproved(ctx *ContextPayload, projectID string) {
	repo := store.NewApprovedContextRepository(l.db)
	reqs, err := repo.ListApproved(projectID, approvedcontext.TypeRequirement)
	if err == nil {
		for _, r := range reqs {
			ctx.Approved.Requirements = append(ctx.Approved.Requirements, r.Content)
		}
	}
	decisions, err := repo.ListApproved(projectID, approvedcontext.TypeDecision)
	if err == nil {
		for _, d := range decisions {
			ctx.Approved.Decisions = append(ctx.Approved.Decisions, d.Content)
		}
	}
	constraints, err := repo.ListApproved(projectID, approvedcontext.TypeConstraint)
	if err == nil {
		for _, c := range constraints {
			ctx.Approved.Constraints = append(ctx.Approved.Constraints, c.Content)
		}
	}
}

func (l *StoreContextLoader) loadVisions(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, title, summary FROM visions WHERE project_id = ? ORDER BY created_at DESC LIMIT 5`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, title, summary string
		if err := rows.Scan(&id, &title, &summary); err != nil {
			continue
		}
		ctx.Visions = append(ctx.Visions, map[string]any{"id": id, "title": title, "summary": summary})
	}
}

func (l *StoreContextLoader) loadDecisions(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, title, decision, status FROM decisions WHERE project_id = ? ORDER BY created_at DESC LIMIT 20`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, title, decision, status string
		if err := rows.Scan(&id, &title, &decision, &status); err != nil {
			continue
		}
		ctx.Decisions = append(ctx.Decisions, map[string]any{"id": id, "title": title, "decision": decision, "status": status})
	}
}

func (l *StoreContextLoader) loadMasterPlans(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, title, summary, status FROM master_plans WHERE project_id = ? ORDER BY created_at DESC LIMIT 10`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, title, summary, status string
		if err := rows.Scan(&id, &title, &summary, &status); err != nil {
			continue
		}
		ctx.Plans = append(ctx.Plans, map[string]any{"id": id, "title": title, "summary": summary, "status": status})
	}
}

func (l *StoreContextLoader) loadKnowledge(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, topic, summary, confidence FROM knowledge_objects WHERE project_id = ? ORDER BY reuse_count DESC, created_at DESC LIMIT 10`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, topic, summary string
		var confidence float64
		if err := rows.Scan(&id, &topic, &summary, &confidence); err != nil {
			continue
		}
		ctx.Knowledge = append(ctx.Knowledge, map[string]any{"id": id, "topic": topic, "summary": summary, "confidence": confidence})
	}
}

func (l *StoreContextLoader) loadResearch(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, topic, summary, confidence FROM research_entries WHERE project_id = ? ORDER BY created_at DESC LIMIT 10`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, topic, summary string
		var confidence float64
		if err := rows.Scan(&id, &topic, &summary, &confidence); err != nil {
			continue
		}
		ctx.Research = append(ctx.Research, map[string]any{"id": id, "topic": topic, "summary": summary, "confidence": confidence})
	}
}

func (l *StoreContextLoader) loadPhases(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, plan_id, title, status FROM phases WHERE project_id = ? ORDER BY position, created_at LIMIT 20`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, planID, title, status string
		if err := rows.Scan(&id, &planID, &title, &status); err != nil {
			continue
		}
		ctx.Phases = append(ctx.Phases, map[string]any{"id": id, "plan_id": planID, "title": title, "status": status})
	}
}

func (l *StoreContextLoader) loadTasks(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, phase_id, title, status, position FROM tasks WHERE project_id = ? ORDER BY position, created_at LIMIT 30`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, phaseID, title, status string
		var position int
		if err := rows.Scan(&id, &phaseID, &title, &status, &position); err != nil {
			continue
		}
		ctx.Tasks = append(ctx.Tasks, map[string]any{"id": id, "phase_id": phaseID, "title": title, "status": status, "position": position})
	}
}

func (l *StoreContextLoader) loadValidations(ctx *ContextPayload, projectID string) {
	rows, err := l.db.Query(`SELECT id, target_type, status, summary FROM validations WHERE project_id = ? ORDER BY created_at DESC LIMIT 20`, projectID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var id, targetType, status, summary string
		if err := rows.Scan(&id, &targetType, &status, &summary); err != nil {
			continue
		}
		ctx.Validations = append(ctx.Validations, map[string]any{"id": id, "target_type": targetType, "status": status, "summary": summary})
	}
}
