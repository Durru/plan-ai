package modelstrategy

// ──────────────────────────────────────────────
// Prompt contracts define what each prompt type must include.
// ──────────────────────────────────────────────

type VisionContract struct {
	ProjectContext string `json:"project_context"`
	RawInputs      string `json:"raw_inputs"`
	TargetUsers    string `json:"target_users,omitempty"`
	Constraints    string `json:"constraints,omitempty"`
}

type ResearchContract struct {
	Topic        string `json:"topic"`
	ProjectScope string `json:"project_scope"`
	ExistingData string `json:"existing_data,omitempty"`
	Depth        string `json:"depth,omitempty"`
}

type PlanningContract struct {
	Vision       string `json:"vision"`
	Requirements string `json:"requirements"`
	Constraints  string `json:"constraints"`
	Research     string `json:"research"`
	Knowledge    string `json:"knowledge"`
	Decisions    string `json:"decisions"`
}

type ImpactContract struct {
	ChangeDescription string `json:"change_description"`
	AffectedEntities  string `json:"affected_entities"`
	CurrentState      string `json:"current_state"`
}

type ValidationContract struct {
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	Requirements string `json:"requirements"`
	Criterias    string `json:"criterias"`
}

// ──────────────────────────────────────────────
// PromptContract is a persisted prompt contract record.
// ──────────────────────────────────────────────

type PromptContract struct {
	ID           string `json:"id"`
	ContractType string `json:"contract_type"`
	Content      string `json:"content"` // JSON-serialized specific contract
	CreatedAt    string `json:"created_at"`
}
