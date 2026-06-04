package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/knowledge"
	"github.com/plan-ai/plan-ai/internal/planning"
	"github.com/plan-ai/plan-ai/internal/research"
	"github.com/plan-ai/plan-ai/internal/workflows"
)

var _ research.RegistryRepository = (*ResearchRepository)(nil)
var _ knowledge.RegistryRepository = KnowledgeRepository{}

func (r *ResearchRepository) CreateResearchJob(job research.ResearchJob) (research.ResearchJob, error) {
	if job.ID == "" {
		job.ID = domain.NewID("research")
	}
	if job.Status == "" {
		job.Status = research.ResearchStatusDraft
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now().UTC()
	}
	c := job.CreatedAt.UTC().Format(time.RFC3339)
	tx, err := r.db.Begin()
	if err != nil {
		return research.ResearchJob{}, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO research_jobs (id, project_id, topic, summary, confidence, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, job.ID, job.ProjectID, job.Topic, job.Summary, job.Confidence, string(job.Status), c); err != nil {
		return research.ResearchJob{}, err
	}
	if _, err := tx.Exec(`INSERT INTO research_entries (id, project_id, topic, source, category, summary, conclusion, status, confidence, created_at, updated_at) VALUES (?, ?, ?, '', 'general', ?, '', ?, ?, ?, ?)`, job.ID, job.ProjectID, job.Topic, job.Summary, string(job.Status), job.Confidence*100, c, c); err != nil {
		return research.ResearchJob{}, err
	}
	for _, f := range job.Findings {
		if f.ID == "" {
			f.ID = domain.NewID("finding")
		}
		if f.CreatedAt.IsZero() {
			f.CreatedAt = job.CreatedAt
		}
		if _, err := tx.Exec(`INSERT INTO research_findings (id, research_id, title, content, importance, created_at) VALUES (?, ?, ?, ?, ?, ?)`, f.ID, job.ID, f.Title, f.Content, f.Importance, f.CreatedAt.UTC().Format(time.RFC3339)); err != nil {
			return research.ResearchJob{}, err
		}
	}
	for _, s := range job.Sources {
		if s.ID == "" {
			s.ID = domain.NewID("source")
		}
		if s.CreatedAt.IsZero() {
			s.CreatedAt = job.CreatedAt
		}
		if _, err := tx.Exec(`INSERT INTO research_sources (id, research_id, title, url, source_type, created_at) VALUES (?, ?, ?, ?, ?, ?)`, s.ID, job.ID, s.Title, s.URL, string(s.SourceType), s.CreatedAt.UTC().Format(time.RFC3339)); err != nil {
			return research.ResearchJob{}, err
		}
	}
	for _, rec := range job.Recommendations {
		if rec.ID == "" {
			rec.ID = domain.NewID("recommendation")
		}
		if rec.CreatedAt.IsZero() {
			rec.CreatedAt = job.CreatedAt
		}
		if _, err := tx.Exec(`INSERT INTO research_recommendations (id, research_id, content, created_at) VALUES (?, ?, ?, ?)`, rec.ID, job.ID, rec.Content, rec.CreatedAt.UTC().Format(time.RFC3339)); err != nil {
			return research.ResearchJob{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return research.ResearchJob{}, err
	}
	return r.GetResearchJob(job.ID)
}

func (r *ResearchRepository) GetResearchJob(id string) (research.ResearchJob, error) {
	items, err := r.listResearchJobs(`WHERE id = ?`, id)
	if err != nil {
		return research.ResearchJob{}, err
	}
	if len(items) == 0 {
		return research.ResearchJob{}, sql.ErrNoRows
	}
	return items[0], nil
}
func (r *ResearchRepository) ListResearchJobs(projectID string) ([]research.ResearchJob, error) {
	if strings.TrimSpace(projectID) == "" {
		return r.listResearchJobs(`ORDER BY created_at, id`)
	}
	return r.listResearchJobs(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (r *ResearchRepository) listResearchJobs(where string, args ...any) ([]research.ResearchJob, error) {
	rows, err := r.db.Query(`SELECT id, project_id, topic, summary, confidence, status, created_at FROM research_jobs `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []research.ResearchJob
	for rows.Next() {
		var j research.ResearchJob
		var status, c string
		if err := rows.Scan(&j.ID, &j.ProjectID, &j.Topic, &j.Summary, &j.Confidence, &status, &c); err != nil {
			return nil, err
		}
		j.Status = research.ResearchStatus(status)
		j.CreatedAt = parseRFC3339(c)
		findings, _ := r.ListFindings(j.ID)
		sources, _ := r.ListSources(j.ID)
		recs, _ := r.listRecommendations(j.ID)
		j.Findings = findings
		j.Sources = sources
		j.Recommendations = recs
		out = append(out, j)
	}
	return out, rows.Err()
}
func (r *ResearchRepository) listRecommendations(researchID string) ([]research.ResearchRecommendation, error) {
	rows, err := r.db.Query(`SELECT id, research_id, content, created_at FROM research_recommendations WHERE research_id = ? ORDER BY created_at, id`, researchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []research.ResearchRecommendation
	for rows.Next() {
		var rec research.ResearchRecommendation
		var c string
		if err := rows.Scan(&rec.ID, &rec.ResearchID, &rec.Content, &c); err != nil {
			return nil, err
		}
		rec.CreatedAt = parseRFC3339(c)
		out = append(out, rec)
	}
	return out, rows.Err()
}

func (r KnowledgeRepository) CreateKnowledgeObject(object knowledge.KnowledgeObject) (knowledge.KnowledgeObject, error) {
	if object.ID == "" {
		object.ID = domain.NewID("knowledge")
	}
	if object.Category == "" {
		object.Category = domain.KnowledgeCategoryGeneral
	}
	now := nowString()
	if object.CreatedAt == "" {
		object.CreatedAt = now
	}
	object.UpdatedAt = now
	tx, err := r.db.Begin()
	if err != nil {
		return knowledge.KnowledgeObject{}, err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`INSERT INTO knowledge_objects (id, project_id, title, topic, category, summary, content, confidence, source_type, reuse_count, status, research_ids, related_decisions, related_requirements, related_constraints, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, '', ?, 'research', 0, 'approved', ?, ?, ?, ?, ?, ?)`, object.ID, object.ProjectID, object.Title, object.Title, string(object.Category), object.Summary, object.Confidence, jsonListLocal(object.ResearchIDs), jsonListLocal(object.RelatedDecisions), jsonListLocal(object.RelatedRequirements), jsonListLocal(object.RelatedConstraints), object.CreatedAt, object.UpdatedAt); err != nil {
		return knowledge.KnowledgeObject{}, err
	}
	for _, id := range object.ResearchIDs {
		if err := r.addKnowledgeLinkTx(tx, object.ID, "research", id); err != nil {
			return knowledge.KnowledgeObject{}, err
		}
		if _, err := tx.Exec(`INSERT INTO research_knowledge_links (id, research_id, knowledge_id, created_at) VALUES (?, ?, ?, ?) ON CONFLICT(research_id, knowledge_id) DO NOTHING`, domain.NewID("rlink"), id, object.ID, nowString()); err != nil {
			return knowledge.KnowledgeObject{}, err
		}
	}
	for _, id := range object.RelatedDecisions {
		if err := r.addKnowledgeLinkTx(tx, object.ID, "decision", id); err != nil {
			return knowledge.KnowledgeObject{}, err
		}
	}
	for _, id := range object.RelatedRequirements {
		if err := r.addKnowledgeLinkTx(tx, object.ID, "requirement", id); err != nil {
			return knowledge.KnowledgeObject{}, err
		}
	}
	for _, id := range object.RelatedConstraints {
		if err := r.addKnowledgeLinkTx(tx, object.ID, "constraint", id); err != nil {
			return knowledge.KnowledgeObject{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return knowledge.KnowledgeObject{}, err
	}
	return r.GetKnowledgeObject(object.ID)
}
func (r KnowledgeRepository) GetKnowledgeObject(id string) (knowledge.KnowledgeObject, error) {
	items, err := r.queryKnowledgeObjects(`WHERE id = ?`, id)
	if err != nil {
		return knowledge.KnowledgeObject{}, err
	}
	if len(items) == 0 {
		return knowledge.KnowledgeObject{}, sql.ErrNoRows
	}
	return items[0], nil
}
func (r KnowledgeRepository) ListKnowledgeObjects(projectID string) ([]knowledge.KnowledgeObject, error) {
	return r.queryKnowledgeObjects(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (r KnowledgeRepository) SearchKnowledgeObjects(query string) ([]knowledge.KnowledgeObject, error) {
	like := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	return r.queryKnowledgeObjects(`WHERE lower(title) LIKE ? OR lower(summary) LIKE ? ORDER BY created_at, id`, like, like)
}
func (r KnowledgeRepository) queryKnowledgeObjects(where string, args ...any) ([]knowledge.KnowledgeObject, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, category, summary, confidence, research_ids, related_decisions, related_requirements, related_constraints, created_at, updated_at FROM knowledge_objects `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []knowledge.KnowledgeObject
	for rows.Next() {
		var o knowledge.KnowledgeObject
		var researchIDs, decisions, requirements, constraints string
		if err := rows.Scan(&o.ID, &o.ProjectID, &o.Title, &o.Category, &o.Summary, &o.Confidence, &researchIDs, &decisions, &requirements, &constraints, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.ResearchIDs = scanJSONListLocal(researchIDs)
		o.RelatedDecisions = scanJSONListLocal(decisions)
		o.RelatedRequirements = scanJSONListLocal(requirements)
		o.RelatedConstraints = scanJSONListLocal(constraints)
		out = append(out, o)
	}
	return out, rows.Err()
}
func (r KnowledgeRepository) addKnowledgeLink(knowledgeID, linkType, targetID string) error {
	_, err := r.db.Exec(`INSERT INTO knowledge_links (id, knowledge_id, link_type, target_id, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(knowledge_id, link_type, target_id) DO NOTHING`, domain.NewID("klink"), knowledgeID, linkType, targetID, nowString())
	return err
}

func (r KnowledgeRepository) addKnowledgeLinkTx(tx *sql.Tx, knowledgeID, linkType, targetID string) error {
	_, err := tx.Exec(`INSERT INTO knowledge_links (id, knowledge_id, link_type, target_id, created_at) VALUES (?, ?, ?, ?, ?) ON CONFLICT(knowledge_id, link_type, target_id) DO NOTHING`, domain.NewID("klink"), knowledgeID, linkType, targetID, nowString())
	return err
}

type PlanningRepository struct{ db *sql.DB }

func NewPlanningRepository(db *sql.DB) PlanningRepository { return PlanningRepository{db: db} }

var _ planning.Repository = PlanningRepository{}

func (r PlanningRepository) CreateMasterPlan(p planning.MasterPlan) (planning.MasterPlan, error) {
	c, u := timestamps(p.CreatedAt, p.UpdatedAt)
	if p.ID == "" {
		p.ID = domain.NewID("master")
	}
	_, err := r.db.Exec(`INSERT INTO master_plans (id, project_id, title, summary, status, version, vision_reference, objectives, scope, out_of_scope, recommended_specific_plans, risks, assumptions, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.ID, p.ProjectID, p.Title, p.Title, string(p.Status), p.VisionReference, jsonListLocal(p.Objectives), jsonListLocal(p.Scope), jsonListLocal(p.OutOfScope), jsonListLocal(p.RecommendedSpecificPlans), jsonListLocal(p.Risks), jsonListLocal(p.Assumptions), c, u)
	if err != nil {
		return planning.MasterPlan{}, err
	}
	return r.GetMasterPlan(p.ID)
}
func (r PlanningRepository) GetMasterPlan(id string) (planning.MasterPlan, error) {
	rows, err := r.listMasterPlans(`WHERE id = ?`, id)
	if err != nil {
		return planning.MasterPlan{}, err
	}
	if len(rows) == 0 {
		return planning.MasterPlan{}, sql.ErrNoRows
	}
	return rows[0], nil
}
func (r PlanningRepository) ListMasterPlans(projectID string) ([]planning.MasterPlan, error) {
	return r.listMasterPlans(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (r PlanningRepository) listMasterPlans(where string, args ...any) ([]planning.MasterPlan, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, vision_reference, objectives, scope, out_of_scope, recommended_specific_plans, risks, assumptions, status, created_at, updated_at FROM master_plans `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []planning.MasterPlan
	for rows.Next() {
		var p planning.MasterPlan
		var obj, scope, outscope, recs, risks, assum, status, c, u string
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Title, &p.VisionReference, &obj, &scope, &outscope, &recs, &risks, &assum, &status, &c, &u); err != nil {
			return nil, err
		}
		p.Objectives = scanJSONListLocal(obj)
		p.Scope = scanJSONListLocal(scope)
		p.OutOfScope = scanJSONListLocal(outscope)
		p.RecommendedSpecificPlans = scanJSONListLocal(recs)
		p.Risks = scanJSONListLocal(risks)
		p.Assumptions = scanJSONListLocal(assum)
		p.Status = planning.Status(status)
		p.CreatedAt = parseRFC3339(c)
		p.UpdatedAt = parseRFC3339(u)
		out = append(out, p)
	}
	return out, rows.Err()
}
func (r PlanningRepository) CreateSpecificPlan(p planning.SpecificPlan) (planning.SpecificPlan, error) {
	c, u := timestamps(p.CreatedAt, p.UpdatedAt)
	if p.ID == "" {
		p.ID = domain.NewID("specific")
	}
	_, err := r.db.Exec(`INSERT INTO specific_plans (id, project_id, master_plan_id, title, summary, status, version, goal, requirements, constraints, decisions, knowledge_used, research_used, implementation_strategy, risks, validation_criteria, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, p.ID, p.ProjectID, p.MasterPlanID, p.Title, p.Goal, string(p.Status), p.Goal, jsonListLocal(p.Requirements), jsonListLocal(p.Constraints), jsonListLocal(p.Decisions), jsonListLocal(p.KnowledgeUsed), jsonListLocal(p.ResearchUsed), p.ImplementationStrategy, jsonListLocal(p.Risks), jsonListLocal(p.ValidationCriteria), c, u)
	if err != nil {
		return planning.SpecificPlan{}, err
	}
	return r.GetSpecificPlan(p.ID)
}
func (r PlanningRepository) GetSpecificPlan(id string) (planning.SpecificPlan, error) {
	row := r.db.QueryRow(`SELECT id, project_id, master_plan_id, title, goal, requirements, constraints, decisions, knowledge_used, research_used, implementation_strategy, risks, validation_criteria, status, created_at, updated_at FROM specific_plans WHERE id = ?`, id)
	var p planning.SpecificPlan
	var req, con, dec, know, res, risks, val, status, c, u string
	if err := row.Scan(&p.ID, &p.ProjectID, &p.MasterPlanID, &p.Title, &p.Goal, &req, &con, &dec, &know, &res, &p.ImplementationStrategy, &risks, &val, &status, &c, &u); err != nil {
		return p, err
	}
	p.Requirements = scanJSONListLocal(req)
	p.Constraints = scanJSONListLocal(con)
	p.Decisions = scanJSONListLocal(dec)
	p.KnowledgeUsed = scanJSONListLocal(know)
	p.ResearchUsed = scanJSONListLocal(res)
	p.Risks = scanJSONListLocal(risks)
	p.ValidationCriteria = scanJSONListLocal(val)
	p.Status = planning.Status(status)
	p.CreatedAt = parseRFC3339(c)
	p.UpdatedAt = parseRFC3339(u)
	return p, nil
}
func (r PlanningRepository) CreateImplementationDocument(d planning.ImplementationDocument) (planning.ImplementationDocument, error) {
	c, u := timestamps(d.CreatedAt, d.UpdatedAt)
	if d.ID == "" {
		d.ID = domain.NewID("impl")
	}
	_, err := r.db.Exec(`INSERT INTO implementation_documents (id, project_id, specific_plan_id, title, content, version, objective, architecture, expected_files, expected_directories, validations, known_risks, testing_strategy, rollback_strategy, created_at, updated_at) VALUES (?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, d.ID, d.ProjectID, d.SpecificPlanID, d.Objective, d.Architecture, d.Objective, d.Architecture, jsonListLocal(d.ExpectedFiles), jsonListLocal(d.ExpectedDirectories), jsonListLocal(d.Validations), jsonListLocal(d.KnownRisks), d.TestingStrategy, d.RollbackStrategy, c, u)
	if err != nil {
		return planning.ImplementationDocument{}, err
	}
	return r.GetImplementationDocument(d.ID)
}
func (r PlanningRepository) GetImplementationDocument(id string) (planning.ImplementationDocument, error) {
	row := r.db.QueryRow(`SELECT id, project_id, specific_plan_id, objective, architecture, expected_files, expected_directories, validations, known_risks, testing_strategy, rollback_strategy, created_at, updated_at FROM implementation_documents WHERE id = ?`, id)
	var d planning.ImplementationDocument
	var files, dirs, vals, risks, c, u string
	if err := row.Scan(&d.ID, &d.ProjectID, &d.SpecificPlanID, &d.Objective, &d.Architecture, &files, &dirs, &vals, &risks, &d.TestingStrategy, &d.RollbackStrategy, &c, &u); err != nil {
		return d, err
	}
	d.ExpectedFiles = scanJSONListLocal(files)
	d.ExpectedDirectories = scanJSONListLocal(dirs)
	d.Validations = scanJSONListLocal(vals)
	d.KnownRisks = scanJSONListLocal(risks)
	d.CreatedAt = parseRFC3339(c)
	d.UpdatedAt = parseRFC3339(u)
	return d, nil
}

type WorkflowRunRepository struct{ db *sql.DB }

func NewWorkflowRunRepository(db *sql.DB) WorkflowRunRepository { return WorkflowRunRepository{db: db} }

var _ workflows.RunRepository = WorkflowRunRepository{}

func (r WorkflowRunRepository) CreateWorkflowRun(run workflows.WorkflowRun) (workflows.WorkflowRun, error) {
	if run.ID == "" {
		run.ID = domain.NewID("workflow")
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now().UTC()
	}
	_, err := r.db.Exec(`INSERT INTO workflow_runs (id, workflow_type, status, started_at, finished_at) VALUES (?, ?, ?, ?, ?)`, run.ID, string(run.WorkflowType), string(run.Status), run.StartedAt.UTC().Format(time.RFC3339), formatOptionalTime(run.FinishedAt))
	if err != nil {
		return workflows.WorkflowRun{}, err
	}
	return r.GetWorkflowRun(run.ID)
}
func (r WorkflowRunRepository) UpdateWorkflowRun(run workflows.WorkflowRun) (workflows.WorkflowRun, error) {
	_, err := r.db.Exec(`UPDATE workflow_runs SET status = ?, finished_at = ? WHERE id = ?`, string(run.Status), formatOptionalTime(run.FinishedAt), run.ID)
	if err != nil {
		return workflows.WorkflowRun{}, err
	}
	return r.GetWorkflowRun(run.ID)
}
func (r WorkflowRunRepository) GetWorkflowRun(id string) (workflows.WorkflowRun, error) {
	row := r.db.QueryRow(`SELECT id, workflow_type, status, started_at, finished_at FROM workflow_runs WHERE id = ?`, id)
	var run workflows.WorkflowRun
	var typ, status, started, finished string
	if err := row.Scan(&run.ID, &typ, &status, &started, &finished); err != nil {
		return run, err
	}
	run.WorkflowType = workflows.WorkflowType(typ)
	run.Status = workflows.RunStatus(status)
	run.StartedAt = parseRFC3339(started)
	if finished != "" {
		run.FinishedAt = parseRFC3339(finished)
	}
	return run, nil
}
func formatOptionalTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func countRows(db *sql.DB, table string) (int, error) {
	var n int
	err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&n)
	return n, err
}
