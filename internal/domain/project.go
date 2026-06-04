package domain

import "time"

// ProjectStatus represents the lifecycle state of a project.
type ProjectStatus string

const (
	ProjectStatusDraft     ProjectStatus = "draft"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusPaused    ProjectStatus = "paused"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusArchived  ProjectStatus = "archived"
)

// Project is the canonical project entity. Every plan-ai operation
// belongs to exactly one project.
type Project struct {
	ID          string
	Name        string
	RootPath    string
	Description string
	Status      ProjectStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ValidProjectTransitions returns allowed transitions for ProjectStatus.
// Prohibited: archived → anything, completed → draft/active.
func ValidProjectTransitions(from, to ProjectStatus) bool {
	switch from {
	case ProjectStatusDraft:
		return to == ProjectStatusActive || to == ProjectStatusArchived
	case ProjectStatusActive:
		return to == ProjectStatusPaused || to == ProjectStatusCompleted || to == ProjectStatusArchived
	case ProjectStatusPaused:
		return to == ProjectStatusActive || to == ProjectStatusCompleted || to == ProjectStatusArchived
	case ProjectStatusCompleted:
		return to == ProjectStatusArchived
	case ProjectStatusArchived:
		return false // terminal state
	default:
		return false
	}
}
