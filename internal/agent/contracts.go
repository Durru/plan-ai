package agent

// ──────────────────────────────────────────────
// Agent Contracts
// ──────────────────────────────────────────────

// IntentDetector detects user intent from input.
type IntentDetector interface {
	DetectIntent(input string) IntentKind
}

// Router routes intents to workflows and capabilities.
type Router interface {
	Route(intent IntentKind, context ContextPayload) RouterDecision
}

// WorkflowSelector picks the right workflow for an intent.
type WorkflowSelector interface {
	Select(intent IntentKind) string
}

// CapabilitySelector picks the right capability for an intent.
type CapabilitySelector interface {
	Select(intent IntentKind) string
}

// ContextLoader loads minimal context for a given intent.
type ContextLoader interface {
	Load(projectID string, keys []string) (ContextPayload, error)
}

// Delegator creates delegated jobs.
type Delegator interface {
	CreateJob(job DelegatedJob) (DelegatedJob, error)
	GetJob(id string) (DelegatedJob, error)
	ListJobs(projectID string) ([]DelegatedJob, error)
	UpdateJobStatus(id string, status JobStatus, summary string) error
}

// ResponseBuilder builds agent responses.
type ResponseBuilder interface {
	BuildSuccess(message string, decision RouterDecision) AgentResponse
	BuildApprovalRequired(message string, decision RouterDecision) AgentResponse
	BuildError(err string) AgentResponse
}

// ──────────────────────────────────────────────
// Repository contracts
// ──────────────────────────────────────────────

// AgentRunRepository persists agent runs and messages.
type AgentRunRepository interface {
	CreateRun(run AgentRunRecord) (AgentRunRecord, error)
	GetRun(id string) (AgentRunRecord, error)
	UpdateRunStatus(id, status, response string) error
	ListRuns(projectID string, limit int) ([]AgentRunRecord, error)

	CreateMessage(msg AgentMessage) (AgentMessage, error)
	ListMessages(runID string) ([]AgentMessage, error)
}

// DelegatedJobRepository persists delegated jobs.
type DelegatedJobRepository interface {
	CreateJob(job DelegatedJob) (DelegatedJob, error)
	GetJob(id string) (DelegatedJob, error)
	ListJobs(projectID string) ([]DelegatedJob, error)
	UpdateJob(id, status, summary string) error
}
