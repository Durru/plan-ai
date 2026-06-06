// Package conversation provides the Plan-AI conversation gateway.
//
// The Gateway is the single entry point for natural-language interaction with
// Plan-AI — both the CLI ("plan-ai agent process") and the MCP server
// ("plan_ai.agent_message") route through it. It wraps the agent.Service
// with conversation-level persistence (run records and per-run messages).
//
// This is the canonical implementation of Phase 4 (Conversation Gateway).
package conversation

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/plan-ai/plan-ai/internal/agent"
	"github.com/plan-ai/plan-ai/internal/guard"
)

// Gateway routes natural-language messages through the agent system
// and persists conversation runs and messages.
type Gateway struct {
	db     *sql.DB
	svcObj *agent.Service
}

// NewGateway creates a conversation gateway backed by the given database.
// Both global and project databases are supported; the caller must ensure
// migrations have been run before using the gateway.
func NewGateway(db *sql.DB) *Gateway {
	return &Gateway{db: db}
}

// Service returns the lazy-initialized agent.Service.
func (g *Gateway) Service() *agent.Service {
	if g.svcObj != nil {
		return g.svcObj
	}

	ws := agent.NewWorkflowSelector()
	cs := agent.NewCapabilitySelector()
	detector := agent.NewIntentDetector()
	router := agent.NewRouter(ws, cs)
	contextLoader := agent.NewContextLoader(g.db)
	responseBuilder := agent.NewResponseBuilder()
	delegator := agent.NewDelegator(g.db, newDelegatedJobAdapter(g.db))
	runRepo := newRunRepoAdapter(g.db)

	g.svcObj = agent.NewService(detector, router, contextLoader, delegator, responseBuilder, runRepo)
	g.svcObj.SetPlanningGuard(&agentGuardAdapter{g.db})
	return g.svcObj
}

// agentGuardAdapter adapts guard.PlanningGuard to agent.PlanningGuard.
type agentGuardAdapter struct {
	db *sql.DB
}

func (a *agentGuardAdapter) IsPlanningAllowed(projectID string) (bool, string) {
	return guard.NewPlanningGuard(a.db).Check(projectID)
}

// ProcessMessage handles a user message and returns a structured response.
// It persists the conversation run and both the user message and agent
// response as records in the agent_runs_v2 and agent_messages tables.
func (g *Gateway) ProcessMessage(projectID string, message string) (agent.AgentResponse, error) {
	svc := g.Service()
	resp, err := svc.ProcessMessage(projectID, message)
	if err != nil {
		return resp, err
	}

	run := agent.AgentRunRecord{
		ID:        fmt.Sprintf("conv_%d", time.Now().UnixNano()),
		ProjectID: projectID,
		Intent:    string(resp.WorkflowTriggered),
		Status:    "processed",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	encoded, err := json.Marshal(resp)
	if err == nil {
		run.Response = string(encoded)
	}

	r := newRunRepoAdapter(g.db)
	if _, err := r.CreateRun(run); err != nil {
		return resp, fmt.Errorf("persist conversation run: %w", err)
	}

	if _, err := r.CreateMessage(agent.AgentMessage{
		ID:        fmt.Sprintf("umsg_%d", time.Now().UnixNano()),
		RunID:     run.ID,
		Role:      "user",
		Content:   message,
		CreatedAt: run.CreatedAt,
	}); err != nil {
		return resp, fmt.Errorf("persist user message: %w", err)
	}

	if _, err := r.CreateMessage(agent.AgentMessage{
		ID:        fmt.Sprintf("amsg_%d", time.Now().UnixNano()),
		RunID:     run.ID,
		Role:      "agent",
		Content:   resp.Message,
		CreatedAt: run.CreatedAt,
	}); err != nil {
		return resp, fmt.Errorf("persist agent message: %w", err)
	}

	return resp, nil
}

// ListRuns returns recent conversation runs for a project.
func (g *Gateway) ListRuns(projectID string, limit int) ([]agent.AgentRunRecord, error) {
	r := newRunRepoAdapter(g.db)
	return r.ListRuns(projectID, limit)
}
