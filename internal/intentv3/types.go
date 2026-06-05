package intentv3

import "time"

// ──────────────────────────────────────────────
// Product Intent Status (Phase 51 lifecycle)
// ──────────────────────────────────────────────

type ProductIntentStatus string

const (
	StatusDraft           ProductIntentStatus = "draft"
	StatusPendingApproval ProductIntentStatus = "pending_approval"
	StatusApproved        ProductIntentStatus = "approved"
	StatusSuperseded      ProductIntentStatus = "superseded"
	StatusArchived        ProductIntentStatus = "archived"
)

// ──────────────────────────────────────────────
// Product Intent (Phase 51)
// ──────────────────────────────────────────────

type ProductIntent struct {
	ID                string
	ProjectID         string
	Description       string
	ExpectedOutcome   string
	DesiredExperience string
	DesiredResult     string
	UserExpectations  []string
	NonExpectations   []string
	SuccessDefinition string
	FailureDefinition string
	Status            ProductIntentStatus
	DiscoveryResultID string // link to Phase 52 result, empty if manual
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// ──────────────────────────────────────────────
// Discovery Result (Phase 52)
// ──────────────────────────────────────────────

type DiscoveryResult struct {
	ID             string
	ProjectID      string
	RawInput       string
	DetectedIntent string
	Objectives     []string
	Restrictions   []string
	Preferences    []string
	References     []string
	Expectations   []string
	Classification string
	Gaps           []string
	Questions      []string
	CreatedAt      time.Time
}

// ──────────────────────────────────────────────
// Repository interface
// ──────────────────────────────────────────────

type ProductIntentRepository interface {
	SaveProductIntent(ProductIntent) (ProductIntent, error)
	GetProductIntent(id string) (ProductIntent, error)
	ListProductIntents(projectID string) ([]ProductIntent, error)
	UpdateProductIntentStatus(id string, status ProductIntentStatus) (ProductIntent, error)
	UpdateProductIntent(ProductIntent) (ProductIntent, error)
}

type DiscoveryResultRepository interface {
	SaveDiscoveryResult(DiscoveryResult) (DiscoveryResult, error)
	GetDiscoveryResult(id string) (DiscoveryResult, error)
	ListDiscoveryResults(projectID string) ([]DiscoveryResult, error)
}

// ──────────────────────────────────────────────
// Valid status transitions
// ──────────────────────────────────────────────

var validTransitions = map[ProductIntentStatus][]ProductIntentStatus{
	StatusDraft:           {StatusPendingApproval, StatusArchived},
	StatusPendingApproval: {StatusApproved, StatusDraft, StatusArchived},
	StatusApproved:        {StatusSuperseded, StatusArchived},
	StatusSuperseded:      {StatusArchived},
	StatusArchived:        {},
}

func IsValidTransition(from, to ProductIntentStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, a := range allowed {
		if a == to {
			return true
		}
	}
	return false
}
