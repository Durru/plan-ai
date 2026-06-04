package mcp

import (
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
	"github.com/plan-ai/plan-ai/internal/intentv3"
	"github.com/plan-ai/plan-ai/internal/store"
)

// ── Helpers ──

func getStringArg(args map[string]any, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func getIntArg(args map[string]any, key string) int {
	v, ok := args[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}

func getProjectRoot(args map[string]any) (string, error) {
	if root := getStringArg(args, "project_root"); root != "" {
		return root, nil
	}
	return store.ResolveProjectRoot()
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// openStore opens the project store for the given root path.
// Callers must close the store when done.
func openStore(projectRoot string) (*store.ProjectStore, error) {
	return store.OpenProjectStore(projectRoot)
}

// projectID returns the canonical project ID for the given root path.
func projectID(projectRoot string) string {
	return store.ProjectID(projectRoot)
}

// ── Core Project Handlers ──

func HandleInitProject(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	name := getStringArg(args, "name")
	if name == "" {
		name = projectRoot
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("init project store: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)
	if err := store.UpsertProjectState(ps.DB, pid, name, projectRoot, "active"); err != nil {
		return nil, fmt.Errorf("save project: %w", err)
	}

	if err := store.UpsertKnownProject(ps.DB, pid, name, projectRoot); err != nil {
		return nil, fmt.Errorf("save known project: %w", err)
	}

	return map[string]any{
		"status":       "initialized",
		"project_root": projectRoot,
		"project_id":   pid,
		"db_path":      ps.Layout.DBPath,
	}, nil
}

func HandleProjectStatus(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	counts, err := store.CountDomainEntities(ps.DB)
	if err != nil {
		return nil, fmt.Errorf("count entities: %w", err)
	}

	return map[string]any{
		"project_root":      projectRoot,
		"plans":             counts.Plans,
		"phases":            counts.Phases,
		"tasks":             counts.Tasks,
		"decisions":         counts.Decisions,
		"research_entries":  counts.ResearchEntries,
		"knowledge_objects": counts.KnowledgeObjects,
		"validations":       counts.Validations,
		"snapshots":         counts.Snapshots,
	}, nil
}

func HandleCreateMasterPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	plan := domain.MasterPlan{
		ID:        domain.NewID("plan"),
		ProjectID: pid,
		Title:     getStringArg(args, "title"),
		Summary:   getStringArg(args, "summary"),
		Status:    domain.StatusDraft,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Plan.SaveMaster(plan); err != nil {
		return nil, fmt.Errorf("save master plan: %w", err)
	}

	return map[string]any{
		"plan_id":    plan.ID,
		"title":      plan.Title,
		"summary":    plan.Summary,
		"status":     string(plan.Status),
		"version":    plan.Version,
		"created_at": formatTime(plan.CreatedAt),
	}, nil
}

func HandleCreateSpecificPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)
	masterPlanID := getStringArg(args, "master_plan_id")
	goal := getStringArg(args, "goal")

	plan := domain.SpecificPlan{
		ID:           domain.NewID("plan"),
		ProjectID:    pid,
		MasterPlanID: masterPlanID,
		Title:        getStringArg(args, "title"),
		Summary:      goal,
		Status:       domain.StatusDraft,
		Version:      1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := repos.Plan.SaveSpecific(plan); err != nil {
		return nil, fmt.Errorf("save specific plan: %w", err)
	}

	return map[string]any{
		"plan_id":        plan.ID,
		"master_plan_id": plan.MasterPlanID,
		"title":          plan.Title,
		"goal":           plan.Summary,
		"status":         string(plan.Status),
		"created_at":     formatTime(plan.CreatedAt),
	}, nil
}

func HandleResearchTopic(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	research := domain.Research{
		ID:         domain.NewID("research"),
		ProjectID:  pid,
		Topic:      getStringArg(args, "topic"),
		Summary:    getStringArg(args, "summary"),
		Confidence: float64(getIntArg(args, "confidence")),
		Status:     domain.ResearchStatusDraft,
		Category:   domain.KnowledgeCategoryGeneral,
		Date:       time.Now(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repos.Research.Save(research); err != nil {
		return nil, fmt.Errorf("save research: %w", err)
	}

	return map[string]any{
		"research_id": research.ID,
		"topic":       research.Topic,
		"summary":     research.Summary,
		"status":      string(research.Status),
		"confidence":  research.Confidence,
		"created_at":  formatTime(research.CreatedAt),
	}, nil
}

func HandleApprovePlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")

	if err := repos.Plan.UpdatePlanStatus(planID, domain.StatusApproved); err != nil {
		return nil, fmt.Errorf("approve plan: %w", err)
	}

	return map[string]any{
		"plan_id": planID,
		"status":  "approved",
	}, nil
}

func HandleRejectPlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")

	if err := repos.Plan.UpdatePlanStatus(planID, domain.StatusRejected); err != nil {
		return nil, fmt.Errorf("reject plan: %w", err)
	}

	return map[string]any{
		"plan_id": planID,
		"status":  "rejected",
	}, nil
}

func HandleAnalyzeImpact(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)
	changeType := getStringArg(args, "change_type")
	summary := getStringArg(args, "summary")

	cr := domain.ChangeRequest{
		ID:          domain.NewID("change"),
		ProjectID:   pid,
		Reason:      changeType,
		Description: summary,
		Status:      domain.ChangeRequestDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repos.Change.SaveChangeRequest(cr); err != nil {
		return nil, fmt.Errorf("save change request: %w", err)
	}

	report := domain.ImpactReport{
		ID:              domain.NewID("impact"),
		ChangeRequestID: cr.ID,
		Summary:         fmt.Sprintf("Impact analysis for %s: %s", changeType, summary),
		CreatedAt:       time.Now(),
	}
	if err := repos.Change.SaveImpactReport(report); err != nil {
		return nil, fmt.Errorf("save impact report: %w", err)
	}

	return map[string]any{
		"change_request_id": cr.ID,
		"change_type":       changeType,
		"status":            string(cr.Status),
		"impact_report_id":  report.ID,
		"summary":           summary,
	}, nil
}

func HandleGetNextTask(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	var id, title, summary, status, phaseID, planID string
	var position int
	err = ps.DB.QueryRow(`SELECT id, phase_id, plan_id, title, summary, status, position FROM tasks WHERE status = 'pending' ORDER BY position, created_at LIMIT 1`).Scan(&id, &phaseID, &planID, &title, &summary, &status, &position)
	if err != nil {
		return map[string]any{
			"found": false,
			"error": "no pending tasks found",
		}, nil
	}

	return map[string]any{
		"found":    true,
		"task_id":  id,
		"phase_id": phaseID,
		"plan_id":  planID,
		"title":    title,
		"summary":  summary,
		"status":   status,
		"position": position,
	}, nil
}

func HandleMarkTaskDone(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	taskID := getStringArg(args, "task_id")

	if err := repos.Task.UpdateStatus(taskID, domain.PlanStatusDone); err != nil {
		return nil, fmt.Errorf("update task status: %w", err)
	}

	return map[string]any{
		"task_id": taskID,
		"status":  "done",
	}, nil
}

func HandleCreateSnapshot(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	snapshot := domain.Snapshot{
		ID:        domain.NewID("snapshot"),
		ProjectID: pid,
		Reason:    getStringArg(args, "reason"),
		Summary:   getStringArg(args, "summary"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := repos.Snapshot.Save(snapshot); err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	return map[string]any{
		"snapshot_id": snapshot.ID,
		"reason":      snapshot.Reason,
		"summary":     snapshot.Summary,
		"created_at":  formatTime(snapshot.CreatedAt),
	}, nil
}

func HandleListPlans(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	masters, err := repos.Plan.ListMastersByProject(pid)
	if err != nil {
		return nil, fmt.Errorf("list master plans: %w", err)
	}

	specifics, err := repos.Plan.ListSpecificsByMaster("")
	if err != nil {
		return nil, fmt.Errorf("list specific plans: %w", err)
	}

	masterList := make([]map[string]any, 0, len(masters))
	for _, m := range masters {
		masterList = append(masterList, map[string]any{
			"id":         m.ID,
			"title":      m.Title,
			"summary":    m.Summary,
			"status":     string(m.Status),
			"version":    m.Version,
			"created_at": formatTime(m.CreatedAt),
		})
	}

	specificList := make([]map[string]any, 0, len(specifics))
	for _, s := range specifics {
		specificList = append(specificList, map[string]any{
			"id":             s.ID,
			"master_plan_id": s.MasterPlanID,
			"title":          s.Title,
			"summary":        s.Summary,
			"status":         string(s.Status),
			"version":        s.Version,
			"created_at":     formatTime(s.CreatedAt),
		})
	}

	return map[string]any{
		"master_plans":   masterList,
		"specific_plans": specificList,
	}, nil
}

func HandleListTasks(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	planID := getStringArg(args, "plan_id")
	statusFilter := getStringArg(args, "status")

	query := `SELECT id, phase_id, plan_id, title, summary, status, position, created_at, updated_at FROM tasks`
	var conditions []string
	var params []any

	if planID != "" {
		conditions = append(conditions, "plan_id = ?")
		params = append(params, planID)
	}
	if statusFilter != "" {
		conditions = append(conditions, "status = ?")
		params = append(params, statusFilter)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY position, created_at"

	rows, err := ps.DB.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query tasks: %w", err)
	}
	defer rows.Close()

	taskList := make([]map[string]any, 0)
	for rows.Next() {
		var id, phaseID, tPlanID, title, summary, status, createdAt, updatedAt string
		var position int
		if err := rows.Scan(&id, &phaseID, &tPlanID, &title, &summary, &status, &position, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		taskList = append(taskList, map[string]any{
			"id":         id,
			"phase_id":   phaseID,
			"plan_id":    tPlanID,
			"title":      title,
			"summary":    summary,
			"status":     status,
			"position":   position,
			"created_at": createdAt,
		})
	}

	return map[string]any{
		"tasks": taskList,
		"count": len(taskList),
	}, nil
}

// ── Agent System Handlers ──

func HandleAgentProcess(args map[string]any) (map[string]any, error) {
	message := getStringArg(args, "message")
	return map[string]any{
		"status":  "processed",
		"message": fmt.Sprintf("Agent processed message (%d chars)", len(message)),
		"intent":  "unknown",
		"response": map[string]any{
			"text": "Agent processing is a stub. Connect real agent services.",
		},
	}, nil
}

func HandleAgentRuns(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)
	limit := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	rows, err := ps.DB.Query(`SELECT id, intent, status, response, created_at FROM agent_runs_v2 WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, pid, limit)
	if err != nil {
		return map[string]any{
			"runs":  []any{},
			"count": 0,
		}, nil
	}
	defer rows.Close()

	runs := make([]map[string]any, 0)
	for rows.Next() {
		var id, intent, status, response, createdAt string
		if err := rows.Scan(&id, &intent, &status, &response, &createdAt); err != nil {
			continue
		}
		runs = append(runs, map[string]any{
			"id":         id,
			"intent":     intent,
			"status":     status,
			"response":   response,
			"created_at": createdAt,
		})
	}

	return map[string]any{
		"runs":  runs,
		"count": len(runs),
	}, nil
}

// ── Continuous Planning Handlers ──

func HandleContinuousStatus(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)

	var id, activePlan, activePhase, nextTask string
	err = ps.DB.QueryRow(`SELECT id, active_plan, active_phase, next_task FROM continuous_status WHERE project_id = ? ORDER BY created_at DESC LIMIT 1`, pid).Scan(&id, &activePlan, &activePhase, &nextTask)
	if err != nil {
		return map[string]any{
			"status":       "inactive",
			"project_root": projectRoot,
		}, nil
	}

	return map[string]any{
		"status":       "active",
		"active_plan":  activePlan,
		"active_phase": activePhase,
		"next_task":    nextTask,
	}, nil
}

func HandleContinuousEvents(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)
	limit := getIntArg(args, "limit")
	if limit <= 0 {
		limit = 10
	}

	rows, err := ps.DB.Query(`SELECT id, event_type, summary, created_at FROM continuous_events WHERE project_id = ? ORDER BY created_at DESC LIMIT ?`, pid, limit)
	if err != nil {
		return map[string]any{
			"events": []any{},
			"count":  0,
		}, nil
	}
	defer rows.Close()

	events := make([]map[string]any, 0)
	for rows.Next() {
		var id, eventType, summary, createdAt string
		if err := rows.Scan(&id, &eventType, &summary, &createdAt); err != nil {
			continue
		}
		events = append(events, map[string]any{
			"id":         id,
			"event_type": eventType,
			"summary":    summary,
			"created_at": createdAt,
		})
	}

	return map[string]any{
		"events": events,
		"count":  len(events),
	}, nil
}

func HandleContinuousProposals(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)
	proposalID := getStringArg(args, "proposal_id")

	// If proposal_id is given, treat this as approve/reject
	if proposalID != "" {
		// Check if there's a status override from the caller context
		// The tool itself doesn't pass status, so we default to "approved"
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := ps.DB.Exec(`UPDATE plan_update_proposals SET status = 'approved', updated_at = ? WHERE id = ?`, now, proposalID)
		if err != nil {
			return nil, fmt.Errorf("update proposal: %w", err)
		}
		return map[string]any{
			"proposal_id": proposalID,
			"status":      "approved",
		}, nil
	}

	// If reason is given, create a new proposal
	reason := getStringArg(args, "reason")
	if reason != "" {
		id := domain.NewID("proposal")
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := ps.DB.Exec(`INSERT INTO plan_update_proposals (id, project_id, reason, status, created_at, updated_at) VALUES (?, ?, ?, 'draft', ?, ?)`, id, pid, reason, now, now)
		if err != nil {
			return nil, fmt.Errorf("create proposal: %w", err)
		}
		return map[string]any{
			"proposal_id": id,
			"reason":      reason,
			"status":      "draft",
		}, nil
	}

	// Otherwise list existing proposals
	rows, err := ps.DB.Query(`SELECT id, reason, status, created_at FROM plan_update_proposals WHERE project_id = ? ORDER BY created_at DESC LIMIT 20`, pid)
	if err != nil {
		return map[string]any{
			"proposals": []any{},
			"count":     0,
		}, nil
	}
	defer rows.Close()

	proposals := make([]map[string]any, 0)
	for rows.Next() {
		var id, reason, status, createdAt string
		if err := rows.Scan(&id, &reason, &status, &createdAt); err != nil {
			continue
		}
		proposals = append(proposals, map[string]any{
			"id":         id,
			"reason":     reason,
			"status":     status,
			"created_at": createdAt,
		})
	}

	return map[string]any{
		"proposals": proposals,
		"count":     len(proposals),
	}, nil
}

func HandleContinuousContext(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	pid := projectID(projectRoot)
	level := getStringArg(args, "level")
	if level == "" {
		level = "L0_Executive"
	}

	// Try to find a cached context delivery first
	var id, content, createdAt string
	err = ps.DB.QueryRow(`SELECT id, content, created_at FROM context_deliveries WHERE project_id = ? AND level = ? ORDER BY created_at DESC LIMIT 1`, pid, level).Scan(&id, &content, &createdAt)
	if err == nil {
		return map[string]any{
			"level":      level,
			"content":    content,
			"cached":     true,
			"created_at": createdAt,
		}, nil
	}

	// Build context from domain data
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("# Context Level: %s\n", level))
	buf.WriteString(fmt.Sprintf("Project ID: %s\n\n", pid))

	// Always include plan summaries
	planRows, err := ps.DB.Query(`SELECT id, type, title, status, summary FROM plans ORDER BY created_at`)
	if err == nil {
		buf.WriteString("## Plans\n\n")
		for planRows.Next() {
			var id, ptype, title, status, summary string
			if err := planRows.Scan(&id, &ptype, &title, &status, &summary); err == nil {
				buf.WriteString(fmt.Sprintf("- **%s** [%s] (%s) - %s\n", title, ptype, status, id))
				if summary != "" {
					buf.WriteString(fmt.Sprintf("  %s\n", summary))
				}
			}
		}
		planRows.Close()
		buf.WriteString("\n")
	}

	switch level {
	case "L0_Executive":
		buf.WriteString("## Executive Summary\n\n")
		counts, _ := store.CountDomainEntities(ps.DB)
		buf.WriteString(fmt.Sprintf("- Plans: %d\n", counts.Plans))
		buf.WriteString(fmt.Sprintf("- Phases: %d\n", counts.Phases))
		buf.WriteString(fmt.Sprintf("- Tasks: %d\n", counts.Tasks))
		buf.WriteString(fmt.Sprintf("- Decisions: %d\n", counts.Decisions))
		buf.WriteString(fmt.Sprintf("- Research entries: %d\n", counts.ResearchEntries))

	case "L1_Planning":
		buf.WriteString("## Planning Context\n\n")
		decRows, err := ps.DB.Query(`SELECT title, decision, status FROM decisions ORDER BY created_at LIMIT 20`)
		if err == nil {
			buf.WriteString("### Decisions\n\n")
			for decRows.Next() {
				var title, decision, status string
				if err := decRows.Scan(&title, &decision, &status); err == nil {
					buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", title, status, decision))
				}
			}
			decRows.Close()
		}

		resRows, err := ps.DB.Query(`SELECT topic, status FROM research_entries ORDER BY created_at LIMIT 20`)
		if err == nil {
			buf.WriteString("\n### Research\n\n")
			for resRows.Next() {
				var topic, status string
				if err := resRows.Scan(&topic, &status); err == nil {
					buf.WriteString(fmt.Sprintf("- **%s** (%s)\n", topic, status))
				}
			}
			resRows.Close()
		}

	case "L2_Implementation":
		buf.WriteString("## Implementation Context\n\n")
		taskRows, err := ps.DB.Query(`SELECT id, title, status, summary FROM tasks ORDER BY position, created_at LIMIT 20`)
		if err == nil {
			for taskRows.Next() {
				var id, title, status, summary string
				if err := taskRows.Scan(&id, &title, &status, &summary); err == nil {
					buf.WriteString(fmt.Sprintf("### %s [%s]\n**ID:** %s\n%s\n\n", title, status, id, summary))
				}
			}
			taskRows.Close()
		}

	case "L3_Research":
		buf.WriteString("## Research Context\n\n")
		resRows, err := ps.DB.Query(`SELECT id, topic, summary, status, confidence FROM research_entries ORDER BY created_at LIMIT 30`)
		if err == nil {
			for resRows.Next() {
				var id, topic, summary, status string
				var confidence float64
				if err := resRows.Scan(&id, &topic, &summary, &status, &confidence); err == nil {
					buf.WriteString(fmt.Sprintf("### %s\n**ID:** %s | **Status:** %s | **Confidence:** %.0f\n%s\n\n", topic, id, status, confidence, summary))
				}
			}
			resRows.Close()
		}
	}

	return map[string]any{
		"level":   level,
		"content": buf.String(),
		"cached":  false,
	}, nil
}

// ── Phase 29 Handlers ──

func HandleGetContext(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	level := getStringArg(args, "level")
	if level == "" {
		level = "L0_executive"
	}
	taskID := getStringArg(args, "task_id")
	topic := getStringArg(args, "topic")

	var buf strings.Builder

	switch level {
	case "L0_executive":
		buf.WriteString("# Executive Context\n\n")
		buf.WriteString(fmt.Sprintf("Project: %s\n\n", projectRoot))

		var planCount, taskCount, decCount, resCount int
		_ = ps.DB.QueryRow(`SELECT COUNT(*) FROM plans`).Scan(&planCount)
		_ = ps.DB.QueryRow(`SELECT COUNT(*) FROM tasks`).Scan(&taskCount)
		_ = ps.DB.QueryRow(`SELECT COUNT(*) FROM decisions`).Scan(&decCount)
		_ = ps.DB.QueryRow(`SELECT COUNT(*) FROM research_entries`).Scan(&resCount)

		buf.WriteString(fmt.Sprintf("- Plans: %d\n- Tasks: %d\n- Decisions: %d\n- Research entries: %d\n", planCount, taskCount, decCount, resCount))

		// Include latest plans
		rows, err := ps.DB.Query(`SELECT title, status FROM plans ORDER BY created_at DESC LIMIT 5`)
		if err == nil {
			defer rows.Close()
			buf.WriteString("\n### Recent Plans\n")
			for rows.Next() {
				var title, status string
				if rows.Scan(&title, &status) == nil {
					buf.WriteString(fmt.Sprintf("- %s (%s)\n", title, status))
				}
			}
		}

	case "L1_planning":
		buf.WriteString("# Planning Context\n\n")
		rows, err := ps.DB.Query(`SELECT id, type, title, status, summary FROM plans ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, ptype, title, status, summary string
				if err := rows.Scan(&id, &ptype, &title, &status, &summary); err == nil {
					buf.WriteString(fmt.Sprintf("## %s [%s]\n**ID:** %s | **Status:** %s\n%s\n\n", title, ptype, id, status, summary))
				}
			}
		}

	case "L2_implementation":
		buf.WriteString("# Implementation Context\n\n")
		if taskID != "" {
			var id, title, summary, status string
			err := ps.DB.QueryRow(`SELECT id, title, summary, status FROM tasks WHERE id = ?`, taskID).Scan(&id, &title, &summary, &status)
			if err == nil {
				buf.WriteString(fmt.Sprintf("## Task: %s\n**ID:** %s | **Status:** %s\n%s\n", title, id, status, summary))
			} else {
				buf.WriteString("Task not found.\n")
			}
		} else {
			rows, err := ps.DB.Query(`SELECT id, title, status FROM tasks WHERE status IN ('pending', 'active') ORDER BY position LIMIT 10`)
			if err == nil {
				defer rows.Close()
				buf.WriteString("### Active/Pending Tasks\n")
				for rows.Next() {
					var id, title, status string
					if rows.Scan(&id, &title, &status) == nil {
						buf.WriteString(fmt.Sprintf("- [%s] %s (%s)\n", status, title, id))
					}
				}
			}
		}

	case "L3_research":
		buf.WriteString("# Research Context\n\n")
		query := `SELECT id, topic, summary, status, confidence FROM research_entries`
		var params []any
		if topic != "" {
			query += " WHERE LOWER(topic) LIKE ?"
			params = append(params, "%"+strings.ToLower(topic)+"%")
		}
		query += " ORDER BY created_at"

		rows, err := ps.DB.Query(query, params...)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, resTopic, summary, status string
				var confidence float64
				if err := rows.Scan(&id, &resTopic, &summary, &status, &confidence); err == nil {
					buf.WriteString(fmt.Sprintf("## %s\n**ID:** %s | **Status:** %s | **Confidence:** %.0f\n%s\n\n", resTopic, id, status, confidence, summary))
				}
			}
		}
	}

	return map[string]any{
		"level":   level,
		"content": buf.String(),
	}, nil
}

func HandleDetectChanges(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)
	changeType := getStringArg(args, "change_type")
	summary := getStringArg(args, "summary")
	description := getStringArg(args, "description")

	cr := domain.ChangeRequest{
		ID:          domain.NewID("change"),
		ProjectID:   pid,
		Reason:      changeType,
		Description: description,
		Status:      domain.ChangeRequestDraft,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := repos.Change.SaveChangeRequest(cr); err != nil {
		return nil, fmt.Errorf("save change request: %w", err)
	}

	// Determine affected entities based on change type
	var affectedPlans, affectedDecisions []string

	switch changeType {
	case "plan_changed", "vision_changed":
		rows, err := ps.DB.Query(`SELECT id FROM plans ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					affectedPlans = append(affectedPlans, id)
				}
			}
		}
	case "decision_changed":
		rows, err := ps.DB.Query(`SELECT id FROM decisions ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					affectedDecisions = append(affectedDecisions, id)
				}
			}
		}
	case "research_updated", "knowledge_updated":
		// General-purpose: flag all plans as potentially affected
		rows, err := ps.DB.Query(`SELECT id FROM plans ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				if rows.Scan(&id) == nil {
					affectedPlans = append(affectedPlans, id)
				}
			}
		}
	}

	report := domain.ImpactReport{
		ID:                domain.NewID("impact"),
		ChangeRequestID:   cr.ID,
		AffectedPlans:     affectedPlans,
		AffectedDecisions: affectedDecisions,
		Summary:           fmt.Sprintf("Change detected: %s - %s", changeType, summary),
		CreatedAt:         time.Now(),
	}
	if err := repos.Change.SaveImpactReport(report); err != nil {
		return nil, fmt.Errorf("save impact report: %w", err)
	}

	return map[string]any{
		"change_request_id":  cr.ID,
		"change_type":        changeType,
		"impact_report_id":   report.ID,
		"affected_plans":     affectedPlans,
		"affected_decisions": affectedDecisions,
		"summary":            summary,
	}, nil
}

func HandleUpdatePlan(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	repos := store.NewRepositories(ps.DB)
	planID := getStringArg(args, "plan_id")
	newTitle := getStringArg(args, "title")
	newSummary := getStringArg(args, "summary")
	newStatus := getStringArg(args, "status")

	// Try master plan first
	master, err := repos.Plan.GetMasterByID(planID)
	if err == nil {
		if newTitle != "" {
			master.Title = newTitle
		}
		if newSummary != "" {
			master.Summary = newSummary
		}
		if newStatus != "" {
			master.Status = domain.Status(newStatus)
		}
		master.UpdatedAt = time.Now()
		if err := repos.Plan.SaveMaster(master); err != nil {
			return nil, fmt.Errorf("update master plan: %w", err)
		}
		return map[string]any{
			"plan_id": master.ID,
			"type":    "master",
			"title":   master.Title,
			"status":  string(master.Status),
		}, nil
	}

	// Try specific plan
	specific, err := repos.Plan.GetSpecificByID(planID)
	if err == nil {
		if newTitle != "" {
			specific.Title = newTitle
		}
		if newSummary != "" {
			specific.Summary = newSummary
		}
		if newStatus != "" {
			specific.Status = domain.Status(newStatus)
		}
		specific.UpdatedAt = time.Now()
		if err := repos.Plan.SaveSpecific(specific); err != nil {
			return nil, fmt.Errorf("update specific plan: %w", err)
		}
		return map[string]any{
			"plan_id": specific.ID,
			"type":    "specific",
			"title":   specific.Title,
			"status":  string(specific.Status),
		}, nil
	}

	// Fallback: if only status is given, try UpdatePlanStatus
	if newStatus != "" {
		if err := repos.Plan.UpdatePlanStatus(planID, domain.Status(newStatus)); err != nil {
			return nil, fmt.Errorf("update plan status: %w", err)
		}
		return map[string]any{
			"plan_id": planID,
			"status":  newStatus,
		}, nil
	}

	return nil, fmt.Errorf("plan not found: %s", planID)
}

func HandleRollbackSnapshot(args map[string]any) (map[string]any, error) {
	snapshotID := getStringArg(args, "snapshot_id")
	return map[string]any{
		"supported":   false,
		"snapshot_id": snapshotID,
		"message":     "Rollback is not yet implemented. This feature will restore project state from a snapshot in a future release.",
	}, nil
}

func HandleExportDocs(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, err
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("open project: %w", err)
	}
	defer ps.Close()

	format := getStringArg(args, "format")
	if format == "" {
		format = "markdown"
	}
	scope := getStringArg(args, "scope")

	repos := store.NewRepositories(ps.DB)
	pid := projectID(projectRoot)

	var buf strings.Builder

	switch scope {
	case "plans":
		buf.WriteString("# Project Plans\n\n")
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, m := range masters {
			buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n- **Version:** %d\n\n%s\n\n", m.Title, m.ID, m.Status, m.Version, m.Summary))
		}
		specifics, _ := repos.Plan.ListSpecificsByMaster("")
		for _, s := range specifics {
			buf.WriteString(fmt.Sprintf("### %s (Master: %s)\n- **ID:** %s\n- **Status:** %s\n\n%s\n\n", s.Title, s.MasterPlanID, s.ID, s.Status, s.Summary))
		}

	case "decisions":
		buf.WriteString("# Project Decisions\n\n")
		rows, err := ps.DB.Query(`SELECT id, title, decision, context, status, impact FROM decisions ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, title, decision, context, status, impact string
				if err := rows.Scan(&id, &title, &decision, &context, &status, &impact); err == nil {
					buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n\n**Decision:** %s\n\n**Context:** %s\n\n**Impact:** %s\n\n", title, id, status, decision, context, impact))
				}
			}
		}

	case "research":
		buf.WriteString("# Research\n\n")
		entries, _ := repos.Research.ListByProject(pid)
		for _, e := range entries {
			buf.WriteString(fmt.Sprintf("## %s\n- **ID:** %s\n- **Status:** %s\n- **Confidence:** %.0f\n\n%s\n\n", e.Topic, e.ID, e.Status, e.Confidence, e.Summary))
		}

	case "all":
		buf.WriteString("# Project Documentation\n\n")

		// Plans
		buf.WriteString("## Plans\n\n")
		masters, _ := repos.Plan.ListMastersByProject(pid)
		for _, m := range masters {
			buf.WriteString(fmt.Sprintf("- **%s** (%s) - %s\n", m.Title, m.ID, m.Status))
			if m.Summary != "" {
				buf.WriteString(fmt.Sprintf("  %s\n", m.Summary))
			}
		}

		// Decisions
		buf.WriteString("\n## Decisions\n\n")
		rows, err := ps.DB.Query(`SELECT id, title, decision, status FROM decisions ORDER BY created_at`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id, title, decision, status string
				if err := rows.Scan(&id, &title, &decision, &status); err == nil {
					buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", title, status, decision))
				}
			}
		}

		// Research
		buf.WriteString("\n## Research\n\n")
		entries, _ := repos.Research.ListByProject(pid)
		for _, e := range entries {
			buf.WriteString(fmt.Sprintf("- **%s** (%s, confidence: %.0f)\n", e.Topic, e.Status, e.Confidence))
		}

		// Snapshots
		buf.WriteString("\n## Snapshots\n\n")
		snaps, _ := repos.Snapshot.ListByProject(pid)
		for _, s := range snaps {
			buf.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", s.Reason, s.ID, s.Summary))
		}
	}

	content := buf.String()

	return map[string]any{
		"format":  format,
		"scope":   scope,
		"content": content,
		"chars":   len(content),
	}, nil
}

// ── Phase 51: Product Intent Engine Handlers ──

func HandleCreateProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	description := getStringArg(args, "description")
	if description == "" {
		return nil, fmt.Errorf("description is required")
	}
	expectedOutcome := getStringArg(args, "expected_outcome")
	desiredExperience := getStringArg(args, "desired_experience")
	desiredResult := getStringArg(args, "desired_result")

	var userExpectations, nonExpectations []string
	if ue := getStringArg(args, "user_expectations"); ue != "" {
		userExpectations = strings.Split(ue, "\n")
	}
	if ne := getStringArg(args, "non_expectations"); ne != "" {
		nonExpectations = strings.Split(ne, "\n")
	}

	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.CreateProductIntent(intentv3.CreateProductIntentInput{
		ProjectID:          projectID(projectRoot),
		Description:        description,
		ExpectedOutcome:    expectedOutcome,
		DesiredExperience:  desiredExperience,
		DesiredResult:      desiredResult,
		UserExpectations:   userExpectations,
		NonExpectations:    nonExpectations,
		SuccessDefinition:  getStringArg(args, "success_definition"),
		FailureDefinition:  getStringArg(args, "failure_definition"),
		DiscoveryResultID:  getStringArg(args, "discovery_result_id"),
	})
	if err != nil {
		return nil, fmt.Errorf("create product intent: %w", err)
	}

	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleListProductIntents(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	list, err := svc.ListProductIntents(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list product intents: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, pi := range list {
		items = append(items, map[string]any{
			"id":          pi.ID,
			"description": pi.Description,
			"status":      string(pi.Status),
			"created_at":  formatTime(pi.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}

func HandleGetProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.GetProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("get product intent: %w", err)
	}
	return map[string]any{
		"id":                  pi.ID,
		"project_id":          pi.ProjectID,
		"description":         pi.Description,
		"status":              string(pi.Status),
		"expected_outcome":    pi.ExpectedOutcome,
		"desired_experience":  pi.DesiredExperience,
		"desired_result":      pi.DesiredResult,
		"user_expectations":   pi.UserExpectations,
		"non_expectations":    pi.NonExpectations,
		"success_definition":  pi.SuccessDefinition,
		"failure_definition":  pi.FailureDefinition,
		"discovery_result_id": pi.DiscoveryResultID,
		"created_at":          formatTime(pi.CreatedAt),
		"updated_at":          formatTime(pi.UpdatedAt),
	}, nil
}

func HandleSubmitProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.SubmitProductIntentForApproval(intentID)
	if err != nil {
		return nil, fmt.Errorf("submit product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleApproveProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.ApproveProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("approve product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

func HandleRejectProductIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	intentID := getStringArg(args, "intent_id")
	if intentID == "" {
		return nil, fmt.Errorf("intent_id is required")
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	pi, err := svc.RejectProductIntent(intentID)
	if err != nil {
		return nil, fmt.Errorf("reject product intent: %w", err)
	}
	return map[string]any{
		"id":     pi.ID,
		"status": string(pi.Status),
	}, nil
}

// ── Phase 52: Discovery Engine Handlers ──

func HandleDiscoverIntent(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	content := getStringArg(args, "content")
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	dr, err := svc.DiscoverIntent(projectID(projectRoot), content)
	if err != nil {
		return nil, fmt.Errorf("discover intent: %w", err)
	}
	return map[string]any{
		"id":              dr.ID,
		"project_id":      dr.ProjectID,
		"raw_input":       dr.RawInput,
		"detected_intent": dr.DetectedIntent,
		"classification":  dr.Classification,
		"objectives":      dr.Objectives,
		"restrictions":    dr.Restrictions,
		"preferences":     dr.Preferences,
		"expectations":    dr.Expectations,
		"gaps":            dr.Gaps,
		"questions":       dr.Questions,
		"created_at":      formatTime(dr.CreatedAt),
	}, nil
}

func HandleListDiscoveryResults(args map[string]any) (map[string]any, error) {
	projectRoot, err := getProjectRoot(args)
	if err != nil {
		return nil, fmt.Errorf("resolve project root: %w", err)
	}
	ps, err := openStore(projectRoot)
	if err != nil {
		return nil, err
	}
	defer ps.Close()

	intentRepo := store.NewIntentV3Repository(ps.DB)
	discRepo := store.NewIntentV3DiscoveryResultRepository(ps.DB)
	svc := intentv3.NewService(intentRepo, discRepo)

	list, err := svc.ListDiscoveryResults(projectID(projectRoot))
	if err != nil {
		return nil, fmt.Errorf("list discovery results: %w", err)
	}
	items := make([]map[string]any, 0, len(list))
	for _, dr := range list {
		items = append(items, map[string]any{
			"id":              dr.ID,
			"detected_intent": dr.DetectedIntent,
			"classification":  dr.Classification,
			"created_at":      formatTime(dr.CreatedAt),
		})
	}
	return map[string]any{"items": items, "count": len(items)}, nil
}
