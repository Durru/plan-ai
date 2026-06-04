package domain

import "time"

// Research is an investigation into a topic that produces findings,
// sources, and conclusions to inform project decisions. It is
// provisional by nature — not yet promoted to reusable knowledge.
type Research struct {
	ID         string
	ProjectID  string
	Topic      string
	Objective  string
	Summary    string
	Confidence float64
	Date       time.Time
	Category   KnowledgeCategory // backward compat: maps to old ResearchEntry.Category
	Status     ResearchStatus    // backward compat: maps to old ResearchEntry.Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ResearchSource is a reference or citation that informed a Research
// entry. Multiple sources can be attached to a single Research entity.
type ResearchSource struct {
	ID         string
	ResearchID string
	URL        string
	Title      string
	SourceType string
	CreatedAt  time.Time
}
