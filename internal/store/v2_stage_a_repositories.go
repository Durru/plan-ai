package store

import (
	"database/sql"

	"github.com/Durru/plan-ai/internal/approval"
	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/requirements"
	"github.com/Durru/plan-ai/internal/vision"
)

type VisionDocumentRepository struct{ db *sql.DB }
type ApprovalRecordRepository struct{ db *sql.DB }
type RequirementCandidateRepository struct{ db *sql.DB }

func NewVisionDocumentRepository(db *sql.DB) VisionDocumentRepository {
	return VisionDocumentRepository{db: db}
}
func NewApprovalRecordRepository(db *sql.DB) ApprovalRecordRepository {
	return ApprovalRecordRepository{db: db}
}
func NewRequirementCandidateRepository(db *sql.DB) RequirementCandidateRepository {
	return RequirementCandidateRepository{db: db}
}

var _ vision.DocumentRepository = VisionDocumentRepository{}
var _ approval.Repository = ApprovalRecordRepository{}
var _ requirements.Repository = RequirementCandidateRepository{}

func (r VisionDocumentRepository) SaveDocument(doc vision.Document) (vision.Document, error) {
	if doc.ID == "" {
		doc.ID = domain.NewID("visiondoc")
	}
	c, u := timestamps(doc.CreatedAt, doc.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO vision_documents (id, project_id, intent_profile_id, source, functional_vision, visual_vision, technical_vision, operational_vision, business_vision, status, approved, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET intent_profile_id=excluded.intent_profile_id, source=excluded.source, functional_vision=excluded.functional_vision, visual_vision=excluded.visual_vision, technical_vision=excluded.technical_vision, operational_vision=excluded.operational_vision, business_vision=excluded.business_vision, status=excluded.status, approved=excluded.approved, updated_at=excluded.updated_at`, doc.ID, doc.ProjectID, doc.IntentProfileID, doc.Source, doc.FunctionalVision, doc.VisualVision, doc.TechnicalVision, doc.OperationalVision, doc.BusinessVision, doc.Status, boolToInt(doc.Approved), c, u)
	if err != nil {
		return vision.Document{}, err
	}
	doc.CreatedAt, doc.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return doc, nil
}

func (r VisionDocumentRepository) GetDocument(id string) (vision.Document, error) {
	items, err := r.listDocuments(`WHERE id = ?`, id)
	if err != nil {
		return vision.Document{}, err
	}
	if len(items) == 0 {
		return vision.Document{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r VisionDocumentRepository) ListDocuments(projectID string) ([]vision.Document, error) {
	return r.listDocuments(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}

func (r VisionDocumentRepository) ApproveDocument(id string) (vision.Document, error) {
	_, err := r.db.Exec(`UPDATE vision_documents SET status = ?, approved = 1, updated_at = ? WHERE id = ?`, vision.DocumentApproved, nowString(), id)
	if err != nil {
		return vision.Document{}, err
	}
	return r.GetDocument(id)
}

func (r VisionDocumentRepository) listDocuments(where string, args ...any) ([]vision.Document, error) {
	rows, err := r.db.Query(`SELECT id, project_id, intent_profile_id, source, functional_vision, visual_vision, technical_vision, operational_vision, business_vision, status, approved, created_at, updated_at FROM vision_documents `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []vision.Document
	for rows.Next() {
		var doc vision.Document
		var approved int
		var c, u string
		if err := rows.Scan(&doc.ID, &doc.ProjectID, &doc.IntentProfileID, &doc.Source, &doc.FunctionalVision, &doc.VisualVision, &doc.TechnicalVision, &doc.OperationalVision, &doc.BusinessVision, &doc.Status, &approved, &c, &u); err != nil {
			return nil, err
		}
		doc.Approved = approved != 0
		doc.CreatedAt, doc.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, doc)
	}
	return out, rows.Err()
}

func (r ApprovalRecordRepository) SaveRecord(record approval.Record) (approval.Record, error) {
	if record.ID == "" {
		record.ID = domain.NewID("approval")
	}
	c, u := timestamps(record.CreatedAt, record.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO approval_records (id, project_id, target_type, target_id, state, reason, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET state=excluded.state, reason=excluded.reason, updated_at=excluded.updated_at`, record.ID, record.ProjectID, record.TargetType, record.TargetID, record.State, record.Reason, c, u)
	if err != nil {
		return approval.Record{}, err
	}
	record.CreatedAt, record.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return record, nil
}

func (r ApprovalRecordRepository) ListRecords(projectID string) ([]approval.Record, error) {
	rows, err := r.db.Query(`SELECT id, project_id, target_type, target_id, state, reason, created_at, updated_at FROM approval_records WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []approval.Record
	for rows.Next() {
		var item approval.Record
		var c, u string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.TargetType, &item.TargetID, &item.State, &item.Reason, &c, &u); err != nil {
			return nil, err
		}
		item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r ApprovalRecordRepository) ApproveRecord(id string) (approval.Record, error) {
	return r.updateRecord(id, approval.StateApproved, "")
}

func (r ApprovalRecordRepository) RejectRecord(id, reason string) (approval.Record, error) {
	return r.updateRecord(id, approval.StateRejected, reason)
}

func (r ApprovalRecordRepository) updateRecord(id string, state approval.State, reason string) (approval.Record, error) {
	_, err := r.db.Exec(`UPDATE approval_records SET state = ?, reason = ?, updated_at = ? WHERE id = ?`, state, reason, nowString(), id)
	if err != nil {
		return approval.Record{}, err
	}
	rows, err := r.db.Query(`SELECT id, project_id, target_type, target_id, state, reason, created_at, updated_at FROM approval_records WHERE id = ?`, id)
	if err != nil {
		return approval.Record{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return approval.Record{}, sql.ErrNoRows
	}
	var item approval.Record
	var c, u string
	if err := rows.Scan(&item.ID, &item.ProjectID, &item.TargetType, &item.TargetID, &item.State, &item.Reason, &c, &u); err != nil {
		return approval.Record{}, err
	}
	item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return item, rows.Err()
}

func (r RequirementCandidateRepository) SaveCandidates(items []requirements.Candidate) ([]requirements.Candidate, error) {
	out := make([]requirements.Candidate, 0, len(items))
	for _, item := range items {
		if item.ID == "" {
			item.ID = domain.NewID("reqcand")
		}
		c, u := timestamps(item.CreatedAt, item.UpdatedAt)
		_, err := r.db.Exec(`INSERT INTO requirement_candidates (id, project_id, source, name, description, reason, dependencies, ambiguities, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET description=excluded.description, reason=excluded.reason, dependencies=excluded.dependencies, ambiguities=excluded.ambiguities, state=excluded.state, updated_at=excluded.updated_at`, item.ID, item.ProjectID, item.Source, item.Name, item.Description, item.Reason, mustJSON(item.Dependencies), mustJSON(item.Ambiguities), item.State, c, u)
		if err != nil {
			return nil, err
		}
		item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, item)
	}
	return out, nil
}

func (r RequirementCandidateRepository) ListCandidates(projectID string) ([]requirements.Candidate, error) {
	rows, err := r.db.Query(`SELECT id, project_id, source, name, description, reason, dependencies, ambiguities, state, created_at, updated_at FROM requirement_candidates WHERE project_id = ? ORDER BY created_at, id`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []requirements.Candidate
	for rows.Next() {
		var item requirements.Candidate
		var deps, amb, c, u string
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.Source, &item.Name, &item.Description, &item.Reason, &deps, &amb, &item.State, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(deps, &item.Dependencies)
		decodeJSON(amb, &item.Ambiguities)
		item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r RequirementCandidateRepository) ApproveCandidate(id string) (requirements.Candidate, error) {
	_, err := r.db.Exec(`UPDATE requirement_candidates SET state = ?, updated_at = ? WHERE id = ?`, requirements.StateApproved, nowString(), id)
	if err != nil {
		return requirements.Candidate{}, err
	}
	rows, err := r.db.Query(`SELECT id, project_id, source, name, description, reason, dependencies, ambiguities, state, created_at, updated_at FROM requirement_candidates WHERE id = ?`, id)
	if err != nil {
		return requirements.Candidate{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return requirements.Candidate{}, sql.ErrNoRows
	}
	var item requirements.Candidate
	var deps, amb, c, u string
	if err := rows.Scan(&item.ID, &item.ProjectID, &item.Source, &item.Name, &item.Description, &item.Reason, &deps, &amb, &item.State, &c, &u); err != nil {
		return requirements.Candidate{}, err
	}
	decodeJSON(deps, &item.Dependencies)
	decodeJSON(amb, &item.Ambiguities)
	item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return item, rows.Err()
}
