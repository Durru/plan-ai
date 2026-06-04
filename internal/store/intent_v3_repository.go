package store

import (
	"database/sql"

	"github.com/plan-ai/plan-ai/internal/intentv3"
)

type IntentV3ProductIntentRepository struct{ db *sql.DB }

func NewIntentV3Repository(db *sql.DB) IntentV3ProductIntentRepository {
	return IntentV3ProductIntentRepository{db: db}
}

var _ intentv3.ProductIntentRepository = IntentV3ProductIntentRepository{}

func (r IntentV3ProductIntentRepository) SaveProductIntent(pi intentv3.ProductIntent) (intentv3.ProductIntent, error) {
	c, u := timestamps(pi.CreatedAt, pi.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO intent_v3_product_intents (id, project_id, description, expected_outcome, desired_experience, desired_result, user_expectations, non_expectations, success_definition, failure_definition, status, discovery_result_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, description=excluded.description, expected_outcome=excluded.expected_outcome, desired_experience=excluded.desired_experience, desired_result=excluded.desired_result, user_expectations=excluded.user_expectations, non_expectations=excluded.non_expectations, success_definition=excluded.success_definition, failure_definition=excluded.failure_definition, status=excluded.status, discovery_result_id=excluded.discovery_result_id, updated_at=excluded.updated_at`,
		pi.ID, pi.ProjectID, pi.Description, pi.ExpectedOutcome, pi.DesiredExperience, pi.DesiredResult,
		mustJSON(pi.UserExpectations), mustJSON(pi.NonExpectations),
		pi.SuccessDefinition, pi.FailureDefinition, string(pi.Status), pi.DiscoveryResultID,
		c, u)
	if err != nil {
		return intentv3.ProductIntent{}, err
	}
	pi.CreatedAt = parseRFC3339(c)
	pi.UpdatedAt = parseRFC3339(u)
	return pi, nil
}

func (r IntentV3ProductIntentRepository) UpdateProductIntent(pi intentv3.ProductIntent) (intentv3.ProductIntent, error) {
	u := nowString()
	_, err := r.db.Exec(`UPDATE intent_v3_product_intents SET description=?, expected_outcome=?, desired_experience=?, desired_result=?, user_expectations=?, non_expectations=?, success_definition=?, failure_definition=?, status=?, updated_at=? WHERE id=?`,
		pi.Description, pi.ExpectedOutcome, pi.DesiredExperience, pi.DesiredResult,
		mustJSON(pi.UserExpectations), mustJSON(pi.NonExpectations),
		pi.SuccessDefinition, pi.FailureDefinition, string(pi.Status), u, pi.ID)
	if err != nil {
		return intentv3.ProductIntent{}, err
	}
	pi.UpdatedAt = parseRFC3339(u)
	return pi, nil
}

func (r IntentV3ProductIntentRepository) GetProductIntent(id string) (intentv3.ProductIntent, error) {
	items, err := r.listIntents("WHERE id = ?", id)
	if err != nil {
		return intentv3.ProductIntent{}, err
	}
	if len(items) == 0 {
		return intentv3.ProductIntent{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r IntentV3ProductIntentRepository) ListProductIntents(projectID string) ([]intentv3.ProductIntent, error) {
	return r.listIntents("WHERE project_id = ? ORDER BY created_at DESC", projectID)
}

func (r IntentV3ProductIntentRepository) UpdateProductIntentStatus(id string, status intentv3.ProductIntentStatus) (intentv3.ProductIntent, error) {
	_, err := r.db.Exec(`UPDATE intent_v3_product_intents SET status = ?, updated_at = ? WHERE id = ?`, string(status), nowString(), id)
	if err != nil {
		return intentv3.ProductIntent{}, err
	}
	return r.GetProductIntent(id)
}

func (r IntentV3ProductIntentRepository) listIntents(where string, args ...any) ([]intentv3.ProductIntent, error) {
	rows, err := r.db.Query(`SELECT id, project_id, description, expected_outcome, desired_experience, desired_result, user_expectations, non_expectations, success_definition, failure_definition, status, discovery_result_id, created_at, updated_at FROM intent_v3_product_intents `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []intentv3.ProductIntent
	for rows.Next() {
		var pi intentv3.ProductIntent
		var ue, ne, c, u string
		var status string
		if err := rows.Scan(&pi.ID, &pi.ProjectID, &pi.Description, &pi.ExpectedOutcome, &pi.DesiredExperience, &pi.DesiredResult, &ue, &ne, &pi.SuccessDefinition, &pi.FailureDefinition, &status, &pi.DiscoveryResultID, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(ue, &pi.UserExpectations)
		decodeJSON(ne, &pi.NonExpectations)
		pi.Status = intentv3.ProductIntentStatus(status)
		pi.CreatedAt = parseRFC3339(c)
		pi.UpdatedAt = parseRFC3339(u)
		out = append(out, pi)
	}
	return out, rows.Err()
}

// ──────────────────────────────────────────────
// Discovery Result Repository
// ──────────────────────────────────────────────

type IntentV3DiscoveryResultRepository struct{ db *sql.DB }

func NewIntentV3DiscoveryResultRepository(db *sql.DB) IntentV3DiscoveryResultRepository {
	return IntentV3DiscoveryResultRepository{db: db}
}

var _ intentv3.DiscoveryResultRepository = IntentV3DiscoveryResultRepository{}

func (r IntentV3DiscoveryResultRepository) SaveDiscoveryResult(dr intentv3.DiscoveryResult) (intentv3.DiscoveryResult, error) {
	c := nowString()
	if dr.CreatedAt.IsZero() {
		dr.CreatedAt = parseRFC3339(c)
	}
	_, err := r.db.Exec(`INSERT INTO intent_v3_discovery_results (id, project_id, raw_input, detected_intent, objectives, restrictions, preferences, refs_list, expectations, classification, gaps, questions, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET project_id=excluded.project_id, raw_input=excluded.raw_input, detected_intent=excluded.detected_intent, objectives=excluded.objectives, restrictions=excluded.restrictions, preferences=excluded.preferences, refs_list=excluded.refs_list, expectations=excluded.expectations, classification=excluded.classification, gaps=excluded.gaps, questions=excluded.questions`,
		dr.ID, dr.ProjectID, dr.RawInput, dr.DetectedIntent,
		mustJSON(dr.Objectives), mustJSON(dr.Restrictions), mustJSON(dr.Preferences),
		mustJSON(dr.References), mustJSON(dr.Expectations),
		dr.Classification, mustJSON(dr.Gaps), mustJSON(dr.Questions), c)
	if err != nil {
		return intentv3.DiscoveryResult{}, err
	}
	dr.CreatedAt = parseRFC3339(c)
	return dr, nil
}

func (r IntentV3DiscoveryResultRepository) GetDiscoveryResult(id string) (intentv3.DiscoveryResult, error) {
	items, err := r.listResults("WHERE id = ?", id)
	if err != nil {
		return intentv3.DiscoveryResult{}, err
	}
	if len(items) == 0 {
		return intentv3.DiscoveryResult{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r IntentV3DiscoveryResultRepository) ListDiscoveryResults(projectID string) ([]intentv3.DiscoveryResult, error) {
	return r.listResults("WHERE project_id = ? ORDER BY created_at DESC", projectID)
}

func (r IntentV3DiscoveryResultRepository) listResults(where string, args ...any) ([]intentv3.DiscoveryResult, error) {
	rows, err := r.db.Query(`SELECT id, project_id, raw_input, detected_intent, objectives, restrictions, preferences, refs_list, expectations, classification, gaps, questions, created_at FROM intent_v3_discovery_results `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []intentv3.DiscoveryResult
	for rows.Next() {
		var dr intentv3.DiscoveryResult
		var obj, rest, pref, ref, exp, gaps, questions, c string
		if err := rows.Scan(&dr.ID, &dr.ProjectID, &dr.RawInput, &dr.DetectedIntent, &obj, &rest, &pref, &ref, &exp, &dr.Classification, &gaps, &questions, &c); err != nil {
			return nil, err
		}
		decodeJSON(obj, &dr.Objectives)
		decodeJSON(rest, &dr.Restrictions)
		decodeJSON(pref, &dr.Preferences)
		decodeJSON(ref, &dr.References)
		decodeJSON(exp, &dr.Expectations)
		decodeJSON(gaps, &dr.Gaps)
		decodeJSON(questions, &dr.Questions)
		dr.CreatedAt = parseRFC3339(c)
		out = append(out, dr)
	}
	return out, rows.Err()
}

// ──────────────────────────────────────────────
// Migration 0040 constant
// ──────────────────────────────────────────────

const projectV3ProductIntentSQL = `
-- 0040_v3_product_intent: Phase 51 Product Intent Engine + Phase 52 Discovery
CREATE TABLE IF NOT EXISTS intent_v3_product_intents (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  description TEXT NOT NULL,
  expected_outcome TEXT NOT NULL DEFAULT '',
  desired_experience TEXT NOT NULL DEFAULT '',
  desired_result TEXT NOT NULL DEFAULT '',
  user_expectations TEXT NOT NULL DEFAULT '[]',
  non_expectations TEXT NOT NULL DEFAULT '[]',
  success_definition TEXT NOT NULL DEFAULT '',
  failure_definition TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  discovery_result_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intent_v3_product_intents_project ON intent_v3_product_intents(project_id);
CREATE INDEX IF NOT EXISTS idx_intent_v3_product_intents_status ON intent_v3_product_intents(status);

CREATE TABLE IF NOT EXISTS intent_v3_discovery_results (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  raw_input TEXT NOT NULL,
  detected_intent TEXT NOT NULL DEFAULT '',
  objectives TEXT NOT NULL DEFAULT '[]',
  restrictions TEXT NOT NULL DEFAULT '[]',
  preferences TEXT NOT NULL DEFAULT '[]',
  refs_list TEXT NOT NULL DEFAULT '[]',
  expectations TEXT NOT NULL DEFAULT '[]',
  classification TEXT NOT NULL DEFAULT '',
  gaps TEXT NOT NULL DEFAULT '[]',
  questions TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intent_v3_discovery_results_project ON intent_v3_discovery_results(project_id);
`
