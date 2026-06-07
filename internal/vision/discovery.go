package vision

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Discovery Types
// ──────────────────────────────────────────────

// DiscoveryStatus defines the status of a discovery session.
type DiscoveryStatus string

const (
	DiscoveryDraft      DiscoveryStatus = "draft"
	DiscoveryInProgress DiscoveryStatus = "in_progress"
	DiscoveryComplete   DiscoveryStatus = "complete"
	DiscoveryCancelled  DiscoveryStatus = "cancelled"
)

// AssumptionCategory categorizes assumptions.
type AssumptionCategory string

const (
	AssumptionTechnical  AssumptionCategory = "technical"
	AssumptionBusiness   AssumptionCategory = "business"
	AssumptionUser       AssumptionCategory = "user"
	AssumptionInfra      AssumptionCategory = "infrastructure"
	AssumptionRegulatory AssumptionCategory = "regulatory"
	AssumptionTiming     AssumptionCategory = "timing"
	AssumptionOther      AssumptionCategory = "other"
)

// AssumptionStatus defines the status of an assumption.
type AssumptionStatus string

const (
	AssumptionUnvalidated AssumptionStatus = "unvalidated"
	AssumptionValidated   AssumptionStatus = "validated"
	AssumptionInvalidated AssumptionStatus = "invalidated"
	AssumptionInProgress  AssumptionStatus = "in_progress"
)

// AmbiguityStatus defines the status of an ambiguity.
type AmbiguityStatus string

const (
	AmbiguityOpen     AmbiguityStatus = "open"
	AmbiguityResolved AmbiguityStatus = "resolved"
	AmbiguityIgnored  AmbiguityStatus = "ignored"
)

// ApprovalStatus defines the status of a vision approval.
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalChanges  ApprovalStatus = "changes_requested"
)

// DiscoverySession represents a full discovery session.
type DiscoverySession struct {
	ID         string          `json:"id"`
	ProjectID  string          `json:"project_id"`
	Status     DiscoveryStatus `json:"status"`
	Summary    string          `json:"summary"`
	RawContext string          `json:"raw_context"`
	Findings   []Finding       `json:"findings"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

// Finding represents a single discovery finding.
type Finding struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Source      string  `json:"source"`
}

// Assumption represents an identified assumption.
type Assumption struct {
	ID          string             `json:"id"`
	ProjectID   string             `json:"project_id"`
	SessionID   string             `json:"session_id"`
	Description string             `json:"description"`
	Category    AssumptionCategory `json:"category"`
	Confidence  float64            `json:"confidence"`
	Status      AssumptionStatus   `json:"status"`
	ValidatedBy string             `json:"validated_by,omitempty"`
	ValidatedAt *time.Time         `json:"validated_at,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// Ambiguity represents an identified ambiguity.
type Ambiguity struct {
	ID          string          `json:"id"`
	ProjectID   string          `json:"project_id"`
	SessionID   string          `json:"session_id"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Resolution  string          `json:"resolution,omitempty"`
	Status      AmbiguityStatus `json:"status"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

// VisionApproval represents an approval for a vision draft.
type VisionApproval struct {
	ID         string         `json:"id"`
	ProjectID  string         `json:"project_id"`
	SessionID  string         `json:"session_id"`
	VisionID   string         `json:"vision_id"`
	Status     ApprovalStatus `json:"status"`
	ApprovedBy string         `json:"approved_by,omitempty"`
	ApprovedAt *time.Time     `json:"approved_at,omitempty"`
	Feedback   string         `json:"feedback,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// ──────────────────────────────────────────────
// Discovery Repository Interface
// ──────────────────────────────────────────────

// DiscoveryRepository provides persistence for discovery entities.
type DiscoveryRepository interface {
	CreateSession(session DiscoverySession) (DiscoverySession, error)
	GetSession(id string) (DiscoverySession, error)
	ListSessions(projectID string) ([]DiscoverySession, error)
	UpdateSession(id string, status DiscoveryStatus, summary string) error

	CreateAssumption(a Assumption) (Assumption, error)
	ListAssumptions(sessionID string) ([]Assumption, error)
	UpdateAssumptionStatus(id string, status AssumptionStatus, validatedBy string) error

	CreateAmbiguity(a Ambiguity) (Ambiguity, error)
	ListAmbiguities(sessionID string) ([]Ambiguity, error)
	ResolveAmbiguity(id string, resolution string) error

	CreateApproval(a VisionApproval) (VisionApproval, error)
	GetApprovalByVision(visionID string) (VisionApproval, error)
	ApproveApproval(id string, approvedBy string, feedback string) error
	RejectApproval(id string, feedback string) error
}

// ──────────────────────────────────────────────
// Discovery Engine
// ──────────────────────────────────────────────

// DiscoveryEngine orchestrates vision discovery sessions.
type DiscoveryEngine struct {
	repo DiscoveryRepository
	now  func() time.Time
}

// NewDiscoveryEngine creates a new DiscoveryEngine.
func NewDiscoveryEngine(repo DiscoveryRepository) *DiscoveryEngine {
	return &DiscoveryEngine{repo: repo, now: time.Now().UTC}
}

// StartSession begins a new discovery session.
func (e *DiscoveryEngine) StartSession(projectID, rawContext string) (DiscoverySession, error) {
	now := e.now()
	session := DiscoverySession{
		ID:         domain.NewID("disc"),
		ProjectID:  projectID,
		Status:     DiscoveryInProgress,
		Summary:    "",
		RawContext: rawContext,
		Findings:   []Finding{},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	return e.repo.CreateSession(session)
}

// AddFinding adds a finding to a discovery session.
func (e *DiscoveryEngine) AddFinding(sessionID string, f Finding) error {
	session, err := e.repo.GetSession(sessionID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}
	session.Findings = append(session.Findings, f)
	if err := e.repo.UpdateSession(sessionID, session.Status, session.Summary); err != nil {
		return fmt.Errorf("update session: %w", err)
	}
	return nil
}

// IdentifyAssumption identifies and persists an assumption from a discovery session.
func (e *DiscoveryEngine) IdentifyAssumption(sessionID, projectID, description string, category AssumptionCategory, confidence float64) (Assumption, error) {
	if strings.TrimSpace(description) == "" {
		return Assumption{}, fmt.Errorf("assumption description is required")
	}
	if confidence < 0 || confidence > 1.0 {
		return Assumption{}, fmt.Errorf("confidence must be between 0 and 1")
	}
	a := Assumption{
		ID:          domain.NewID("asmp"),
		ProjectID:   projectID,
		SessionID:   sessionID,
		Description: description,
		Category:    category,
		Confidence:  confidence,
		Status:      AssumptionUnvalidated,
		CreatedAt:   e.now(),
	}
	return e.repo.CreateAssumption(a)
}

// IdentifyAmbiguity identifies and persists an ambiguity from a discovery session.
func (e *DiscoveryEngine) IdentifyAmbiguity(sessionID, projectID, description, category string) (Ambiguity, error) {
	if strings.TrimSpace(description) == "" {
		return Ambiguity{}, fmt.Errorf("ambiguity description is required")
	}
	a := Ambiguity{
		ID:          domain.NewID("ambg"),
		ProjectID:   projectID,
		SessionID:   sessionID,
		Description: description,
		Category:    category,
		Status:      AmbiguityOpen,
		CreatedAt:   e.now(),
	}
	return e.repo.CreateAmbiguity(a)
}

// ValidateAssumption marks an assumption as validated or invalidated.
func (e *DiscoveryEngine) ValidateAssumption(id string, validated bool, validatedBy string) error {
	status := AssumptionValidated
	if !validated {
		status = AssumptionInvalidated
	}
	return e.repo.UpdateAssumptionStatus(id, status, validatedBy)
}

// ResolveAmbiguity records a resolution for an ambiguity.
func (e *DiscoveryEngine) ResolveAmbiguity(id, resolution string) error {
	if strings.TrimSpace(resolution) == "" {
		return fmt.Errorf("resolution is required")
	}
	return e.repo.ResolveAmbiguity(id, resolution)
}

// CompleteSession finalizes a discovery session.
func (e *DiscoveryEngine) CompleteSession(id, summary string) error {
	if strings.TrimSpace(summary) == "" {
		return fmt.Errorf("summary is required to complete a session")
	}
	return e.repo.UpdateSession(id, DiscoveryComplete, summary)
}

// SubmitForApproval submits a vision for approval with discovery context.
func (e *DiscoveryEngine) SubmitForApproval(projectID, sessionID, visionID string) (VisionApproval, error) {
	approval := VisionApproval{
		ID:        domain.NewID("vappr"),
		ProjectID: projectID,
		SessionID: sessionID,
		VisionID:  visionID,
		Status:    ApprovalPending,
		CreatedAt: e.now(),
	}
	return e.repo.CreateApproval(approval)
}

// ApproveVision approves a vision.
func (e *DiscoveryEngine) ApproveVision(approvalID, approvedBy, feedback string) error {
	return e.repo.ApproveApproval(approvalID, approvedBy, feedback)
}

// RejectVision rejects a vision with feedback.
func (e *DiscoveryEngine) RejectVision(approvalID, feedback string) error {
	if strings.TrimSpace(feedback) == "" {
		return fmt.Errorf("rejection requires feedback")
	}
	return e.repo.RejectApproval(approvalID, feedback)
}

// GetDiscoverySummary returns a human-readable summary of a discovery session.
func (e *DiscoveryEngine) GetDiscoverySummary(sessionID string) (string, error) {
	session, err := e.repo.GetSession(sessionID)
	if err != nil {
		return "", fmt.Errorf("get session: %w", err)
	}
	assumptions, err := e.repo.ListAssumptions(sessionID)
	if err != nil {
		return "", fmt.Errorf("list assumptions: %w", err)
	}
	ambiguities, err := e.repo.ListAmbiguities(sessionID)
	if err != nil {
		return "", fmt.Errorf("list ambiguities: %w", err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("## Discovery Session: %s\n", session.ID[:8]))
	b.WriteString(fmt.Sprintf("Status: %s\n", session.Status))
	if session.Summary != "" {
		b.WriteString(fmt.Sprintf("Summary: %s\n", session.Summary))
	}

	if len(assumptions) > 0 {
		b.WriteString(fmt.Sprintf("\n### Assumptions (%d)\n", len(assumptions)))
		for _, a := range assumptions {
			b.WriteString(fmt.Sprintf("- [%s] %s (confidence: %.0f%%, status: %s)\n", a.Category, a.Description, a.Confidence*100, a.Status))
		}
	}

	if len(ambiguities) > 0 {
		b.WriteString(fmt.Sprintf("\n### Ambiguities (%d)\n", len(ambiguities)))
		for _, a := range ambiguities {
			b.WriteString(fmt.Sprintf("- [%s] %s (status: %s)\n", a.Category, a.Description, a.Status))
		}
	}

	return b.String(), nil
}

// HeuristicAnalyze performs a basic heuristic analysis of raw context to extract
// potential assumptions and ambiguities.
func (e *DiscoveryEngine) HeuristicAnalyze(sessionID, projectID, rawContext string) ([]Assumption, []Ambiguity, error) {
	var assumptions []Assumption
	var ambiguities []Ambiguity

	lines := strings.Split(rawContext, "\n")
	for _, line := range lines {
		lower := strings.ToLower(strings.TrimSpace(line))
		if lower == "" {
			continue
		}

		// Detect assumption-like statements
		if strings.Contains(lower, "assume") || strings.Contains(lower, "presumably") ||
			strings.Contains(lower, "probably") || strings.Contains(lower, "should work") {
			cat := AssumptionTechnical
			if strings.Contains(lower, "user") || strings.Contains(lower, "customer") {
				cat = AssumptionUser
			}
			if strings.Contains(lower, "cost") || strings.Contains(lower, "budget") {
				cat = AssumptionBusiness
			}
			a, err := e.IdentifyAssumption(sessionID, projectID, line, cat, 0.5)
			if err == nil {
				assumptions = append(assumptions, a)
			}
		}

		// Detect ambiguity-like statements
		if strings.Contains(lower, "unclear") || strings.Contains(lower, "not sure") ||
			strings.Contains(lower, "maybe") || strings.Contains(lower, "need to decide") ||
			strings.Contains(lower, "?)") || strings.Contains(lower, "tbd") {
			a, err := e.IdentifyAmbiguity(sessionID, projectID, line, "general")
			if err == nil {
				ambiguities = append(ambiguities, a)
			}
		}
	}

	return assumptions, ambiguities, nil
}
