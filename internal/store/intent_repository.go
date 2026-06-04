package store

import (
	"database/sql"
	"encoding/json"

	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/intent"
)

type IntentProfileRepository struct{ db *sql.DB }

func NewIntentProfileRepository(db *sql.DB) IntentProfileRepository {
	return IntentProfileRepository{db: db}
}

var _ intent.Repository = IntentProfileRepository{}

func (r IntentProfileRepository) SaveProfile(profile intent.Profile) (intent.Profile, error) {
	if profile.ID == "" {
		profile.ID = domain.NewID("intent")
	}
	c, u := timestamps(profile.CreatedAt, profile.UpdatedAt)
	_, err := r.db.Exec(`INSERT INTO intent_profiles (id, project_id, source, primary_intent, secondary_goals, constraints_json, expectations, success_criteria, priorities, status, approved, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO UPDATE SET source=excluded.source, primary_intent=excluded.primary_intent, secondary_goals=excluded.secondary_goals, constraints_json=excluded.constraints_json, expectations=excluded.expectations, success_criteria=excluded.success_criteria, priorities=excluded.priorities, status=excluded.status, approved=excluded.approved, updated_at=excluded.updated_at`, profile.ID, profile.ProjectID, profile.Source, mustJSON(profile.PrimaryIntent), mustJSON(profile.SecondaryGoals), mustJSON(profile.Constraints), mustJSON(profile.Expectations), mustJSON(profile.SuccessCriteria), mustJSON(profile.Priorities), profile.Status, boolToInt(profile.Approved), c, u)
	if err != nil {
		return intent.Profile{}, err
	}
	profile.CreatedAt = parseRFC3339(c)
	profile.UpdatedAt = parseRFC3339(u)
	return profile, nil
}

func (r IntentProfileRepository) GetProfile(id string) (intent.Profile, error) {
	items, err := r.listProfiles(`WHERE id = ?`, id)
	if err != nil {
		return intent.Profile{}, err
	}
	if len(items) == 0 {
		return intent.Profile{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r IntentProfileRepository) LatestProfile(projectID string) (intent.Profile, error) {
	items, err := r.listProfiles(`WHERE project_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`, projectID)
	if err != nil {
		return intent.Profile{}, err
	}
	if len(items) == 0 {
		return intent.Profile{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r IntentProfileRepository) ApproveProfile(id string) (intent.Profile, error) {
	_, err := r.db.Exec(`UPDATE intent_profiles SET status = ?, approved = 1, updated_at = ? WHERE id = ?`, intent.StatusApproved, nowString(), id)
	if err != nil {
		return intent.Profile{}, err
	}
	return r.GetProfile(id)
}

func (r IntentProfileRepository) listProfiles(where string, args ...any) ([]intent.Profile, error) {
	rows, err := r.db.Query(`SELECT id, project_id, source, primary_intent, secondary_goals, constraints_json, expectations, success_criteria, priorities, status, approved, created_at, updated_at FROM intent_profiles `+where, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []intent.Profile
	for rows.Next() {
		var profile intent.Profile
		var primary, goals, constraints, expectations, success, priorities, c, u string
		var approved int
		if err := rows.Scan(&profile.ID, &profile.ProjectID, &profile.Source, &primary, &goals, &constraints, &expectations, &success, &priorities, &profile.Status, &approved, &c, &u); err != nil {
			return nil, err
		}
		decodeJSON(primary, &profile.PrimaryIntent)
		decodeJSON(goals, &profile.SecondaryGoals)
		decodeJSON(constraints, &profile.Constraints)
		decodeJSON(expectations, &profile.Expectations)
		decodeJSON(success, &profile.SuccessCriteria)
		decodeJSON(priorities, &profile.Priorities)
		profile.Approved = approved != 0
		profile.CreatedAt = parseRFC3339(c)
		profile.UpdatedAt = parseRFC3339(u)
		out = append(out, profile)
	}
	return out, rows.Err()
}

func mustJSON(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		return "null"
	}
	return string(b)
}

func decodeJSON(value string, target any) {
	_ = json.Unmarshal([]byte(value), target)
}
