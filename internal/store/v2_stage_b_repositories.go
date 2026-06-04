package store

import (
	"database/sql"

	ctx "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/reference"
	"github.com/plan-ai/plan-ai/internal/research"
)

type SmartContextPackageRepository struct{ db *sql.DB }
type ResearchOrchestrationRepository struct{ db *sql.DB }
type ReferenceRepository struct{ db *sql.DB }

func NewSmartContextPackageRepository(db *sql.DB) SmartContextPackageRepository {
	return SmartContextPackageRepository{db: db}
}
func NewResearchOrchestrationRepository(db *sql.DB) ResearchOrchestrationRepository {
	return ResearchOrchestrationRepository{db: db}
}
func NewReferenceRepository(db *sql.DB) ReferenceRepository { return ReferenceRepository{db: db} }

var _ ctx.SmartPackageRepository = SmartContextPackageRepository{}
var _ research.OrchestrationRepository = ResearchOrchestrationRepository{}
var _ reference.Repository = ReferenceRepository{}

func (r SmartContextPackageRepository) SavePackage(pkg ctx.SmartPackage) (ctx.SmartPackage, error) {
	if pkg.ID == "" {
		pkg.ID = domain.NewID("ctxpkg")
	}
	c, u := timestamps(pkg.CreatedAt, pkg.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO context_packages_v2 (id, project_id, package_type, model_target, summary, content, priority, token_budget, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET summary=excluded.summary, content=excluded.content, priority=excluded.priority, token_budget=excluded.token_budget, updated_at=excluded.updated_at`, pkg.ID, pkg.ProjectID, pkg.Type, pkg.ModelTarget, pkg.Summary, pkg.Content, pkg.Priority, pkg.TokenBudget, c, u)
	if err != nil {
		return ctx.SmartPackage{}, err
	}
	pkg.CreatedAt, pkg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return pkg, nil
}

func (r SmartContextPackageRepository) ListPackages(projectID string) ([]ctx.SmartPackage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, package_type, model_target, summary, content, priority, token_budget, created_at, updated_at FROM context_packages_v2 WHERE project_id = ? ORDER BY priority, created_at`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ctx.SmartPackage
	for rows.Next() {
		var pkg ctx.SmartPackage
		var c, u string
		if err := rows.Scan(&pkg.ID, &pkg.ProjectID, &pkg.Type, &pkg.ModelTarget, &pkg.Summary, &pkg.Content, &pkg.Priority, &pkg.TokenBudget, &c, &u); err != nil {
			return nil, err
		}
		pkg.CreatedAt, pkg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, pkg)
	}
	return out, rows.Err()
}

func (r SmartContextPackageRepository) GetPackage(id string) (ctx.SmartPackage, error) {
	rows, err := r.db.Query(`SELECT id, project_id, package_type, model_target, summary, content, priority, token_budget, created_at, updated_at FROM context_packages_v2 WHERE id = ?`, id)
	if err != nil {
		return ctx.SmartPackage{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return ctx.SmartPackage{}, sql.ErrNoRows
	}
	var pkg ctx.SmartPackage
	var c, u string
	if err := rows.Scan(&pkg.ID, &pkg.ProjectID, &pkg.Type, &pkg.ModelTarget, &pkg.Summary, &pkg.Content, &pkg.Priority, &pkg.TokenBudget, &c, &u); err != nil {
		return ctx.SmartPackage{}, err
	}
	pkg.CreatedAt, pkg.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return pkg, rows.Err()
}

func (r ResearchOrchestrationRepository) SaveRun(run research.OrchestrationRun) (research.OrchestrationRun, error) {
	if run.ID == "" {
		run.ID = domain.NewID("researchrun")
	}
	c, u := timestamps(run.CreatedAt, run.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO research_orchestration_runs (id, project_id, agent_type, topic, summary, evidence, confidence, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET summary=excluded.summary, evidence=excluded.evidence, confidence=excluded.confidence, status=excluded.status, updated_at=excluded.updated_at`, run.ID, run.ProjectID, run.Agent, run.Topic, run.Summary, mustJSON(run.Evidence), run.Confidence, run.Status, c, u)
	if err != nil {
		return research.OrchestrationRun{}, err
	}
	run.CreatedAt, run.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return run, nil
}

func (r ResearchOrchestrationRepository) ListRuns(projectID string) ([]research.OrchestrationRun, error) {
	rows, err := r.db.Query(`SELECT id, project_id, agent_type, topic, summary, evidence, confidence, status, created_at, updated_at FROM research_orchestration_runs WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []research.OrchestrationRun
	for rows.Next() {
		var run research.OrchestrationRun
		var evidence, c, u string
		if err := rows.Scan(&run.ID, &run.ProjectID, &run.Agent, &run.Topic, &run.Summary, &evidence, &run.Confidence, &run.Status, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(evidence, &run.Evidence)
		run.CreatedAt, run.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, run)
	}
	return out, rows.Err()
}

func (r ResearchOrchestrationRepository) GetRun(id string) (research.OrchestrationRun, error) {
	rows, err := r.db.Query(`SELECT id, project_id, agent_type, topic, summary, evidence, confidence, status, created_at, updated_at FROM research_orchestration_runs WHERE id = ?`, id)
	if err != nil {
		return research.OrchestrationRun{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return research.OrchestrationRun{}, sql.ErrNoRows
	}
	var run research.OrchestrationRun
	var evidence, c, u string
	if err := rows.Scan(&run.ID, &run.ProjectID, &run.Agent, &run.Topic, &run.Summary, &evidence, &run.Confidence, &run.Status, &c, &u); err != nil {
		return research.OrchestrationRun{}, err
	}
	decodeJSON(evidence, &run.Evidence)
	run.CreatedAt, run.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return run, rows.Err()
}

func (r ReferenceRepository) SaveReference(ref reference.Reference) (reference.Reference, error) {
	if ref.ID == "" {
		ref.ID = domain.NewID("ref")
	}
	c, u := timestamps(ref.CreatedAt, ref.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO project_references_v2 (id, project_id, source_type, uri, title, category, notes, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, category=excluded.category, notes=excluded.notes, status=excluded.status, updated_at=excluded.updated_at`, ref.ID, ref.ProjectID, ref.Source, ref.URI, ref.Title, ref.Category, ref.Notes, ref.Status, c, u)
	if err != nil {
		return reference.Reference{}, err
	}
	ref.CreatedAt, ref.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return ref, nil
}

func (r ReferenceRepository) ListReferences(projectID string) ([]reference.Reference, error) {
	rows, err := r.db.Query(`SELECT id, project_id, source_type, uri, title, category, notes, status, created_at, updated_at FROM project_references_v2 WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []reference.Reference
	for rows.Next() {
		var ref reference.Reference
		var c, u string
		if err := rows.Scan(&ref.ID, &ref.ProjectID, &ref.Source, &ref.URI, &ref.Title, &ref.Category, &ref.Notes, &ref.Status, &c, &u); err != nil {
			return nil, err
		}
		ref.CreatedAt, ref.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, ref)
	}
	return out, rows.Err()
}

func (r ReferenceRepository) ApproveReference(id string) (reference.Reference, error) {
	return r.updateReference(id, reference.StatusApproved)
}

func (r ReferenceRepository) RejectReference(id string) (reference.Reference, error) {
	return r.updateReference(id, reference.StatusRejected)
}

func (r ReferenceRepository) updateReference(id string, status reference.Status) (reference.Reference, error) {
	_, err := r.db.Exec(`UPDATE project_references_v2 SET status = ?, updated_at = ? WHERE id = ?`, status, nowString(), id)
	if err != nil {
		return reference.Reference{}, err
	}
	rows, err := r.db.Query(`SELECT id, project_id, source_type, uri, title, category, notes, status, created_at, updated_at FROM project_references_v2 WHERE id = ?`, id)
	if err != nil {
		return reference.Reference{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return reference.Reference{}, sql.ErrNoRows
	}
	var ref reference.Reference
	var c, u string
	if err := rows.Scan(&ref.ID, &ref.ProjectID, &ref.Source, &ref.URI, &ref.Title, &ref.Category, &ref.Notes, &ref.Status, &c, &u); err != nil {
		return reference.Reference{}, err
	}
	ref.CreatedAt, ref.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return ref, rows.Err()
}
