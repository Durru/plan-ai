package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
	"github.com/Durru/plan-ai/internal/intentv3"
	"github.com/Durru/plan-ai/internal/store"
)

// ── Phase 51: Product Intent Handlers (store repos directly, no intentv3.Service) ──

func HandleCreateProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	description := getStringArg(args, "description")
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	expectedOutcome := getStringArg(args, "expected_outcome")
	desiredExperience := getStringArg(args, "desired_experience")
	desiredResult := getStringArg(args, "desired_result")
	userExpectations := []string{}
	if ue := getStringArg(args, "user_expectations"); ue != "" {
		userExpectations = strings.Split(ue, "\n")
	}
	nonExpectations := []string{}
	if ne := getStringArg(args, "non_expectations"); ne != "" {
		nonExpectations = strings.Split(ne, "\n")
	}

	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	now := time.Now().UTC()
	pi := intentv3.ProductIntent{
		ID:                domain.NewID("pintent"),
		ProjectID:         projectID(projectRoot),
		Description:       description,
		ExpectedOutcome:   expectedOutcome,
		DesiredExperience: desiredExperience,
		DesiredResult:     desiredResult,
		UserExpectations:  userExpectations,
		NonExpectations:   nonExpectations,
		SuccessDefinition: getStringArg(args, "success_definition"),
		FailureDefinition: getStringArg(args, "failure_definition"),
		Status:            intentv3.StatusDraft,
		DiscoveryResultID: getStringArg(args, "discovery_result_id"),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	pi, err = intentRepo.SaveProductIntent(pi)
	if err != nil {
		return nil, fmt.Errorf("create product intent: %w", err)
	}

	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleListProductIntents(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	list, err := intentRepo.ListProductIntents(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list product intents: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, pi := range list {
		items = append(items, map[string]any{
			"id":          pi.ID,
			"description": pi.Description,
			"status":      string(pi.Status),
			"created_at":  formatTime(pi.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}

func HandleGetProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	pi, err := intentRepo.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleSubmitProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	existing, err := intentRepo.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	if !intentv3.IsValidTransition(existing.Status, intentv3.StatusPendingApproval) {
		return nil, fmt.Errorf("cannot submit intent with status %s for approval", existing.Status)
	}
	pi, err := intentRepo.UpdateProductIntentStatus(intentID, intentv3.StatusPendingApproval)
	if err != nil {
		return nil, fmt.Errorf("submit product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleApproveProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	existing, err := intentRepo.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	if !intentv3.IsValidTransition(existing.Status, intentv3.StatusApproved) {
		return nil, fmt.Errorf("cannot approve intent with status %s", existing.Status)
	}
	pi, err := intentRepo.UpdateProductIntentStatus(intentID, intentv3.StatusApproved)
	if err != nil {
		return nil, fmt.Errorf("approve product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleRejectProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	intentRepo := store.NewIntentV3Repository(ps.DB)

	existing, err := intentRepo.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	if !intentv3.IsValidTransition(existing.Status, intentv3.StatusDraft) {
		return nil, fmt.Errorf("cannot reject intent with status %s", existing.Status)
	}
	pi, err := intentRepo.UpdateProductIntentStatus(intentID, intentv3.StatusDraft)
	if err != nil {
		return nil, fmt.Errorf("reject product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

// ── Phase 52: Discovery Engine Handlers (store repos directly, no intentv3.Service) ──

func HandleDiscoverIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	content := getStringArg(args, "content")
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	detector := intentv3.NewDetector()
	result := detector.Discover(projectID(projectRoot), content)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	dr, err := discRepo.SaveDiscoveryResult(result)
	if err != nil {
		return nil, fmt.Errorf("discover intent: %w", err)
	}
	return map[string]any{
		"id":              dr.ID,
		"project_id":      dr.ProjectID,
		"raw_input":       dr.RawInput,
		"detected_intent": dr.DetectedIntent,
		"classification":  dr.Classification,
		"objectives":      dr.Objectives,
		"restrictions":    dr.Restrictions,
		"preferences":     dr.Preferences,
		"expectations":    dr.Expectations,
		"gaps":            dr.Gaps,
		"questions":       dr.Questions,
		"created_at":      formatTime(dr.CreatedAt),
	}, nil
}

func HandleListDiscoveryResults(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, cleanup, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)

	list, err := discRepo.ListDiscoveryResults(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list discovery results: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, dr := range list {
		items = append(items, map[string]any{
			"id":              dr.ID,
			"detected_intent": dr.DetectedIntent,
			"classification":  dr.Classification,
			"created_at":      formatTime(dr.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}
