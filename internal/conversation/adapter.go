package conversation

import (
	"database/sql"

	"github.com/Durru/plan-ai/internal/agent"
	"github.com/Durru/plan-ai/internal/store"
)

// runRepoAdapter adapts store.AgentRunV2Repository to the agent.AgentRunRepository
// interface so the conversation gateway (and any other package outside cmd/plan-ai)
// can use the store without depending on cmd-level type wrappers.
type runRepoAdapter struct {
	db *sql.DB
}

func newRunRepoAdapter(db *sql.DB) *runRepoAdapter {
	return &runRepoAdapter{db: db}
}

func (w *runRepoAdapter) repo() *store.AgentRunV2Repository {
	return store.NewAgentRunV2Repository(w.db)
}

func (w *runRepoAdapter) CreateRun(run agent.AgentRunRecord) (agent.AgentRunRecord, error) {
	rec, err := w.repo().CreateRun(toStoreRun(run))
	if err != nil {
		return agent.AgentRunRecord{}, err
	}
	return toAgentRun(rec), nil
}

func (w *runRepoAdapter) GetRun(id string) (agent.AgentRunRecord, error) {
	rec, err := w.repo().GetRun(id)
	if err != nil {
		return agent.AgentRunRecord{}, err
	}
	return toAgentRun(rec), nil
}

func (w *runRepoAdapter) UpdateRunStatus(id, status, response string) error {
	return w.repo().UpdateRunStatus(id, status, response)
}

func (w *runRepoAdapter) ListRuns(projectID string, limit int) ([]agent.AgentRunRecord, error) {
	records, err := w.repo().ListRuns(projectID, limit)
	if err != nil {
		return nil, err
	}
	runs := make([]agent.AgentRunRecord, len(records))
	for i, r := range records {
		runs[i] = toAgentRun(r)
	}
	return runs, nil
}

func (w *runRepoAdapter) CreateMessage(msg agent.AgentMessage) (agent.AgentMessage, error) {
	storeMsg, err := w.repo().CreateMessage(toStoreMsg(msg))
	if err != nil {
		return agent.AgentMessage{}, err
	}
	return toAgentMsg(storeMsg), nil
}

func (w *runRepoAdapter) ListMessages(runID string) ([]agent.AgentMessage, error) {
	records, err := w.repo().ListMessages(runID)
	if err != nil {
		return nil, err
	}
	msgs := make([]agent.AgentMessage, len(records))
	for i, m := range records {
		msgs[i] = toAgentMsg(m)
	}
	return msgs, nil
}

func toStoreRun(run agent.AgentRunRecord) store.AgentRunV2Record {
	return store.AgentRunV2Record{
		ID: run.ID, ProjectID: run.ProjectID, Intent: run.Intent,
		Status: run.Status, Response: run.Response,
		CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt,
	}
}

func toAgentRun(rec store.AgentRunV2Record) agent.AgentRunRecord {
	return agent.AgentRunRecord{
		ID: rec.ID, ProjectID: rec.ProjectID, Intent: rec.Intent,
		Status: rec.Status, Response: rec.Response,
		CreatedAt: rec.CreatedAt, UpdatedAt: rec.UpdatedAt,
	}
}

func toStoreMsg(msg agent.AgentMessage) store.AgentMessageRecord {
	return store.AgentMessageRecord{
		ID: msg.ID, RunID: msg.RunID, Role: msg.Role,
		Content: msg.Content, CreatedAt: msg.CreatedAt,
	}
}

func toAgentMsg(rec store.AgentMessageRecord) agent.AgentMessage {
	return agent.AgentMessage{
		ID: rec.ID, RunID: rec.RunID, Role: rec.Role,
		Content: rec.Content, CreatedAt: rec.CreatedAt,
	}
}

// delegatedJobAdapter adapts store.DelegatedJobRepository to agent.DelegatedJobRepository.
type delegatedJobAdapter struct {
	db *sql.DB
}

func newDelegatedJobAdapter(db *sql.DB) *delegatedJobAdapter {
	return &delegatedJobAdapter{db: db}
}

func (w *delegatedJobAdapter) repo() *store.DelegatedJobRepository {
	return store.NewDelegatedJobRepository(w.db)
}

func (w *delegatedJobAdapter) CreateJob(job agent.DelegatedJob) (agent.DelegatedJob, error) {
	storeJob, err := w.repo().CreateJob(toStoreDelegated(job))
	if err != nil {
		return agent.DelegatedJob{}, err
	}
	return toAgentDelegated(storeJob), nil
}

func (w *delegatedJobAdapter) GetJob(id string) (agent.DelegatedJob, error) {
	rec, err := w.repo().GetJob(id)
	if err != nil {
		return agent.DelegatedJob{}, err
	}
	return toAgentDelegated(rec), nil
}

func (w *delegatedJobAdapter) ListJobs(projectID string) ([]agent.DelegatedJob, error) {
	records, err := w.repo().ListJobs(projectID)
	if err != nil {
		return nil, err
	}
	jobs := make([]agent.DelegatedJob, len(records))
	for i, r := range records {
		jobs[i] = toAgentDelegated(r)
	}
	return jobs, nil
}

func (w *delegatedJobAdapter) UpdateJob(id, status, summary string) error {
	return w.repo().UpdateJob(id, status, summary)
}

func toStoreDelegated(job agent.DelegatedJob) store.AgentDelegatedJobRecord {
	return store.AgentDelegatedJobRecord{
		ID: job.ID, ProjectID: job.ProjectID, Intent: string(job.Intent),
		Capability: job.Capability, WorkflowType: job.WorkflowType,
		JobType: string(job.JobType), Status: string(job.Status),
		ResultSummary: job.ResultSummary, CreatedAt: job.CreatedAt,
		CompletedAt: job.CompletedAt,
	}
}

func toAgentDelegated(rec store.AgentDelegatedJobRecord) agent.DelegatedJob {
	return agent.DelegatedJob{
		ID: rec.ID, ProjectID: rec.ProjectID, Intent: agent.IntentKind(rec.Intent),
		Capability: rec.Capability, WorkflowType: rec.WorkflowType,
		JobType: agent.DelegatedJobType(rec.JobType), Status: agent.JobStatus(rec.Status),
		ResultSummary: rec.ResultSummary, CreatedAt: rec.CreatedAt,
		CompletedAt: rec.CompletedAt,
	}
}
