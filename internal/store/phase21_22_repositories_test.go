package store

import (
	"testing"
)

// ──────────────────────────────────────────────
// Phase 21: Agent System Repositories
// ──────────────────────────────────────────────

func TestAgentRunV2RepositoryCreateGetUpdateList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewAgentRunV2Repository(db)

	// Create
	rec, err := r.CreateRun(AgentRunV2Record{
		ID:        "ar:v2:1",
		ProjectID: "project:test",
		Intent:    "create_plan",
		Status:    "processed",
		Response:  `{"plan_id": "plan:1"}`,
	})
	if err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	if rec.ID != "ar:v2:1" {
		t.Errorf("id = %q", rec.ID)
	}
	if rec.Intent != "create_plan" {
		t.Errorf("intent = %q", rec.Intent)
	}
	if rec.CreatedAt == "" {
		t.Error("created_at is empty")
	}

	// Get
	got, err := r.GetRun("ar:v2:1")
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if got.Status != "processed" {
		t.Errorf("status = %q", got.Status)
	}

	// UpdateRunStatus
	if err := r.UpdateRunStatus("ar:v2:1", "completed", `{"done": true}`); err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}
	got, _ = r.GetRun("ar:v2:1")
	if got.Status != "completed" {
		t.Errorf("status after update = %q", got.Status)
	}

	// ListRuns
	records, err := r.ListRuns("project:test", 10)
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}
}

func TestAgentRunV2RepositoryCreateMessageListMessages(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewAgentRunV2Repository(db)

	// Need a run first
	r.CreateRun(AgentRunV2Record{
		ID: "ar:msg:1", ProjectID: "project:test",
		Intent: "research", Status: "processed",
	})

	// Create message
	msg, err := r.CreateMessage(AgentMessageRecord{
		ID:      "amsg:1",
		RunID:   "ar:msg:1",
		Role:    "user",
		Content: "Research the topic",
	})
	if err != nil {
		t.Fatalf("CreateMessage: %v", err)
	}
	if msg.ID != "amsg:1" {
		t.Errorf("id = %q", msg.ID)
	}
	if msg.Role != "user" {
		t.Errorf("role = %q", msg.Role)
	}
	if msg.CreatedAt == "" {
		t.Error("created_at is empty")
	}

	// ListMessages
	msgs, err := r.ListMessages("ar:msg:1")
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("len = %d, want 1", len(msgs))
	}
	if msgs[0].Content != "Research the topic" {
		t.Errorf("content = %q", msgs[0].Content)
	}
}

func TestDelegatedJobRepositoryCreateGetListUpdate(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewDelegatedJobRepository(db)

	job, err := r.CreateJob(AgentDelegatedJobRecord{
		ID:            "dj:1",
		ProjectID:     "project:test",
		Intent:        "research",
		Capability:    "research",
		WorkflowType:  "research_workflow",
		JobType:       "research_task",
		Status:        "pending",
		ResultSummary: "",
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	if job.ID != "dj:1" {
		t.Errorf("id = %q", job.ID)
	}
	if job.Status != "pending" {
		t.Errorf("status = %q", job.Status)
	}

	// GetJob
	got, err := r.GetJob("dj:1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Intent != "research" {
		t.Errorf("intent = %q", got.Intent)
	}

	// ListJobs
	jobs, err := r.ListJobs("project:test")
	if err != nil {
		t.Fatalf("ListJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("len = %d, want 1", len(jobs))
	}

	// UpdateJob
	if err := r.UpdateJob("dj:1", "completed", "Research complete"); err != nil {
		t.Fatalf("UpdateJob: %v", err)
	}
	got, _ = r.GetJob("dj:1")
	if got.Status != "completed" {
		t.Errorf("status after update = %q", got.Status)
	}
	if got.ResultSummary != "Research complete" {
		t.Errorf("result_summary = %q", got.ResultSummary)
	}
}

// ──────────────────────────────────────────────
// Phase 22: Continuous Planning Repositories
// ──────────────────────────────────────────────

func TestContinuousEventRepositoryCreateList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewContinuousEventRepository(db)

	ev, err := r.CreateEvent(ContinuousEventRecord{
		ID:        "cev:1",
		ProjectID: "project:test",
		EventType: "file_changed",
		Summary:   "Source file modified",
		Details:   `{"path": "/src/main.go"}`,
		Source:    "watcher",
	})
	if err != nil {
		t.Fatalf("CreateEvent: %v", err)
	}
	if ev.ID != "cev:1" {
		t.Errorf("id = %q", ev.ID)
	}
	if ev.EventType != "file_changed" {
		t.Errorf("event_type = %q", ev.EventType)
	}

	events, err := r.ListEvents("project:test", 10)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len = %d, want 1", len(events))
	}
	if events[0].Summary != "Source file modified" {
		t.Errorf("summary = %q", events[0].Summary)
	}
}

func TestContinuousEventRepositoryListLimits(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewContinuousEventRepository(db)

	for i := 0; i < 5; i++ {
		r.CreateEvent(ContinuousEventRecord{
			ID: "cev:l:" + itoa(i), ProjectID: "project:test",
			EventType: "test", Summary: "Event " + itoa(i),
		})
	}

	// Default limit = 50 returns all 5
	events, err := r.ListEvents("project:test", 0)
	if err != nil {
		t.Fatalf("ListEvents default: %v", err)
	}
	if len(events) != 5 {
		t.Errorf("default limit len = %d, want 5", len(events))
	}

	// Explicit limit 2
	events, err = r.ListEvents("project:test", 2)
	if err != nil {
		t.Fatalf("ListEvents limit 2: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("limit 2 len = %d, want 2", len(events))
	}
}

func TestPlanUpdateProposalRepositoryCreateGetListUpdateStatus(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewPlanUpdateProposalRepository(db)

	p, err := r.CreateProposal(PlanUpdateProposalRecord{
		ID:                "pup:1",
		ProjectID:         "project:test",
		Reason:            "Requirement changed",
		AffectedPlans:     `["plan:1"]`,
		AffectedTasks:     `["task:1"]`,
		AffectedDecisions: `["dec:1"]`,
		SuggestedUpdates:  "Update plan scope",
		RequiresResearch:  0,
		RequiresApproval:  1,
		Status:            "draft",
	})
	if err != nil {
		t.Fatalf("CreateProposal: %v", err)
	}
	if p.ID != "pup:1" {
		t.Errorf("id = %q", p.ID)
	}
	if p.Status != "draft" {
		t.Errorf("status = %q", p.Status)
	}

	// GetProposal
	got, err := r.GetProposal("pup:1")
	if err != nil {
		t.Fatalf("GetProposal: %v", err)
	}
	if got.Reason != "Requirement changed" {
		t.Errorf("reason = %q", got.Reason)
	}

	// ListProposals
	proposals, err := r.ListProposals("project:test")
	if err != nil {
		t.Fatalf("ListProposals: %v", err)
	}
	if len(proposals) != 1 {
		t.Fatalf("len = %d, want 1", len(proposals))
	}

	// UpdateProposalStatus
	if err := r.UpdateProposalStatus("pup:1", "approved"); err != nil {
		t.Fatalf("UpdateProposalStatus: %v", err)
	}
	got, _ = r.GetProposal("pup:1")
	if got.Status != "approved" {
		t.Errorf("status after update = %q", got.Status)
	}
}

func TestContextDeliveryRepositoryCreateList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewContextDeliveryRepository(db)

	d, err := r.CreateDelivery(ContextDeliveryRecord{
		ID:        "cd:1",
		ProjectID: "project:test",
		Level:     "L1_Planning",
		Content:   "All plans are up to date",
	})
	if err != nil {
		t.Fatalf("CreateDelivery: %v", err)
	}
	if d.ID != "cd:1" {
		t.Errorf("id = %q", d.ID)
	}
	if d.Level != "L1_Planning" {
		t.Errorf("level = %q", d.Level)
	}

	// ListDeliveries by level
	deliveries, err := r.ListDeliveries("project:test", "L1_Planning", 10)
	if err != nil {
		t.Fatalf("ListDeliveries: %v", err)
	}
	if len(deliveries) != 1 {
		t.Fatalf("len = %d, want 1", len(deliveries))
	}
	if deliveries[0].Content != "All plans are up to date" {
		t.Errorf("content = %q", deliveries[0].Content)
	}

	// ListDeliveries for non-existent level returns empty
	deliveries, err = r.ListDeliveries("project:test", "L5_Nonexistent", 10)
	if err != nil {
		t.Fatalf("ListDeliveries empty: %v", err)
	}
	if len(deliveries) != 0 {
		t.Errorf("len = %d, want 0", len(deliveries))
	}
}

func TestContextDeliveryRepositoryListLimits(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewContextDeliveryRepository(db)

	for i := 0; i < 5; i++ {
		r.CreateDelivery(ContextDeliveryRecord{
			ID: "cd:l:" + itoa(i), ProjectID: "project:test",
			Level: "L0_Executive", Content: "Status " + itoa(i),
		})
	}

	// Default limit = 50 returns all 5
	deliveries, err := r.ListDeliveries("project:test", "L0_Executive", 0)
	if err != nil {
		t.Fatalf("ListDeliveries default: %v", err)
	}
	if len(deliveries) != 5 {
		t.Errorf("default limit len = %d, want 5", len(deliveries))
	}

	// Explicit limit 2
	deliveries, err = r.ListDeliveries("project:test", "L0_Executive", 2)
	if err != nil {
		t.Fatalf("ListDeliveries limit 2: %v", err)
	}
	if len(deliveries) != 2 {
		t.Errorf("limit 2 len = %d, want 2", len(deliveries))
	}
}

func TestPhase21And22CompatibilityNames(t *testing.T) {
	db := openStoreTestDB(t)

	runs := NewAgentRunV2Repository(db)
	if _, err := runs.CreateRun(AgentRunV2Record{ID: "ar:compat:1", ProjectID: "project:test", Intent: "project_status", Status: "processed", Response: `{"ok":true}`}); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	jobs := NewDelegatedJobRepository(db)
	if _, err := jobs.CreateJob(AgentDelegatedJobRecord{ID: "dj:compat:1", ProjectID: "project:test", Intent: "research_topic", Capability: "research", WorkflowType: "research", JobType: "research_job", Status: "pending"}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}
	deliveries := NewContextDeliveryRepository(db)
	if _, err := deliveries.CreateDelivery(ContextDeliveryRecord{ID: "cd:compat:1", ProjectID: "project:test", Level: "L0_Executive", Content: "status"}); err != nil {
		t.Fatalf("CreateDelivery: %v", err)
	}

	checks := map[string]string{
		"agent_runs":            "SELECT COUNT(*) FROM agent_runs WHERE id = 'ar:compat:1'",
		"delegated_jobs":        "SELECT COUNT(*) FROM delegated_jobs WHERE id = 'dj:compat:1'",
		"agent_responses":       "SELECT COUNT(*) FROM agent_responses WHERE id = 'ar:compat:1'",
		"context_delivery_logs": "SELECT COUNT(*) FROM context_delivery_logs WHERE id = 'cd:compat:1'",
	}
	for name, query := range checks {
		t.Run(name, func(t *testing.T) {
			var count int
			if err := db.QueryRow(query).Scan(&count); err != nil {
				t.Fatalf("query %s: %v", name, err)
			}
			if count != 1 {
				t.Fatalf("%s count = %d, want 1", name, count)
			}
		})
	}

	if _, err := db.Exec(`INSERT INTO continuous_status (id, project_id, active_plan, active_phase, next_task, created_at, updated_at) VALUES ('cs:1', 'project:test', 'plan:1', 'phase:1', 'task:1', 'now', 'now')`); err != nil {
		t.Fatalf("insert continuous_status: %v", err)
	}
	var nextTask string
	if err := db.QueryRow(`SELECT next_task FROM continuous_status WHERE id = 'cs:1'`).Scan(&nextTask); err != nil {
		t.Fatalf("select continuous_status: %v", err)
	}
	if nextTask != "task:1" {
		t.Fatalf("next_task = %q, want task:1", nextTask)
	}
}
