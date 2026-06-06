package continuous

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

// LoopService orchestrates the full continuous planning cycle:
// detect → analyze → propose → approve → apply.
//
// It is the single entry point for both CLI and MCP continuous planning
// workflows, ensuring consistent invariants across all paths.
type LoopService struct {
	detector  *Detector
	planner   *Planner
	approval  *ApprovalService
	updater   *Updater
	statusSvc *StatusService
	db        *sql.DB
}

// NewLoopService creates a loop service with all dependencies.
func NewLoopService(db *sql.DB, eventRepo ContinuousEventRepository, proposalRepo PlanUpdateProposalRepository) *LoopService {
	return &LoopService{
		detector:  NewDetector(db),
		planner:   NewPlanner(proposalRepo),
		approval:  NewApprovalService(proposalRepo),
		updater:   NewUpdater(proposalRepo),
		statusSvc: NewStatusService(db),
		db:        db,
	}
}

// DetectAndPropose detects recent events and creates plan update proposals
// from them. Returns the created proposals and any error.
func (s *LoopService) DetectAndPropose(projectID string) ([]PlanUpdateProposal, error) {
	events, err := s.detector.Detect(projectID)
	if err != nil {
		return nil, fmt.Errorf("detect: %w", err)
	}

	var proposals []PlanUpdateProposal
	for _, ev := range events {
		outdated, err := s.detector.DetectOutdatedPlans(projectID)
		if err != nil {
			continue
		}
		affected := append(outdated, string(ev.EventType))
		prop, err := s.planner.CreateProposal(projectID, ev, affected, []string{}, []string{})
		if err != nil {
			return proposals, err
		}
		proposals = append(proposals, prop)
	}

	return proposals, nil
}

// ApproveProposal moves a proposal from draft/pending to approved.
func (s *LoopService) ApproveProposal(proposalID string) (PlanUpdateProposal, error) {
	prop, err := s.updater.Approve(proposalID)
	if err != nil {
		return prop, err
	}
	// Record memory entry for the approval.
	s.recordMemory(prop.ProjectID, "proposal_approved", fmt.Sprintf("Proposal %s approved: %s", proposalID, prop.Reason))
	return prop, nil
}

// RejectProposal moves a proposal to rejected status.
func (s *LoopService) RejectProposal(proposalID string) (PlanUpdateProposal, error) {
	return s.updater.Reject(proposalID)
}

// ApplyProposal applies an approved proposal. It marks the proposal as
// applied and records the application in project memory. Idempotent —
// calling it on an already-applied proposal returns successfully.
func (s *LoopService) ApplyProposal(proposalID string) (PlanUpdateProposal, error) {
	prop, err := s.updater.Apply(proposalID)
	if err != nil {
		return prop, err
	}
	// Record memory for the applied change.
	s.recordMemory(prop.ProjectID, "proposal_applied", fmt.Sprintf("Applied proposal %s: %s", proposalID, prop.Reason))
	return prop, nil
}

// RunLoop executes the full detect → propose → approve → apply cycle
// for a project. Proposals that require approval are left in pending;
// proposals that can be auto-applied are applied immediately.
func (s *LoopService) RunLoop(projectID string) (*LoopResult, error) {
	result := &LoopResult{ProjectID: projectID}

	// 1. Detect
	events, err := s.detector.Detect(projectID)
	if err != nil {
		return result, fmt.Errorf("detect: %w", err)
	}
	result.EventsDetected = len(events)

	// 2. Propose from events
	for _, ev := range events {
		outdated, _ := s.detector.DetectOutdatedPlans(projectID)
		affected := append(outdated, string(ev.EventType))
		prop, err := s.planner.CreateProposal(projectID, ev, affected, []string{}, []string{})
		if err != nil {
			return result, fmt.Errorf("propose: %w", err)
		}
		result.ProposalsCreated = append(result.ProposalsCreated, prop)
	}

	// 3. Request approval for draft proposals
	for _, prop := range result.ProposalsCreated {
		if _, err := s.approval.RequestApproval(prop.ID); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("request approval %s: %v", prop.ID, err))
		} else {
			result.ApprovalsRequested++
		}
	}

	// 4. Status
	status, err := s.statusSvc.GetStatus(projectID)
	if err == nil {
		result.Status = &status
	}

	return result, nil
}

// LoopResult captures the outcome of a continuous planning loop run.
type LoopResult struct {
	ProjectID          string             `json:"project_id"`
	EventsDetected     int                `json:"events_detected"`
	ProposalsCreated   []PlanUpdateProposal `json:"proposals_created"`
	ApprovalsRequested int                `json:"approvals_requested"`
	Status             *ProjectStatus     `json:"status,omitempty"`
	Errors             []string           `json:"errors,omitempty"`
}

// ── memory recording ──

func (s *LoopService) recordMemory(projectID, eventType, summary string) {
	if s.db == nil {
		return
	}
	id := domain.NewID("mem")
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.Exec(`INSERT INTO project_memory_v2 (id, project_id, entry_type, title, content, source, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, projectID, "change", truncate80(summary), summary, eventType, "active", now, now); err != nil {
		fmt.Fprintf(os.Stderr, "memory record: %v\n", err)
	}
}

func truncate80(s string) string {
	if len(s) <= 80 {
		return s
	}
	return s[:77] + "..."
}
