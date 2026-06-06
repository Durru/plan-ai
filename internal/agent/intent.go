package agent

import "strings"

// DefaultIntentDetector identifies user intent from input text.
type DefaultIntentDetector struct{}

// NewIntentDetector creates a new DefaultIntentDetector.
func NewIntentDetector() *DefaultIntentDetector {
	return &DefaultIntentDetector{}
}

// DetectIntent analyzes input and returns the matched intent.
func (d *DefaultIntentDetector) DetectIntent(input string) IntentKind {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return IntentUnknown
	}

	// Check for approval first (short inputs)
	if input == "yes" || input == "approve" || input == "approved" || input == "ok" || input == "y" || strings.HasPrefix(input, "approve ") {
		return IntentApprove
	}
	if input == "no" || input == "reject" || input == "rejected" || strings.HasPrefix(input, "reject ") {
		return IntentReject
	}

	// Multi-word intent detection
	words := strings.Fields(input)
	full := strings.Join(words, " ")

	// Check for database plan (before broad "plan"/"create" checks)
	if containsAny(words, "database", "db", "postgres", "mysql", "sqlite") && containsAny(words, "design", "schema", "plan", "model") {
		return IntentDatabasePlan
	}

	// Check for create master plan
	if containsAny(words, "master", "plan") && containsAny(words, "create", "new", "make") {
		return IntentCreateMasterPlan
	}

	// Check for create specific plan
	if containsAny(words, "specific", "detail", "implementation") && containsAny(words, "plan", "implement") {
		return IntentCreateSpecificPlan
	}

	// Check for product creation (before broad "build"/"create" checks)
	if containsAny(words, "create", "build", "make", "start") && containsAny(words, "saas", "app", "application", "product", "service", "platform") {
		return IntentCreateProduct
	}

	// Check for research
	if containsAny(words, "research", "investigate", "find", "search", "lookup") {
		return IntentResearchTopic
	}

	// Impact analysis (BEFORE ChangeRequest so "what if I change X" routes correctly)
	if containsAny(words, "impact", "impacts", "consequences", "consequence", "risk", "risks", "affect", "affects", "break") || strings.Contains(full, "what if") || strings.Contains(full, "what would happen") || strings.Contains(full, "what are the consequences") {
		return IntentImpactAnalysis
	}

	// Check for update plan
	if containsAny(words, "update", "modify", "change", "revise") && containsAny(words, "plan", "phase", "task") {
		return IntentUpdatePlan
	}

	// Check for change request
	if containsAny(words, "change", "add", "remove", "different", "instead") {
		return IntentChangeRequest
	}

	// Check for implementation help
	if containsAny(words, "implement", "code", "build", "write", "develop") {
		return IntentImplementationHelp
	}

	// Conversation intents — check BEFORE broad status/next-task checks
	// so phrases like "where are we" route to AnalyzeProject, not Status.

	// Analyze project
	if containsAny(words, "analyze", "analysis", "summary", "summarize", "overview", "review", "audit") && containsAny(words, "project", "state", "status", "codebase") {
		return IntentAnalyzeProject
	}
	if strings.Contains(full, "what do i have") || strings.Contains(full, "show me what") || strings.Contains(full, "where are we") {
		return IntentAnalyzeProject
	}

	// Impact analysis
	if containsAny(words, "impact", "consequence", "risk", "affect", "break") || strings.Contains(full, "what if") || strings.Contains(full, "what would happen") {
		return IntentImpactAnalysis
	}

	// Check for status
	if containsAny(words, "status", "progress", "where", "what's next", "what is next") {
		return IntentProjectStatus
	}

	// Check for next task
	if containsAny(words, "next", "todo", "pending") {
		return IntentNextTask
	}

	// Check for validate
	if containsAny(words, "validate", "verify", "check", "review") {
		return IntentValidate
	}

	return IntentUnknown
}

func containsAny(words []string, targets ...string) bool {
	for _, w := range words {
		for _, t := range targets {
			if w == t {
				return true
			}
		}
	}
	// Also check the full input against any multi-word targets
	full := strings.Join(words, " ")
	for _, t := range targets {
		if strings.Contains(t, " ") && strings.Contains(full, t) {
			return true
		}
	}
	return false
}
