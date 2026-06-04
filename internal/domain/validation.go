package domain

import "time"

// ValidationType classifies the kind of validation to perform.
type ValidationType string

const (
	ValidationTypeManual    ValidationType = "manual"
	ValidationTypeAutomatic ValidationType = "automatic"
)

// ValidationStatus represents the outcome of a validation check.
type ValidationStatus string

const (
	ValidationPending ValidationStatus = "pending"
	ValidationPassed  ValidationStatus = "passed"
	ValidationFailed  ValidationStatus = "failed"
)

// Validation records a single check against a target entity (plan,
// phase, task, or decision). It captures what was validated, how,
// and whether it passed.
type Validation struct {
	ID         string
	TargetType ValidationTargetType
	TargetID   string
	Type       ValidationType
	Status     Status // domain.Status for backward compat; canonical values are ValidationStatus constants
	Summary    string
	Details    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
