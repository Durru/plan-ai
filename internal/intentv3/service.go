package intentv3

import (
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

// Service implements Phase 51 Product Intent Engine + Phase 52 Discovery.
type Service struct {
	intentRepo   ProductIntentRepository
	discoverRepo DiscoveryResultRepository
	detector     Detector
	now          func() time.Time
}

func NewService(intentRepo ProductIntentRepository, discoverRepo DiscoveryResultRepository) Service {
	return Service{
		intentRepo:   intentRepo,
		discoverRepo: discoverRepo,
		detector:     NewDetector(),
		now:          time.Now,
	}
}

// ──────────────────────────────────────────────
// Phase 51: Product Intent CRUD & lifecycle
// ──────────────────────────────────────────────

type CreateProductIntentInput struct {
	ProjectID         string
	Description       string
	ExpectedOutcome   string
	DesiredExperience string
	DesiredResult     string
	UserExpectations  []string
	NonExpectations   []string
	SuccessDefinition string
	FailureDefinition string
	DiscoveryResultID string // optional link to Phase 52 result
}

func (s Service) CreateProductIntent(input CreateProductIntentInput) (ProductIntent, error) {
	if input.ProjectID == "" {
		return ProductIntent{}, fmt.Errorf("project_id is required")
	}
	if input.Description == "" {
		return ProductIntent{}, fmt.Errorf("description is required")
	}
	now := s.now().UTC()
	pi := ProductIntent{
		ID:                domain.NewID("pintent"),
		ProjectID:         input.ProjectID,
		Description:       input.Description,
		ExpectedOutcome:   input.ExpectedOutcome,
		DesiredExperience: input.DesiredExperience,
		DesiredResult:     input.DesiredResult,
		UserExpectations:  input.UserExpectations,
		NonExpectations:   input.NonExpectations,
		SuccessDefinition: input.SuccessDefinition,
		FailureDefinition: input.FailureDefinition,
		Status:            StatusDraft,
		DiscoveryResultID: input.DiscoveryResultID,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	return s.intentRepo.SaveProductIntent(pi)
}

func (s Service) GetProductIntent(id string) (ProductIntent, error) {
	return s.intentRepo.GetProductIntent(id)
}

func (s Service) ListProductIntents(projectID string) ([]ProductIntent, error) {
	return s.intentRepo.ListProductIntents(projectID)
}

func (s Service) UpdateProductIntent(id string, input CreateProductIntentInput) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if existing.Status == StatusApproved || existing.Status == StatusArchived {
		return ProductIntent{}, fmt.Errorf("cannot update intent with status %s", existing.Status)
	}
	if input.Description != "" {
		existing.Description = input.Description
	}
	if input.ExpectedOutcome != "" {
		existing.ExpectedOutcome = input.ExpectedOutcome
	}
	if input.DesiredExperience != "" {
		existing.DesiredExperience = input.DesiredExperience
	}
	if input.DesiredResult != "" {
		existing.DesiredResult = input.DesiredResult
	}
	if len(input.UserExpectations) > 0 {
		existing.UserExpectations = input.UserExpectations
	}
	if len(input.NonExpectations) > 0 {
		existing.NonExpectations = input.NonExpectations
	}
	if input.SuccessDefinition != "" {
		existing.SuccessDefinition = input.SuccessDefinition
	}
	if input.FailureDefinition != "" {
		existing.FailureDefinition = input.FailureDefinition
	}
	existing.UpdatedAt = s.now().UTC()
	return s.intentRepo.UpdateProductIntent(existing)
}

func (s Service) SubmitProductIntentForApproval(id string) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if !IsValidTransition(existing.Status, StatusPendingApproval) {
		return ProductIntent{}, fmt.Errorf("cannot submit intent with status %s for approval", existing.Status)
	}
	return s.intentRepo.UpdateProductIntentStatus(id, StatusPendingApproval)
}

func (s Service) ApproveProductIntent(id string) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if !IsValidTransition(existing.Status, StatusApproved) {
		return ProductIntent{}, fmt.Errorf("cannot approve intent with status %s", existing.Status)
	}
	return s.intentRepo.UpdateProductIntentStatus(id, StatusApproved)
}

func (s Service) RejectProductIntent(id string) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if !IsValidTransition(existing.Status, StatusDraft) {
		return ProductIntent{}, fmt.Errorf("cannot reject intent with status %s", existing.Status)
	}
	return s.intentRepo.UpdateProductIntentStatus(id, StatusDraft)
}

func (s Service) ArchiveProductIntent(id string) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if !IsValidTransition(existing.Status, StatusArchived) {
		return ProductIntent{}, fmt.Errorf("cannot archive intent with status %s", existing.Status)
	}
	return s.intentRepo.UpdateProductIntentStatus(id, StatusArchived)
}

func (s Service) SupersedeProductIntent(id string) (ProductIntent, error) {
	existing, err := s.intentRepo.GetProductIntent(id)
	if err != nil {
		return ProductIntent{}, err
	}
	if !IsValidTransition(existing.Status, StatusSuperseded) {
		return ProductIntent{}, fmt.Errorf("cannot supersede intent with status %s", existing.Status)
	}
	return s.intentRepo.UpdateProductIntentStatus(id, StatusSuperseded)
}

// ──────────────────────────────────────────────
// Phase 52: Intent Discovery
// ──────────────────────────────────────────────

func (s Service) DiscoverIntent(projectID, content string) (DiscoveryResult, error) {
	if projectID == "" {
		return DiscoveryResult{}, fmt.Errorf("project_id is required")
	}
	if content == "" {
		return DiscoveryResult{}, fmt.Errorf("content is required")
	}
	result := s.detector.Discover(projectID, content)
	return s.discoverRepo.SaveDiscoveryResult(result)
}

func (s Service) GetDiscoveryResult(id string) (DiscoveryResult, error) {
	return s.discoverRepo.GetDiscoveryResult(id)
}

func (s Service) ListDiscoveryResults(projectID string) ([]DiscoveryResult, error) {
	return s.discoverRepo.ListDiscoveryResults(projectID)
}

// IsApprovedProductIntent checks if there is an approved V3 product intent for the given project.
// This is the guard primitive for "no planning without approved intent".
func (s Service) IsApprovedProductIntent(projectID string) bool {
	intents, err := s.intentRepo.ListProductIntents(projectID)
	if err != nil {
		return false
	}
	for _, pi := range intents {
		if pi.Status == StatusApproved {
			return true
		}
	}
	return false
}
