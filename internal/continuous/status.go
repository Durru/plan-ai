package continuous

import (
	"database/sql"
	"fmt"
	"strings"
)

// StatusService generates continuous status for a project.
type StatusService struct {
	db *sql.DB
}

// NewStatusService creates a new StatusService.
func NewStatusService(db *sql.DB) *StatusService {
	return &StatusService{db: db}
}

// GetStatus returns the continuous status of a project.
func (s *StatusService) GetStatus(projectID string) (ProjectStatus, error) {
	var status ProjectStatus
	status.ProjectID = projectID

	// Find active plan
	status.ActivePlan = s.findActivePlan(projectID)

	// Count recent events
	status.RecentEvents = s.countRecentEvents(projectID)

	// Count pending proposals
	status.PendingProposals = s.countPendingProposals(projectID)

	// Find next task
	status.NextTask = s.findNextTask(projectID)

	// Check blocked items
	status.BlockedItems = s.findBlockedItems(projectID)

	// Check approvals needed
	status.ApprovalsNeeded = s.findApprovalsNeeded(projectID)

	// Check outdated plans
	status.OutdatedPlans = s.findOutdatedPlans(projectID)

	return status, nil
}

func (s *StatusService) findActivePlan(projectID string) string {
	var title string
	err := s.db.QueryRow(
		`SELECT title FROM master_plans WHERE project_id = ? AND status IN ('in_progress', 'active') ORDER BY updated_at DESC LIMIT 1`,
		projectID).Scan(&title)
	if err != nil {
		var specificTitle string
		err2 := s.db.QueryRow(
			`SELECT title FROM specific_plans WHERE project_id = ? AND status IN ('in_progress', 'active') ORDER BY updated_at DESC LIMIT 1`,
			projectID).Scan(&specificTitle)
		if err2 == nil {
			return specificTitle
		}
		return "No active plan"
	}
	return title
}

func (s *StatusService) findNextTask(projectID string) string {
	var title, id string
	err := s.db.QueryRow(
		`SELECT id, title FROM tasks WHERE project_id = ? AND status IN ('pending', 'draft')
		 ORDER BY position, created_at LIMIT 1`, projectID).Scan(&id, &title)
	if err != nil {
		return "All tasks completed"
	}
	return fmt.Sprintf("%s (%s)", title, id)
}

func (s *StatusService) countRecentEvents(projectID string) int {
	var count int
	_ = s.db.QueryRow(`SELECT COUNT(*) FROM continuous_events WHERE project_id = ?`, projectID).Scan(&count)
	return count
}

func (s *StatusService) countPendingProposals(projectID string) int {
	var count int
	_ = s.db.QueryRow(
		`SELECT COUNT(*) FROM plan_update_proposals WHERE project_id = ? AND status IN ('draft', 'pending_approval')`,
		projectID).Scan(&count)
	return count
}

func (s *StatusService) findBlockedItems(projectID string) []string {
	rows, err := s.db.Query(
		`SELECT title FROM tasks WHERE project_id = ? AND status = 'blocked' LIMIT 10`, projectID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			continue
		}
		items = append(items, title)
	}
	if items == nil {
		return []string{}
	}
	return items
}

func (s *StatusService) findApprovalsNeeded(projectID string) []string {
	rows, err := s.db.Query(
		`SELECT reason FROM plan_update_proposals WHERE project_id = ? AND status IN ('draft', 'pending_approval') LIMIT 10`,
		projectID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var reason string
		if err := rows.Scan(&reason); err != nil {
			continue
		}
		items = append(items, reason)
	}
	return items
}

func (s *StatusService) findOutdatedPlans(projectID string) []string {
	// Check for events that typically make plans outdated
	var eventCount int
	_ = s.db.QueryRow(
		`SELECT COUNT(*) FROM continuous_events
		 WHERE project_id = ? AND event_type IN (?, ?, ?)
		 AND created_at > (SELECT COALESCE(MIN(created_at), '1970-01-01') FROM plan_update_proposals WHERE project_id = ? AND status = 'applied')`,
		projectID, EventDecisionChanged, EventNewApprovedContext, EventChangeRequestCreated, projectID).Scan(&eventCount)
	if eventCount > 0 {
		return []string{fmt.Sprintf("Potentially outdated — %d relevant events since last applied update", eventCount)}
	}
	return []string{}
}

// ──────────────────────────────────────────────
// Context Level Generation
// ──────────────────────────────────────────────

// ContextGenerator generates context at different detail levels.
type ContextGenerator struct {
	db *sql.DB
}

// NewContextGenerator creates a new ContextGenerator.
func NewContextGenerator(db *sql.DB) *ContextGenerator {
	return &ContextGenerator{db: db}
}

// Generate generates context at the requested level.
func (g *ContextGenerator) Generate(projectID string, level ContextLevel) (string, error) {
	switch level {
	case ContextL0Executive:
		return g.generateL0(projectID)
	case ContextL1Planning:
		return g.generateL1(projectID)
	case ContextL2Plan:
		return g.generateL2(projectID)
	case ContextL3Task:
		return g.generateL3(projectID)
	case ContextL4Implementation:
		return g.generateL4(projectID)
	default:
		return g.generateL0(projectID)
	}
}

func (g *ContextGenerator) generateL0(projectID string) (string, error) {
	var b strings.Builder
	b.WriteString("# Executive Summary\n\n")

	var planCount, taskCount, decisionCount int
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM master_plans WHERE project_id = ?`, projectID).Scan(&planCount)
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status IN ('pending', 'draft')`, projectID).Scan(&taskCount)
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM decisions WHERE project_id = ?`, projectID).Scan(&decisionCount)

	b.WriteString(fmt.Sprintf("**Plans**: %d  \n", planCount))
	b.WriteString(fmt.Sprintf("**Pending tasks**: %d  \n", taskCount))
	b.WriteString(fmt.Sprintf("**Decisions**: %d  \n", decisionCount))
	b.WriteString(fmt.Sprintf("**Progress**: %s  \n", g.progressIndicator(projectID)))

	nextTask := ""
	_ = g.db.QueryRow(`SELECT title FROM tasks WHERE project_id = ? AND status IN ('pending', 'draft') ORDER BY position, created_at LIMIT 1`, projectID).Scan(&nextTask)
	if nextTask != "" {
		b.WriteString(fmt.Sprintf("\n**Next step**: %s\n", nextTask))
	} else {
		b.WriteString("\n**Next step**: No pending tasks\n")
	}

	// Blockers
	var blockers int
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status = 'blocked'`, projectID).Scan(&blockers)
	if blockers > 0 {
		b.WriteString(fmt.Sprintf("\n**Blockers**: %d task(s) blocked\n", blockers))
	}

	return b.String(), nil
}

func (g *ContextGenerator) generateL1(projectID string) (string, error) {
	var b strings.Builder
	b.WriteString("# Planning Context\n\n")

	// Vision
	var visionTitle, visionSummary string
	err := g.db.QueryRow(`SELECT title, summary FROM visions WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`, projectID).Scan(&visionTitle, &visionSummary)
	if err == nil {
		b.WriteString("## Vision\n")
		b.WriteString(fmt.Sprintf("**%s**: %s\n\n", visionTitle, visionSummary))
	}

	// Requirements
	reqRows, err := g.db.Query(`SELECT content FROM approved_requirements WHERE project_id = ? LIMIT 20`, projectID)
	if err == nil {
		defer reqRows.Close()
		b.WriteString("## Approved Requirements\n")
		for reqRows.Next() {
			var content string
			if err := reqRows.Scan(&content); err != nil {
				continue
			}
			b.WriteString(fmt.Sprintf("- %s\n", content))
		}
		b.WriteString("\n")
	}

	// Decisions
	decRows, err := g.db.Query(`SELECT title, decision FROM decisions WHERE project_id = ? ORDER BY created_at DESC LIMIT 10`, projectID)
	if err == nil {
		defer decRows.Close()
		b.WriteString("## Key Decisions\n")
		for decRows.Next() {
			var title, decision string
			if err := decRows.Scan(&title, &decision); err != nil {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", title, decision))
		}
		b.WriteString("\n")
	}

	// Research
	resRows, err := g.db.Query(`SELECT topic, summary FROM research_entries WHERE project_id = ? ORDER BY created_at DESC LIMIT 5`, projectID)
	if err == nil {
		defer resRows.Close()
		b.WriteString("## Relevant Research\n")
		for resRows.Next() {
			var topic, summary string
			if err := resRows.Scan(&topic, &summary); err != nil {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", topic, summary))
		}
		b.WriteString("\n")
	}

	// Knowledge
	knRows, err := g.db.Query(`SELECT topic, summary FROM knowledge_objects WHERE project_id = ? ORDER BY reuse_count DESC LIMIT 5`, projectID)
	if err == nil {
		defer knRows.Close()
		b.WriteString("## Relevant Knowledge\n")
		for knRows.Next() {
			var topic, summary string
			if err := knRows.Scan(&topic, &summary); err != nil {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s**: %s\n", topic, summary))
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (g *ContextGenerator) generateL2(projectID string) (string, error) {
	var b strings.Builder
	b.WriteString("# Specific Plan Context\n\n")

	rows, err := g.db.Query(
		`SELECT sp.id, sp.title, sp.summary, sp.status, sp.goal
		 FROM specific_plans sp WHERE sp.project_id = ? ORDER BY sp.created_at DESC LIMIT 5`, projectID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, summary, status, goal string
		if err := rows.Scan(&id, &title, &summary, &status, &goal); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("## %s (%s)\n", title, status))
		b.WriteString(fmt.Sprintf("**Summary**: %s\n", summary))
		if goal != "" {
			b.WriteString(fmt.Sprintf("**Goal**: %s\n", goal))
		}

		// Related decisions
		decRows, err := g.db.Query(
			`SELECT title, decision FROM decisions WHERE project_id = ? ORDER BY created_at DESC LIMIT 5`, projectID)
		if err == nil {
			b.WriteString("\n**Related Decisions**:\n")
			for decRows.Next() {
				var dt, dd string
				if err := decRows.Scan(&dt, &dd); err != nil {
					continue
				}
				b.WriteString(fmt.Sprintf("- %s: %s\n", dt, dd))
			}
			decRows.Close()
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (g *ContextGenerator) generateL3(projectID string) (string, error) {
	var b strings.Builder
	b.WriteString("# Task Context\n\n")

	rows, err := g.db.Query(
		`SELECT t.id, t.title, t.summary, t.status, t.position, p.title as phase
		 FROM tasks t LEFT JOIN phases p ON t.phase_id = p.id
		 WHERE t.project_id = ? ORDER BY t.position, t.created_at LIMIT 5`, projectID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, summary, status string
		var position int
		var phase string
		if err := rows.Scan(&id, &title, &summary, &status, &position, &phase); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("## %s (%s)\n", title, status))
		if phase != "" {
			b.WriteString(fmt.Sprintf("**Phase**: %s\n", phase))
		}
		b.WriteString(fmt.Sprintf("**Position**: %d\n", position))
		b.WriteString(fmt.Sprintf("**Summary**: %s\n", summary))

		// Validations
		vRows, err := g.db.Query(
			`SELECT status, summary FROM validations WHERE target_type = 'task' AND target_id = ? LIMIT 5`, id)
		if err == nil {
			b.WriteString("\n**Validations**:\n")
			for vRows.Next() {
				var vs, vsum string
				if err := vRows.Scan(&vs, &vsum); err != nil {
					continue
				}
				b.WriteString(fmt.Sprintf("- [%s] %s\n", vs, vsum))
			}
			vRows.Close()
		}
		b.WriteString("\n")
	}

	return b.String(), nil
}

func (g *ContextGenerator) generateL4(projectID string) (string, error) {
	var b strings.Builder
	b.WriteString("# Implementation Context\n\n")

	// Implementation documents
	rows, err := g.db.Query(
		`SELECT id, title, objective, content, architecture, expected_files, validations, known_risks, testing_strategy, rollback_strategy
		 FROM implementation_documents WHERE project_id = ? ORDER BY version DESC LIMIT 3`, projectID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var id, title, objective, content, architecture, expectedFiles, validations, knownRisks, testingStrategy, rollbackStrategy string
		if err := rows.Scan(&id, &title, &objective, &content, &architecture, &expectedFiles, &validations, &knownRisks, &testingStrategy, &rollbackStrategy); err != nil {
			continue
		}
		b.WriteString(fmt.Sprintf("## %s\n", title))
		if objective != "" {
			b.WriteString(fmt.Sprintf("**Objective**: %s\n\n", objective))
		}
		if architecture != "" {
			b.WriteString(fmt.Sprintf("**Architecture**: %s\n\n", architecture))
		}
		if expectedFiles != "[]" && expectedFiles != "" {
			b.WriteString(fmt.Sprintf("**Expected Files**: %s\n\n", expectedFiles))
		}
		if testingStrategy != "" {
			b.WriteString(fmt.Sprintf("**Testing**: %s\n\n", testingStrategy))
		}
		if rollbackStrategy != "" {
			b.WriteString(fmt.Sprintf("**Rollback**: %s\n\n", rollbackStrategy))
		}
		if knownRisks != "[]" && knownRisks != "" {
			b.WriteString(fmt.Sprintf("**Known Risks**: %s\n\n", knownRisks))
		}
		if validations != "[]" && validations != "" {
			b.WriteString(fmt.Sprintf("**Validations**: %s\n\n", validations))
		}
	}

	// Decisions needed
	b.WriteString("\n## Relevant Decisions\n")
	decRows, err := g.db.Query(
		`SELECT title, decision, rationale FROM decisions WHERE project_id = ? ORDER BY created_at DESC LIMIT 5`, projectID)
	if err == nil {
		defer decRows.Close()
		for decRows.Next() {
			var title, decision, rationale string
			if err := decRows.Scan(&title, &decision, &rationale); err != nil {
				continue
			}
			b.WriteString(fmt.Sprintf("- **%s**: %s", title, decision))
			if rationale != "" {
				b.WriteString(fmt.Sprintf(" (%s)", rationale))
			}
			b.WriteString("\n")
		}
	}

	return b.String(), nil
}

func (g *ContextGenerator) progressIndicator(projectID string) string {
	var total, completed int
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ?`, projectID).Scan(&total)
	_ = g.db.QueryRow(`SELECT COUNT(*) FROM tasks WHERE project_id = ? AND status = 'completed'`, projectID).Scan(&completed)
	if total == 0 {
		return "Not started"
	}
	return fmt.Sprintf("%d/%d tasks completed (%.0f%%)", completed, total, float64(completed)/float64(total)*100)
}
