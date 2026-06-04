package modelstrategy

// ──────────────────────────────────────────────
// Output schemas define what each response must contain.
// ──────────────────────────────────────────────

type VisionSchema struct {
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	TargetUsers     []string `json:"target_users"`
	FunctionalGoals []string `json:"functional_goals"`
	UXGoals         []string `json:"ux_goals"`
	BusinessGoals   []string `json:"business_goals"`
	Constraints     []string `json:"constraints"`
	Assumptions     []string `json:"assumptions"`
	MissingInfo     []string `json:"missing_information"`
	SuccessCriteria []string `json:"success_criteria"`
}

type ResearchSchema struct {
	Topic       string        `json:"topic"`
	Summary     string        `json:"summary"`
	Findings    []FindingItem `json:"findings"`
	Sources     []SourceItem  `json:"sources"`
	Conclusions []string      `json:"conclusions"`
	Confidence  float64       `json:"confidence"`
}

type FindingItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SourceItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Type  string `json:"type"`
}

type KnowledgeSchema struct {
	Topic      string   `json:"topic"`
	Category   string   `json:"category"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	Confidence float64  `json:"confidence"`
	Tags       []string `json:"tags"`
}

type PlanSchema struct {
	Title      string      `json:"title"`
	Summary    string      `json:"summary"`
	Phases     []PhaseItem `json:"phases"`
	Risks      []string    `json:"risks"`
	Validation []string    `json:"validation_criteria"`
}

type PhaseItem struct {
	Title    string `json:"title"`
	Summary  string `json:"summary"`
	Position int    `json:"position"`
}

type ImpactSchema struct {
	Summary        string   `json:"summary"`
	AffectedPlans  []string `json:"affected_plans"`
	AffectedPhases []string `json:"affected_phases"`
	AffectedTasks  []string `json:"affected_tasks"`
	Risks          []string `json:"risks"`
}

type ValidationSchema struct {
	Valid  bool     `json:"valid"`
	Reason string   `json:"reason"`
	Issues []string `json:"issues"`
}

// ──────────────────────────────────────────────
// OutputSchema is a persisted output schema record.
// ──────────────────────────────────────────────

type OutputSchema struct {
	ID         string `json:"id"`
	SchemaType string `json:"schema_type"`
	Fields     string `json:"fields"`   // JSON describing expected fields
	Required   string `json:"required"` // JSON list of required field names
	CreatedAt  string `json:"created_at"`
}
