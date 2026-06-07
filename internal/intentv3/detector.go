package intentv3

import (
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// Detector performs deterministic intent discovery (Phase 52).
// It uses pattern/keyword matching — no network calls.
type Detector struct{}

func NewDetector() Detector { return Detector{} }

// Discover extracts structured intent from raw text.
func (d Detector) Discover(projectID, content string) DiscoveryResult {
	content = strings.TrimSpace(content)
	classification := d.classifyContent(content)

	return DiscoveryResult{
		ID:             domain.NewID("discv3"),
		ProjectID:      projectID,
		RawInput:       content,
		DetectedIntent: classification,
		Objectives:     d.detectObjectives(content),
		Restrictions:   d.detectRestrictions(content),
		Preferences:    d.detectPreferences(content),
		References:     d.detectReferences(content),
		Expectations:   d.detectExpectations(content),
		Classification: classification,
		Gaps:           d.detectGaps(content),
		Questions:      d.detectQuestions(content),
		CreatedAt:      time.Now().UTC(),
	}
}

// classifyContent assigns a high-level classification.
func (d Detector) classifyContent(content string) string {
	lower := strings.ToLower(content)

	if matchesAny(lower, "saas", "multi-tenant", "subscription", "cloud service") {
		return "saas_platform"
	}
	if matchesAny(lower, "mobile", "ios", "android", "app store") {
		return "mobile_application"
	}
	if matchesAny(lower, "crm", "erp", "billing", "invoicing", "payment") {
		return "business_tool"
	}
	if matchesAny(lower, "api", "sdk", "library", "package", "cli") {
		return "developer_tool"
	}
	if matchesAny(lower, "dashboard", "analytics", "report", "metrics", "monitoring") {
		return "analytics_tool"
	}
	if matchesAny(lower, "ai", "machine learning", "ml", "llm", "model", "chatbot") {
		return "ai_feature"
	}
	if matchesAny(lower, "website", "landing page", "ecommerce", "blog", "portfolio") {
		return "web_platform"
	}
	if matchesAny(lower, "automation", "workflow", "pipeline", "integration") {
		return "automation_tool"
	}
	return "general"
}

func (d Detector) detectObjectives(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if matchesAny(lower, "automate", "automation") {
		out = append(out, "Automate repetitive tasks")
	}
	if matchesAny(lower, "scale", "scalable") {
		out = append(out, "Scale the solution")
	}
	if matchesAny(lower, "integrate", "integration") {
		out = append(out, "Integrate with existing systems")
	}
	if matchesAny(lower, "simplify", "easy to use", "ux") {
		out = append(out, "Simplify user experience")
	}
	if matchesAny(lower, "secure", "security", "auth", "permission") {
		out = append(out, "Ensure security and access control")
	}
	if matchesAny(lower, "realtime", "real-time", "live") {
		out = append(out, "Support real-time updates")
	}
	if matchesAny(lower, "mobile", "responsive") {
		out = append(out, "Support mobile access")
	}
	if matchesAny(lower, "offline", "local") {
		out = append(out, "Support offline access")
	}
	if len(out) == 0 {
		out = append(out, "Deliver the described product or feature")
	}
	return out
}

func (d Detector) detectRestrictions(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if matchesAny(lower, "not a", "is not", "should not", "must not", "avoid", "without") {
		out = append(out, extractPhrases(lower, []string{"not a", "is not", "should not", "must not", "avoid", "without"})...)
	}
	if matchesAny(lower, "budget", "cost", "cheap", "affordable") {
		out = append(out, "Budget or cost constraints apply")
	}
	if matchesAny(lower, "deadline", "urgent", "asap", "quick") {
		out = append(out, "Time constraints apply")
	}
	if matchesAny(lower, "legacy", "existing system", "migrate") {
		out = append(out, "Must work with existing legacy systems")
	}
	if len(out) == 0 {
		out = append(out, "No explicit restrictions detected")
	}
	return out
}

func (d Detector) detectPreferences(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if matchesAny(lower, "prefer", "like", "want") {
		out = append(out, extractPhrases(lower, []string{"prefer", "like", "want"})...)
	}
	if matchesAny(lower, "typescript", "go", "rust", "python", "java") {
		out = append(out, "Specific language preference")
	}
	if matchesAny(lower, "react", "vue", "angular", "svelte") {
		out = append(out, "Specific framework preference")
	}
	if matchesAny(lower, "postgres", "mysql", "sqlite", "mongodb") {
		out = append(out, "Specific database preference")
	}
	if matchesAny(lower, "docker", "kubernetes", "serverless") {
		out = append(out, "Specific deployment preference")
	}
	if len(out) == 0 {
		out = append(out, "No explicit preferences detected")
	}
	return out
}

func (d Detector) detectReferences(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if matchesAny(lower, "like ", "similar to", "inspired by", "clone of") {
		out = append(out, "External reference or inspiration mentioned")
	}
	if matchesAny(lower, "https://", "http://", "github.com") {
		out = append(out, "External URL or repository mentioned")
	}
	if len(out) == 0 {
		out = append(out, "No explicit references detected")
	}
	return out
}

func (d Detector) detectExpectations(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if matchesAny(lower, "should ", "must ", "need ", "require") {
		out = append(out, extractPhrases(lower, []string{"should ", "must ", "need ", "require"})...)
	}
	if matchesAny(lower, "fast", "performant", "slow", "speed") {
		out = append(out, "Performance expectations")
	}
	if matchesAny(lower, "reliable", "uptime", "available") {
		out = append(out, "Reliability expectations")
	}
	if matchesAny(lower, "test", "testing", "tdd") {
		out = append(out, "Testing expectations")
	}
	if len(out) == 0 {
		out = append(out, "No explicit expectations detected")
	}
	return out
}

func (d Detector) detectGaps(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if !strings.Contains(lower, "test") &&
		!strings.Contains(lower, "testing") &&
		!strings.Contains(lower, "qa") {
		out = append(out, "No testing or QA strategy described")
	}
	if !strings.Contains(lower, "deploy") &&
		!strings.Contains(lower, "deployment") {
		out = append(out, "No deployment strategy described")
	}
	if !strings.Contains(lower, "security") &&
		!strings.Contains(lower, "auth") &&
		!strings.Contains(lower, "permission") {
		out = append(out, "No security or access control described")
	}
	if !strings.Contains(lower, "monitor") &&
		!strings.Contains(lower, "logging") &&
		!strings.Contains(lower, "observability") {
		out = append(out, "No monitoring or observability described")
	}
	return out
}

func (d Detector) detectQuestions(content string) []string {
	lower := strings.ToLower(content)
	var out []string

	if strings.Contains(lower, "?") {
		// extract sentences with question marks
		for _, part := range strings.Split(lower, "?") {
			part = strings.TrimSpace(part)
			if part != "" && strings.Contains(part, " ") {
				out = append(out, part+"?")
			}
		}
	}
	if len(out) > 5 {
		out = out[:5]
	}
	if len(out) == 0 {
		out = append(out, "Clarify the specific requirements further")
	}
	return out
}

// ──────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────

func matchesAny(lower string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func extractPhrases(lower string, triggers []string) []string {
	var out []string
	for _, trigger := range triggers {
		idx := strings.Index(lower, trigger)
		if idx < 0 {
			continue
		}
		// grab up to 80 chars after the trigger
		start := idx + len(trigger)
		end := start + 80
		if end > len(lower) {
			end = len(lower)
		}
		phrase := strings.TrimSpace(lower[start:end])
		// trim at first period, newline, or comma
		if p := strings.IndexAny(phrase, ".\n,"); p > 0 {
			phrase = phrase[:p]
		}
		if phrase != "" {
			out = append(out, phrase)
		}
	}
	return out
}
