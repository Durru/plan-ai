package intent

import "strings"

type Detector struct{}

func NewDetector() Detector { return Detector{} }

func (d Detector) Detect(projectID, content string) Profile {
	normalized := strings.ToLower(content)
	profile := Profile{
		ProjectID: projectID,
		Source:    content,
		Status:    StatusDraft,
	}

	if containsAny(normalized, "crm", "clientes", "customer") {
		profile.PrimaryIntent = Intent{Name: "CRM", Confidence: 90, State: SignalCandidate}
		profile.SecondaryGoals = appendUniqueGoals(profile.SecondaryGoals,
			"multi-user",
			"admin panel",
			"subscriptions",
			"reports",
			"automations",
		)
	}
	if containsAny(normalized, "saas", "software as a service") {
		if profile.PrimaryIntent.Name == "" {
			profile.PrimaryIntent = Intent{Name: "SaaS", Confidence: 85, State: SignalCandidate}
		} else {
			profile.SecondaryGoals = appendUniqueGoals(profile.SecondaryGoals, "SaaS")
		}
		profile.Expectations = appendUniqueExpectations(profile.Expectations, "subscription-ready product", "tenant-aware access")
	}
	if containsAny(normalized, "ecommerce", "e-commerce", "tienda", "shop") {
		if profile.PrimaryIntent.Name == "" {
			profile.PrimaryIntent = Intent{Name: "Ecommerce", Confidence: 90, State: SignalCandidate}
		}
		profile.SecondaryGoals = appendUniqueGoals(profile.SecondaryGoals, "cart", "checkout", "payments", "catalog")
	}
	if containsAny(normalized, "mcp") {
		if profile.PrimaryIntent.Name == "" {
			profile.PrimaryIntent = Intent{Name: "MCP Server", Confidence: 85, State: SignalCandidate}
		}
	}
	if profile.PrimaryIntent.Name == "" {
		profile.PrimaryIntent = Intent{Name: "Application", Confidence: 50, State: SignalCandidate}
	}

	profile.Constraints = detectConstraints(normalized)
	profile.SuccessCriteria = append(profile.SuccessCriteria, SuccessCriteria{Name: "user approves the vision before planning", State: SignalCandidate})
	profile.Priorities = []UserPriority{{Name: "clarify user intent", Rank: 1, State: SignalCandidate}}
	return profile
}

func containsAny(value string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func detectConstraints(value string) []string {
	var constraints []string
	for _, token := range []string{"sin asumir", "no asumir", "sandbox", "no romper", "compatibilidad", "presupuesto", "budget"} {
		if strings.Contains(value, token) {
			constraints = append(constraints, token)
		}
	}
	return constraints
}

func appendUniqueGoals(items []Goal, names ...string) []Goal {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Name] = true
	}
	for _, name := range names {
		if !seen[name] {
			items = append(items, Goal{Name: name, State: SignalCandidate})
			seen[name] = true
		}
	}
	return items
}

func appendUniqueExpectations(items []UserExpectation, names ...string) []UserExpectation {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Name] = true
	}
	for _, name := range names {
		if !seen[name] {
			items = append(items, UserExpectation{Name: name, State: SignalCandidate})
			seen[name] = true
		}
	}
	return items
}
