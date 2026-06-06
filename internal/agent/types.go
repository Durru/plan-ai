package agent

import "time"

// ──────────────────────────────────────────────
// Intent Types
// ──────────────────────────────────────────────

// IntentKind represents the detected user intent.
type IntentKind string

const (
	IntentCreateMasterPlan   IntentKind = "create_master_plan"
	IntentCreateSpecificPlan IntentKind = "create_specific_plan"
	IntentResearchTopic      IntentKind = "research_topic"
	IntentUpdatePlan         IntentKind = "update_plan"
	IntentChangeRequest      IntentKind = "change_request"
	IntentImplementationHelp IntentKind = "implementation_help"
	IntentProjectStatus      IntentKind = "project_status"
	IntentApprove            IntentKind = "approve"
	IntentReject             IntentKind = "reject"
	IntentValidate           IntentKind = "validate"
	IntentNextTask           IntentKind = "next_task"
	IntentAnalyzeProject     IntentKind = "analyze_project"
	IntentCreateProduct      IntentKind = "create_product"
	IntentDatabasePlan       IntentKind = "database_plan"
	IntentImpactAnalysis     IntentKind = "impact_analysis"
	IntentUnknown            IntentKind = "unknown"
)

// ──────────────────────────────────────────────
// Delegated Job Types
// ──────────────────────────────────────────────

// DelegatedJobType represents the type of a delegated job.
type DelegatedJobType string

const (
	JobTypeVision     DelegatedJobType = "vision_job"
	JobTypeResearch   DelegatedJobType = "research_job"
	JobTypePlanning   DelegatedJobType = "planning_job"
	JobTypeValidation DelegatedJobType = "validation_job"
	JobTypeImpact     DelegatedJobType = "impact_job"
	JobTypeContext    DelegatedJobType = "context_job"
)

// JobStatus represents the lifecycle status of a delegated job.
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// DelegatedJob represents a temporary delegated job created by the agent.
type DelegatedJob struct {
	ID            string           `json:"id"`
	ProjectID     string           `json:"project_id"`
	Intent        IntentKind       `json:"intent"`
	Capability    string           `json:"capability"`
	WorkflowType  string           `json:"workflow_type"`
	JobType       DelegatedJobType `json:"job_type"`
	Status        JobStatus        `json:"status"`
	ResultSummary string           `json:"result_summary"`
	CreatedAt     string           `json:"created_at"`
	CompletedAt   string           `json:"completed_at"`
}

// ──────────────────────────────────────────────
// Agent Response Types
// ──────────────────────────────────────────────

// AgentResponse is the structured response from the agent.
type AgentResponse struct {
	Message             string            `json:"message"`
	Status              string            `json:"status"`
	RequiresApproval    bool              `json:"requires_approval"`
	SuggestedNextAction string            `json:"suggested_next_action"`
	ContextUsed         []string          `json:"context_used"`
	WorkflowTriggered   string            `json:"workflow_triggered"`
	CreatedEntities     map[string]string `json:"created_entities"`
}

// ──────────────────────────────────────────────
// Agent Run & Message Types
// ──────────────────────────────────────────────

// AgentRunRecord represents a persisted agent execution.
type AgentRunRecord struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Intent    string `json:"intent"`
	Status    string `json:"status"`
	Response  string `json:"response"` // JSON
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AgentMessage represents a single message in an agent conversation.
type AgentMessage struct {
	ID        string `json:"id"`
	RunID     string `json:"run_id"`
	Role      string `json:"role"` // user, agent
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ──────────────────────────────────────────────
// Workflow Type Constants
// ──────────────────────────────────────────────

const (
	WorkflowVision   = "vision"
	WorkflowResearch = "research"
	WorkflowPlanning = "planning"
	WorkflowApproval = "approval"
	WorkflowImpact   = "impact"
	WorkflowContext  = "context"
	WorkflowStatus   = "status"

	CapabilityPlanning = "planning"
	CapabilityResearch = "research"
	CapabilityVision   = "vision"
	CapabilityChange   = "change"
	CapabilityContext  = "context"

	ModelStrategyDefault = "default"
)

// ──────────────────────────────────────────────
// Router Decision
// ──────────────────────────────────────────────

// RouterDecision is the output of the agent router.
type RouterDecision struct {
	Workflow         string   `json:"workflow"`
	Capability       string   `json:"capability"`
	ModelStrategy    string   `json:"model_strategy"`
	ContextKeys      []string `json:"context_keys"`
	RequiresApproval bool     `json:"requires_approval"`
}

// ContextPayload is loaded context for the response.
type ContextPayload struct {
	ProjectID   string
	Plans       []map[string]any
	Phases      []map[string]any
	Tasks       []map[string]any
	Decisions   []map[string]any
	Research    []map[string]any
	Knowledge   []map[string]any
	Visions     []map[string]any
	Validations []map[string]any
	Approved    struct {
		Requirements []string
		Decisions    []string
		Constraints  []string
	}
}

// ──────────────────────────────────────────────
// Time helper
// ──────────────────────────────────────────────

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}
