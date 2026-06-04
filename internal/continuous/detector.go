package continuous

import (
	"database/sql"
)

// Detector detects events that may trigger plan updates.
type Detector struct {
	db *sql.DB
}

// NewDetector creates a new Detector.
func NewDetector(db *sql.DB) *Detector {
	return &Detector{db: db}
}

// Detect checks for new events since the given event count baseline.
func (d *Detector) Detect(projectID string) ([]ContinuousEvent, error) {
	rows, err := d.db.Query(
		`SELECT id, project_id, event_type, summary, COALESCE(details, ''), COALESCE(source, ''), created_at
		 FROM continuous_events WHERE project_id = ? ORDER BY created_at DESC LIMIT 50`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []ContinuousEvent
	for rows.Next() {
		var ev ContinuousEvent
		if err := rows.Scan(&ev.ID, &ev.ProjectID, &ev.EventType, &ev.Summary, &ev.Details, &ev.Source, &ev.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, rows.Err()
}

// DetectOutdatedPlans checks if any plans may be outdated based on recent events.
func (d *Detector) DetectOutdatedPlans(projectID string) ([]string, error) {
	// Look for decision changes or new approved context that might affect plans
	rows, err := d.db.Query(
		`SELECT DISTINCT event_type FROM continuous_events
		 WHERE project_id = ? AND event_type IN (?, ?, ?, ?)
		 ORDER BY created_at DESC LIMIT 10`,
		projectID, EventDecisionChanged, EventNewApprovedContext, EventNewKnowledge, EventChangeRequestCreated)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outdated []string
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			continue
		}
		outdated = append(outdated, eventType)
	}
	return outdated, rows.Err()
}
