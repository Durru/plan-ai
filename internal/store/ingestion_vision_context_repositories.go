package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	approvedcontext "github.com/plan-ai/plan-ai/internal/context"
	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/ingestion"
	"github.com/plan-ai/plan-ai/internal/vision"
)

type IngestionRepository struct{ db *sql.DB }
type VisionDraftRepository struct{ db *sql.DB }
type ApprovedContextRepository struct{ db *sql.DB }

func NewIngestionRepository(db *sql.DB) IngestionRepository     { return IngestionRepository{db: db} }
func NewVisionDraftRepository(db *sql.DB) VisionDraftRepository { return VisionDraftRepository{db: db} }
func NewApprovedContextRepository(db *sql.DB) ApprovedContextRepository {
	return ApprovedContextRepository{db: db}
}

var _ ingestion.Repository = IngestionRepository{}
var _ vision.Repository = VisionDraftRepository{}
var _ approvedcontext.Repository = ApprovedContextRepository{}

func (r IngestionRepository) CreateRawInput(input ingestion.RawInput) (ingestion.RawInput, error) {
	if input.ID == "" {
		input.ID = domain.NewID("raw")
	}
	c, u := timestamps(input.CreatedAt, input.UpdatedAt)
	metadata := encodeStringMap(input.Metadata)
	_, err := r.db.Exec(`INSERT INTO raw_inputs (id, project_id, source_type, content, raw_content, metadata, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, source_type=excluded.source_type, content=excluded.content, raw_content=excluded.raw_content, metadata=excluded.metadata, updated_at=excluded.updated_at`, input.ID, input.ProjectID, input.SourceType, input.RawContent, input.RawContent, metadata, c, u)
	if err != nil {
		return ingestion.RawInput{}, err
	}
	input.CreatedAt, input.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return input, nil
}

func (r IngestionRepository) CreateIngestedSource(source ingestion.IngestedSource) (ingestion.IngestedSource, error) {
	if source.ID == "" {
		source.ID = domain.NewID("source")
	}
	c, u := timestamps(source.CreatedAt, source.UpdatedAt)
	metadata := encodeStringMap(source.Metadata)
	_, err := r.db.Exec(`INSERT INTO ingested_sources (id, project_id, raw_input_id, source_type, normalized_content, classification, metadata, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, 'ingested', ?, ?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, raw_input_id=excluded.raw_input_id, source_type=excluded.source_type, normalized_content=excluded.normalized_content, classification=excluded.classification, metadata=excluded.metadata, updated_at=excluded.updated_at`, source.ID, source.ProjectID, source.RawInputID, source.SourceType, source.NormalizedContent, source.Classification, metadata, c, u)
	if err != nil {
		return ingestion.IngestedSource{}, err
	}
	source.CreatedAt, source.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return source, nil
}

func (r IngestionRepository) GetRawInput(id string) (ingestion.RawInput, error) {
	var input ingestion.RawInput
	var metadata, c, u string
	err := r.db.QueryRow(`SELECT id, project_id, source_type, COALESCE(raw_content, content), metadata, created_at, updated_at FROM raw_inputs WHERE id = ?`, id).Scan(&input.ID, &input.ProjectID, &input.SourceType, &input.RawContent, &metadata, &c, &u)
	if err != nil {
		return input, err
	}
	input.Metadata = decodeStringMap(metadata)
	input.CreatedAt = parseRFC3339(c)
	input.UpdatedAt = parseRFC3339(u)
	return input, nil
}

func (r IngestionRepository) GetIngestedSource(id string) (ingestion.IngestedSource, error) {
	items, err := r.listSources(`WHERE id = ?`, id)
	if err != nil {
		return ingestion.IngestedSource{}, err
	}
	if len(items) == 0 {
		return ingestion.IngestedSource{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r IngestionRepository) ListIngestedSources(projectID string) ([]ingestion.IngestedSource, error) {
	return r.listSources(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}

func (r IngestionRepository) listSources(where string, args ...any) ([]ingestion.IngestedSource, error) {
	rows, err := r.db.Query(`SELECT id, project_id, raw_input_id, source_type, normalized_content, classification, metadata, created_at, updated_at FROM ingested_sources `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ingestion.IngestedSource
	for rows.Next() {
		var source ingestion.IngestedSource
		var metadata, c, u string
		if err := rows.Scan(&source.ID, &source.ProjectID, &source.RawInputID, &source.SourceType, &source.NormalizedContent, &source.Classification, &metadata, &c, &u); err != nil {
			return nil, err
		}
		source.Metadata = decodeStringMap(metadata)
		source.CreatedAt = parseRFC3339(c)
		source.UpdatedAt = parseRFC3339(u)
		out = append(out, source)
	}
	return out, rows.Err()
}

func (r VisionDraftRepository) SaveVision(draft vision.Draft) (vision.Draft, error) {
	if draft.ID == "" {
		draft.ID = domain.NewID("vision")
	}
	c, u := timestamps(draft.CreatedAt, draft.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO visions (id, project_id, title, summary, target_users, expected_outcome, functional_goals, ux_goals, business_goals, constraints, assumptions, missing_information, visual_references, success_criteria, approved, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET title=excluded.title, summary=excluded.summary, target_users=excluded.target_users, expected_outcome=excluded.expected_outcome, functional_goals=excluded.functional_goals, ux_goals=excluded.ux_goals, business_goals=excluded.business_goals, constraints=excluded.constraints, assumptions=excluded.assumptions, missing_information=excluded.missing_information, visual_references=excluded.visual_references, success_criteria=excluded.success_criteria, approved=excluded.approved, updated_at=excluded.updated_at`, draft.ID, draft.ProjectID, draft.Title, draft.Summary, jsonListLocal(draft.TargetUsers), draft.ExpectedOutcome, jsonListLocal(draft.FunctionalGoals), jsonListLocal(draft.UXGoals), jsonListLocal(draft.BusinessGoals), jsonListLocal(draft.Constraints), jsonListLocal(draft.Assumptions), jsonListLocal(draft.MissingInformation), jsonListLocal(draft.VisualReferences), jsonListLocal(draft.SuccessCriteria), boolToInt(draft.Approved), c, u)
	if err != nil {
		return vision.Draft{}, err
	}
	draft.CreatedAt, draft.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return draft, nil
}

func (r VisionDraftRepository) GetVision(id string) (vision.Draft, error) {
	items, err := r.listVisions(`WHERE id = ?`, id)
	if err != nil {
		return vision.Draft{}, err
	}
	if len(items) == 0 {
		return vision.Draft{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r VisionDraftRepository) ListVisions(projectID string) ([]vision.Draft, error) {
	return r.listVisions(`WHERE project_id = ? ORDER BY created_at, id`, projectID)
}

func (r VisionDraftRepository) ApproveVision(id string) (vision.Draft, error) {
	_, err := r.db.Exec(`UPDATE visions SET approved = 1, updated_at = ? WHERE id = ?`, nowString(), id)
	if err != nil {
		return vision.Draft{}, err
	}
	return r.GetVision(id)
}

func (r VisionDraftRepository) listVisions(where string, args ...any) ([]vision.Draft, error) {
	rows, err := r.db.Query(`SELECT id, project_id, title, summary, target_users, expected_outcome, functional_goals, ux_goals, business_goals, constraints, assumptions, missing_information, visual_references, success_criteria, approved, created_at, updated_at FROM visions `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []vision.Draft
	for rows.Next() {
		var d vision.Draft
		var targetUsers, functionalGoals, uxGoals, businessGoals, constraints, assumptions, missing, visual, success, c, u string
		var approved int
		if err := rows.Scan(&d.ID, &d.ProjectID, &d.Title, &d.Summary, &targetUsers, &d.ExpectedOutcome, &functionalGoals, &uxGoals, &businessGoals, &constraints, &assumptions, &missing, &visual, &success, &approved, &c, &u); err != nil {
			return nil, err
		}
		d.TargetUsers = scanJSONListLocal(targetUsers)
		d.FunctionalGoals = scanJSONListLocal(functionalGoals)
		d.UXGoals = scanJSONListLocal(uxGoals)
		d.BusinessGoals = scanJSONListLocal(businessGoals)
		d.Constraints = scanJSONListLocal(constraints)
		d.Assumptions = scanJSONListLocal(assumptions)
		d.MissingInformation = scanJSONListLocal(missing)
		d.VisualReferences = scanJSONListLocal(visual)
		d.SuccessCriteria = scanJSONListLocal(success)
		d.Approved = approved != 0
		d.CreatedAt = parseRFC3339(c)
		d.UpdatedAt = parseRFC3339(u)
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r ApprovedContextRepository) StoreApproved(item approvedcontext.ApprovedItem) (approvedcontext.ApprovedItem, error) {
	table, err := approvedTable(item.Type)
	if err != nil {
		return approvedcontext.ApprovedItem{}, err
	}
	if existing, err := r.findDuplicate(table, item.ProjectID, item.Content); err == nil {
		return existing, nil
	} else if err != sql.ErrNoRows {
		return approvedcontext.ApprovedItem{}, err
	}
	if item.ID == "" {
		item.ID = domain.NewID(string(item.Type))
	}
	item.State = approvedcontext.StateApproved
	c, u := timestamps(item.CreatedAt, item.UpdatedAt)
	_, err = r.db.Exec(`INSERT INTO `+table+` (id, project_id, source_id, content, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, item.ID, item.ProjectID, item.SourceID, item.Content, item.State, c, u)
	if err != nil {
		return approvedcontext.ApprovedItem{}, err
	}
	item.CreatedAt, item.UpdatedAt = parseRFC3339(c), parseRFC3339(u)
	return item, nil
}

func (r ApprovedContextRepository) GetApproved(itemType approvedcontext.ApprovedType, id string) (approvedcontext.ApprovedItem, error) {
	table, err := approvedTable(itemType)
	if err != nil {
		return approvedcontext.ApprovedItem{}, err
	}
	return r.scanApproved(table, itemType, `WHERE id = ?`, id)
}

func (r ApprovedContextRepository) ListApproved(projectID string, itemType approvedcontext.ApprovedType) ([]approvedcontext.ApprovedItem, error) {
	if itemType != "" {
		table, err := approvedTable(itemType)
		if err != nil {
			return nil, err
		}
		return r.listApproved(table, itemType, `WHERE project_id = ? ORDER BY created_at, id`, projectID)
	}
	var all []approvedcontext.ApprovedItem
	for _, typ := range approvedTypes() {
		table, _ := approvedTable(typ)
		items, err := r.listApproved(table, typ, `WHERE project_id = ? ORDER BY created_at, id`, projectID)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
	}
	return all, nil
}

func (r ApprovedContextRepository) FindApproved(projectID string, itemType approvedcontext.ApprovedType, query string) ([]approvedcontext.ApprovedItem, error) {
	table, err := approvedTable(itemType)
	if err != nil {
		return nil, err
	}
	return r.listApproved(table, itemType, `WHERE project_id = ? AND content LIKE ? ORDER BY created_at, id`, projectID, "%"+query+"%")
}

func (r ApprovedContextRepository) findDuplicate(table, projectID, content string) (approvedcontext.ApprovedItem, error) {
	typ := typeFromTable(table)
	return r.scanApproved(table, typ, `WHERE project_id = ? AND lower(content) = lower(?)`, projectID, strings.TrimSpace(content))
}

func (r ApprovedContextRepository) scanApproved(table string, typ approvedcontext.ApprovedType, where string, args ...any) (approvedcontext.ApprovedItem, error) {
	items, err := r.listApproved(table, typ, where, args...)
	if err != nil {
		return approvedcontext.ApprovedItem{}, err
	}
	if len(items) == 0 {
		return approvedcontext.ApprovedItem{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r ApprovedContextRepository) listApproved(table string, typ approvedcontext.ApprovedType, where string, args ...any) ([]approvedcontext.ApprovedItem, error) {
	rows, err := r.db.Query(`SELECT id, project_id, source_id, content, state, created_at, updated_at FROM `+table+` `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []approvedcontext.ApprovedItem
	for rows.Next() {
		var item approvedcontext.ApprovedItem
		var c, u string
		item.Type = typ
		if err := rows.Scan(&item.ID, &item.ProjectID, &item.SourceID, &item.Content, &item.State, &c, &u); err != nil {
			return nil, err
		}
		item.CreatedAt = parseRFC3339(c)
		item.UpdatedAt = parseRFC3339(u)
		out = append(out, item)
	}
	return out, rows.Err()
}

func approvedTypes() []approvedcontext.ApprovedType {
	return []approvedcontext.ApprovedType{approvedcontext.TypeRequirement, approvedcontext.TypeConstraint, approvedcontext.TypeDecision, approvedcontext.TypePreference, approvedcontext.TypeGoal, approvedcontext.TypeReference}
}

func approvedTable(typ approvedcontext.ApprovedType) (string, error) {
	switch typ {
	case approvedcontext.TypeRequirement:
		return "approved_requirements", nil
	case approvedcontext.TypeConstraint:
		return "approved_constraints", nil
	case approvedcontext.TypeDecision:
		return "approved_decisions", nil
	case approvedcontext.TypePreference:
		return "approved_preferences", nil
	case approvedcontext.TypeGoal:
		return "approved_goals", nil
	case approvedcontext.TypeReference:
		return "approved_references", nil
	default:
		return "", fmt.Errorf("unknown approved type %q", typ)
	}
}

func typeFromTable(table string) approvedcontext.ApprovedType {
	switch table {
	case "approved_requirements":
		return approvedcontext.TypeRequirement
	case "approved_constraints":
		return approvedcontext.TypeConstraint
	case "approved_decisions":
		return approvedcontext.TypeDecision
	case "approved_preferences":
		return approvedcontext.TypePreference
	case "approved_goals":
		return approvedcontext.TypeGoal
	case "approved_references":
		return approvedcontext.TypeReference
	default:
		return ""
	}
}

func timestamps(created, updated time.Time) (string, string) {
	n := time.Now().UTC()
	if created.IsZero() {
		created = n
	}
	if updated.IsZero() {
		updated = created
	}
	return created.UTC().Format(time.RFC3339), updated.UTC().Format(time.RFC3339)
}
func nowString() string                   { return time.Now().UTC().Format(time.RFC3339) }
func parseRFC3339(value string) time.Time { t, _ := time.Parse(time.RFC3339, value); return t }
func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func encodeStringMap(values map[string]string) string {
	if values == nil {
		values = map[string]string{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}
func decodeStringMap(value string) map[string]string {
	var out map[string]string
	_ = json.Unmarshal([]byte(value), &out)
	if out == nil {
		out = map[string]string{}
	}
	return out
}
func jsonListLocal(values []string) string { data, _ := json.Marshal(values); return string(data) }
func scanJSONListLocal(value string) []string {
	var out []string
	_ = json.Unmarshal([]byte(value), &out)
	return out
}
