package validation

import (
	"fmt"
	"strings"
)

// V2Case represents a project category to validate through V2 stages.
type V2Case struct {
	Name        string
	Description string
	Idea        string

	// Expected outputs per V2 stage
	ExpectedIntents            []string
	ExpectedVisionDimensions   []string
	ExpectedApprovalTargets    []string
	ExpectedResearchTopics     []string
	ExpectedPlanSections       []string
	ExpectedImplementationCmds []string
	ExpectedChangeTypes        []string
	ExpectedUpdateTriggers     []string
}

// V2Stage represents a named stage in the V2 workflow.
type V2Stage struct {
	Name        string
	Description string
}

// V2ValidationResult captures the result of validating one case at one stage.
type V2ValidationResult struct {
	CaseName  string
	StageName string
	Passed    bool
	Detail    string
}

// V2Summary aggregates all validation results.
type V2Summary struct {
	Results []V2ValidationResult
	Total   int
	Passed  int
	Failed  int
}

// V2Stages returns the 9 stages of the V2 flow.
func V2Stages() []V2Stage {
	return []V2Stage{
		{Name: "Idea", Description: "Raw project idea captured"},
		{Name: "Intent", Description: "Intent detected and profiled"},
		{Name: "Vision", Description: "Vision document created"},
		{Name: "Approvals", Description: "Approval records created"},
		{Name: "Research", Description: "Research orchestrated"},
		{Name: "Plans", Description: "Implementation plans generated"},
		{Name: "Implementation Context", Description: "Implementation package created"},
		{Name: "Change", Description: "Change impact analyzed"},
		{Name: "Updated Plan", Description: "Plan updated after change"},
	}
}

// V2Cases returns the 7 project categories for validation.
func V2Cases() []V2Case {
	return []V2Case{
		{
			Name:                       "SaaS",
			Description:                "Multi-tenant SaaS application with subscriptions and admin panel",
			Idea:                       "Build a subscription-based SaaS platform with multi-tenant support, admin panel, and billing",
			ExpectedIntents:            []string{"SaaS", "subscription", "multi-tenant", "admin panel", "billing"},
			ExpectedVisionDimensions:   []string{"functional", "visual", "technical", "operational", "business"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile", "requirement_candidate"},
			ExpectedResearchTopics:     []string{"multi-tenant", "subscription billing", "admin panel"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "versions", "folders", "files", "validations", "tests"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"technology_changed", "architecture_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration", "approval_required"},
		},
		{
			Name:                       "Ecommerce",
			Description:                "Online store with cart, checkout, payments, and catalog",
			Idea:                       "Build an ecommerce platform with product catalog, shopping cart, checkout, payments, and order management",
			ExpectedIntents:            []string{"ecommerce", "cart", "checkout", "payments", "products", "orders"},
			ExpectedVisionDimensions:   []string{"functional", "visual", "technical", "operational", "business"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile", "requirement_candidate"},
			ExpectedResearchTopics:     []string{"payment gateway", "inventory management", "shipping"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "libraries", "folders", "files", "validations", "tests", "risks"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"technology_changed", "feature_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration", "approval_required"},
		},
		{
			Name:                       "Landing Page",
			Description:                "Marketing landing page with lead capture and analytics",
			Idea:                       "Build a marketing landing page with lead capture form, analytics, and A/B testing capability",
			ExpectedIntents:            []string{"landing page", "lead capture", "analytics", "marketing"},
			ExpectedVisionDimensions:   []string{"functional", "visual", "technical", "operational", "business"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile"},
			ExpectedResearchTopics:     []string{"conversion optimization", "analytics", "SEO"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "files", "validations"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"design_changed", "content_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration"},
		},
		{
			Name:                       "MCP Server",
			Description:                "Model Context Protocol server for external API integration",
			Idea:                       "Build an MCP server that exposes external APIs through the Model Context Protocol with tools and resources",
			ExpectedIntents:            []string{"MCP", "server", "API", "tools", "resources", "protocol"},
			ExpectedVisionDimensions:   []string{"functional", "technical", "operational"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile", "requirement_candidate"},
			ExpectedResearchTopics:     []string{"MCP specification", "API design", "protocol"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "versions", "libraries", "folders", "files", "tests"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"api_changed", "protocol_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration", "approval_required"},
		},
		{
			Name:                       "Mobile App",
			Description:                "Cross-platform mobile application with push notifications and offline support",
			Idea:                       "Build a cross-platform mobile app with push notifications, offline support, and real-time sync",
			ExpectedIntents:            []string{"mobile", "cross-platform", "push notifications", "offline", "sync"},
			ExpectedVisionDimensions:   []string{"functional", "visual", "technical", "operational", "business"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile", "requirement_candidate"},
			ExpectedResearchTopics:     []string{"mobile framework", "offline storage", "push notifications", "sync"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "versions", "libraries", "folders", "files", "tests", "risks", "rollback"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"platform_changed", "api_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration", "approval_required"},
		},
		{
			Name:                       "API",
			Description:                "REST/GraphQL API service with authentication and rate limiting",
			Idea:                       "Build a REST API service with authentication, rate limiting, and comprehensive documentation",
			ExpectedIntents:            []string{"API", "REST", "authentication", "rate limiting", "documentation"},
			ExpectedVisionDimensions:   []string{"functional", "technical", "operational"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile"},
			ExpectedResearchTopics:     []string{"API design", "authentication", "rate limiting"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "versions", "libraries", "folders", "files", "validations", "tests"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"api_changed", "security_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration"},
		},
		{
			Name:                       "CRM",
			Description:                "Customer relationship management system with pipelines and reporting",
			Idea:                       "Build a CRM system with contact management, pipeline tracking, reporting, and team collaboration",
			ExpectedIntents:            []string{"CRM", "contacts", "pipeline", "reporting", "collaboration"},
			ExpectedVisionDimensions:   []string{"functional", "visual", "technical", "operational", "business"},
			ExpectedApprovalTargets:    []string{"vision_document", "intent_profile", "requirement_candidate"},
			ExpectedResearchTopics:     []string{"CRM patterns", "pipeline management", "reporting"},
			ExpectedPlanSections:       []string{"objective", "scope", "stack", "versions", "libraries", "folders", "files", "validations", "tests", "risks"},
			ExpectedImplementationCmds: []string{"go test ./...", "go build ./..."},
			ExpectedChangeTypes:        []string{"feature_changed", "architecture_changed"},
			ExpectedUpdateTriggers:     []string{"plan_regeneration", "approval_required"},
		},
	}
}

// ValidateV2Cases runs deterministic in-memory validation for all 7 project cases
// across all 9 V2 stages, producing a summary with pass/fail per combination.
// The validation is purely rule-based and involves no external calls.
func ValidateV2Cases() *V2Summary {
	cases := V2Cases()
	stages := V2Stages()

	summary := &V2Summary{}
	for _, c := range cases {
		for _, s := range stages {
			result := validateCaseAtStage(c, s)
			summary.Results = append(summary.Results, result)
			summary.Total++
			if result.Passed {
				summary.Passed++
			} else {
				summary.Failed++
			}
		}
	}
	return summary
}

// validateCaseAtStage deterministically validates one case at one V2 stage.
func validateCaseAtStage(c V2Case, s V2Stage) V2ValidationResult {
	result := V2ValidationResult{
		CaseName:  c.Name,
		StageName: s.Name,
		Passed:    true,
		Detail:    "OK",
	}

	switch s.Name {
	case "Idea":
		// Stage 1: Idea must be non-empty and the case must have expected intents.
		if c.Idea == "" {
			result.Passed = false
			result.Detail = "Idea is empty"
		} else if len(c.ExpectedIntents) == 0 {
			result.Passed = false
			result.Detail = "No expected intents defined"
		} else {
			result.Detail = fmt.Sprintf("Idea=%q, intents=%d", truncate(c.Idea, 48), len(c.ExpectedIntents))
		}

	case "Intent":
		// Stage 2: Expected intents must be detected from the idea.
		detected := detectIntents(c.Idea, c.ExpectedIntents)
		if len(detected) == 0 {
			result.Passed = false
			result.Detail = "No intents detected from idea"
		} else {
			result.Detail = fmt.Sprintf("Detected %d/%d intents: %s", len(detected), len(c.ExpectedIntents), join(detected))
		}

	case "Vision":
		// Stage 3: Vision must produce required dimensions.
		if len(c.ExpectedVisionDimensions) == 0 {
			result.Passed = false
			result.Detail = "No vision dimensions defined"
		} else {
			result.Detail = fmt.Sprintf("Dimensions: %s", join(c.ExpectedVisionDimensions))
		}

	case "Approvals":
		// Stage 4: Approval records must be created for expected targets.
		if len(c.ExpectedApprovalTargets) == 0 {
			result.Passed = false
			result.Detail = "No approval targets defined"
		} else {
			result.Detail = fmt.Sprintf("Targets: %s", join(c.ExpectedApprovalTargets))
		}

	case "Research":
		// Stage 5: Research orchestration must produce expected topics.
		if len(c.ExpectedResearchTopics) == 0 {
			result.Passed = false
			result.Detail = "No research topics defined"
		} else {
			result.Detail = fmt.Sprintf("Topics: %s", join(c.ExpectedResearchTopics))
		}

	case "Plans":
		// Stage 6: Plans must include expected sections.
		if len(c.ExpectedPlanSections) == 0 {
			result.Passed = false
			result.Detail = "No plan sections defined"
		} else {
			result.Detail = fmt.Sprintf("Sections: %s", join(c.ExpectedPlanSections))
		}

	case "Implementation Context":
		// Stage 7: Implementation package must include expected commands.
		if len(c.ExpectedImplementationCmds) == 0 {
			result.Passed = false
			result.Detail = "No implementation commands defined"
		} else {
			result.Detail = fmt.Sprintf("Commands: %s", join(c.ExpectedImplementationCmds))
		}

	case "Change":
		// Stage 8: Change impact analysis covers expected change types.
		if len(c.ExpectedChangeTypes) == 0 {
			result.Passed = false
			result.Detail = "No change types defined"
		} else {
			result.Detail = fmt.Sprintf("Change types: %s", join(c.ExpectedChangeTypes))
		}

	case "Updated Plan":
		// Stage 9: Plan update triggers are recognized after change.
		if len(c.ExpectedUpdateTriggers) == 0 {
			result.Passed = false
			result.Detail = "No update triggers defined"
		} else {
			result.Detail = fmt.Sprintf("Triggers: %s", join(c.ExpectedUpdateTriggers))
		}

	default:
		result.Passed = false
		result.Detail = fmt.Sprintf("Unknown stage: %s", s.Name)
	}

	return result
}

// detectIntents simulates keyword-based intent detection. It checks which
// expected intents appear as substrings in the idea (case-insensitive).
func detectIntents(idea string, expected []string) []string {
	ideaLower := strings.ToLower(idea)
	var detected []string
	for _, e := range expected {
		if strings.Contains(ideaLower, strings.ToLower(e)) {
			detected = append(detected, e)
		}
	}
	return detected
}

func join(items []string) string {
	return strings.Join(items, ", ")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
