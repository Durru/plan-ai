package store

import (
	"database/sql"

	contextpkg "github.com/Durru/plan-ai/internal/context"
)

// DomainQuerier provides read access to domain entities for context building.
// It implements context.ProjectQuerier.
type DomainQuerier struct{ db *sql.DB }

func NewDomainQuerier(db *sql.DB) *DomainQuerier {
	return &DomainQuerier{db: db}
}

var _ contextpkg.ProjectQuerier = (*DomainQuerier)(nil)

func (q *DomainQuerier) GetProjectBrief(id string) (contextpkg.ProjectBrief, error) {
	var p contextpkg.ProjectBrief
	err := q.db.QueryRow(`SELECT id, name, status FROM projects WHERE id = ?`, id).Scan(&p.ID, &p.Name, &p.Status)
	return p, err
}

func (q *DomainQuerier) ListPlanBriefs(projectID string) ([]contextpkg.PlanBrief, error) {
	rows, err := q.db.Query(`SELECT id, title, status FROM master_plans WHERE project_id = ?`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var plans []contextpkg.PlanBrief
	for rows.Next() {
		var p contextpkg.PlanBrief
		if err := rows.Scan(&p.ID, &p.Title, &p.Status); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (q *DomainQuerier) ListPhaseBriefs(planID string) ([]contextpkg.PhaseBrief, error) {
	rows, err := q.db.Query(`SELECT id, title, status FROM phases WHERE plan_id = ? ORDER BY position`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var phases []contextpkg.PhaseBrief
	for rows.Next() {
		var ph contextpkg.PhaseBrief
		if err := rows.Scan(&ph.ID, &ph.Title, &ph.Status); err != nil {
			return nil, err
		}
		phases = append(phases, ph)
	}
	return phases, rows.Err()
}

func (q *DomainQuerier) ListTaskBriefs(phaseID string) ([]contextpkg.TaskBrief, error) {
	rows, err := q.db.Query(`SELECT id, title, status FROM tasks WHERE phase_id = ? ORDER BY position`, phaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []contextpkg.TaskBrief
	for rows.Next() {
		var t contextpkg.TaskBrief
		if err := rows.Scan(&t.ID, &t.Title, &t.Status); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (q *DomainQuerier) ListDecisionBriefs(projectID string) ([]contextpkg.DecisionBrief, error) {
	rows, err := q.db.Query(`SELECT id, title, status FROM decisions WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var decisions []contextpkg.DecisionBrief
	for rows.Next() {
		var d contextpkg.DecisionBrief
		if err := rows.Scan(&d.ID, &d.Title, &d.Status); err != nil {
			return nil, err
		}
		decisions = append(decisions, d)
	}
	return decisions, rows.Err()
}

func (q *DomainQuerier) ListResearchBriefs(projectID string) ([]contextpkg.ResearchBrief, error) {
	rows, err := q.db.Query(`SELECT id, topic, summary FROM research_entries WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []contextpkg.ResearchBrief
	for rows.Next() {
		var r contextpkg.ResearchBrief
		if err := rows.Scan(&r.ID, &r.Topic, &r.Summary); err != nil {
			return nil, err
		}
		entries = append(entries, r)
	}
	return entries, rows.Err()
}

func (q *DomainQuerier) ListKnowledgeBriefs(projectID string) ([]contextpkg.KnowledgeBrief, error) {
	rows, err := q.db.Query(`SELECT id, topic, summary FROM knowledge_objects WHERE project_id = ? ORDER BY created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var objects []contextpkg.KnowledgeBrief
	for rows.Next() {
		var k contextpkg.KnowledgeBrief
		if err := rows.Scan(&k.ID, &k.Topic, &k.Summary); err != nil {
			return nil, err
		}
		objects = append(objects, k)
	}
	return objects, rows.Err()
}
