package store

import (
	"database/sql"
	"testing"
	"time"
)

// ──────────────────────────────────────────────
// Phase 18: Change Engine
// ──────────────────────────────────────────────

func TestChangeEventRepositoryCreateGetListUpdate(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewChangeEventRepository(db)

	ev, err := r.Create(ChangeEventRecord{
		ID:          "ce:1",
		ProjectID:   "project:test",
		ChangeType:  "requirement",
		Summary:     "New requirement added",
		Description: "Added a new compliance requirement",
		Severity:    "medium",
		Status:      "pending",
		Source:      "user",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if ev.ID != "ce:1" {
		t.Errorf("id = %q", ev.ID)
	}
	if ev.ProjectID != "project:test" {
		t.Errorf("project_id = %q", ev.ProjectID)
	}

	// Get
	got, err := r.Get("ce:1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Summary != "New requirement added" {
		t.Errorf("summary = %q", got.Summary)
	}
	if got.CreatedAt == "" {
		t.Error("created_at is empty")
	}

	// ListByProject
	records, err := r.ListByProject("project:test", 10)
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("len = %d, want 1", len(records))
	}

	// UpdateStatus
	if err := r.UpdateStatus("ce:1", "approved"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}
	got, _ = r.Get("ce:1")
	if got.Status != "approved" {
		t.Errorf("status after update = %q", got.Status)
	}
}

func TestChangeEventRepositoryListLimits(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewChangeEventRepository(db)

	for i := 0; i < 5; i++ {
		id := domainID("ce", i)
		r.Create(ChangeEventRecord{
			ID: id, ProjectID: "project:test", ChangeType: "test",
			Summary: "Event " + id, Severity: "low", Status: "pending",
		})
	}

	// Default limit = 50 returns all 5
	records, err := r.ListByProject("project:test", 0)
	if err != nil {
		t.Fatalf("ListByProject default: %v", err)
	}
	if len(records) != 5 {
		t.Errorf("default limit len = %d, want 5", len(records))
	}

	// Explicit limit 2
	records, err = r.ListByProject("project:test", 2)
	if err != nil {
		t.Fatalf("ListByProject limit 2: %v", err)
	}
	if len(records) != 2 {
		t.Errorf("limit 2 len = %d, want 2", len(records))
	}
}

func TestChangeReportRepositoryCreateGetByEvent(t *testing.T) {
	db := openStoreTestDB(t)
	evRepo := NewChangeEventRepository(db)
	reportRepo := NewChangeReportRepository(db)

	// Need a change event first
	evRepo.Create(ChangeEventRecord{
		ID: "ce:report-test", ProjectID: "project:test",
		ChangeType: "feature", Summary: "Test report", Severity: "high", Status: "pending",
	})

	report, err := reportRepo.Create(ChangeReportRecord{
		ID:               "cr:1",
		ChangeEventID:    "ce:report-test",
		ProjectID:        "project:test",
		Analysis:         `{"risk": "low"}`,
		AffectedEntities: `["plan:1", "task:1"]`,
		ReviewRequired:   1,
		Summary:          "Low risk change",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if report.ID != "cr:1" {
		t.Errorf("id = %q", report.ID)
	}

	got, err := reportRepo.GetByEvent("ce:report-test")
	if err != nil {
		t.Fatalf("GetByEvent: %v", err)
	}
	if got.Summary != "Low risk change" {
		t.Errorf("summary = %q", got.Summary)
	}
	if got.ReviewRequired != 1 {
		t.Errorf("review_required = %d", got.ReviewRequired)
	}
}

func TestSnapshotV2RepositoryCreateList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewSnapshotV2Repository(db)

	snap, err := r.Create(SnapshotV2Record{
		ID:             "snap:v2:1",
		ProjectID:      "project:test",
		Reason:         "pre-approval checkpoint",
		EntitySnapshot: `{"plans": 3, "tasks": 5}`,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if snap.ID != "snap:v2:1" {
		t.Errorf("id = %q", snap.ID)
	}

	snaps, err := r.ListByProject("project:test", 10)
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(snaps) != 1 {
		t.Fatalf("len = %d, want 1", len(snaps))
	}
	if snaps[0].Reason != "pre-approval checkpoint" {
		t.Errorf("reason = %q", snaps[0].Reason)
	}
}

func TestEntityStateRepositoryUpsertGetList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewEntityStateRepository(db)

	// Upsert
	err := r.Upsert(EntityStateRecord{
		ID: "es:1", ProjectID: "project:test",
		EntityType: "plan", EntityID: "plan:1",
		Status: "valid", LastChangeID: "", Reason: "",
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// Get
	got, err := r.Get("plan", "plan:1", "project:test")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Status != "valid" {
		t.Errorf("status = %q", got.Status)
	}

	// Upsert update
	err = r.Upsert(EntityStateRecord{
		ID: "es:1", ProjectID: "project:test",
		EntityType: "plan", EntityID: "plan:1",
		Status: "invalid", LastChangeID: "ce:1", Reason: "changed",
	})
	if err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	got, _ = r.Get("plan", "plan:1", "project:test")
	if got.Status != "invalid" {
		t.Errorf("status after update = %q", got.Status)
	}

	// ListByProject
	states, err := r.ListByProject("project:test")
	if err != nil {
		t.Fatalf("ListByProject: %v", err)
	}
	if len(states) != 1 {
		t.Fatalf("len = %d, want 1", len(states))
	}
}

// ──────────────────────────────────────────────
// Phase 19: MCP Server
// ──────────────────────────────────────────────

func TestMCPToolRepositoryUpsertList(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewMCPToolRepository(db)

	err := r.Upsert(MCPToolRecord{
		ID: "tool:1", Name: "analyze", Description: "Analyze code",
		SchemaDef: `{"type": "object"}`, Enabled: 1,
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// Upsert again (update)
	err = r.Upsert(MCPToolRecord{
		ID: "tool:1", Name: "analyze", Description: "Analyze code v2",
		SchemaDef: `{"type": "object", "properties": {}}`, Enabled: 1,
	})
	if err != nil {
		t.Fatalf("Upsert update: %v", err)
	}

	tools, err := r.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("len = %d, want 1", len(tools))
	}
	if tools[0].Description != "Analyze code v2" {
		t.Errorf("description = %q", tools[0].Description)
	}

	// Disabled tool should not appear in List
	err = r.Upsert(MCPToolRecord{
		ID: "tool:2", Name: "deprecated", Description: "Old tool",
		SchemaDef: `{}`, Enabled: 0,
	})
	if err != nil {
		t.Fatalf("Upsert disabled: %v", err)
	}
	tools, _ = r.List()
	if len(tools) != 1 {
		t.Errorf("disabled tool listed: len = %d, want 1", len(tools))
	}
}

func TestMCPRunRepositoryCreateListByTool(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewMCPRunRepository(db)

	err := r.Create(MCPRunRecord{
		ID: "run:1", ToolName: "analyze",
		Arguments: `{"path": "/src"}`, Success: 1, ErrorMessage: "",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = r.Create(MCPRunRecord{
		ID: "run:2", ToolName: "analyze",
		Arguments: `{"path": "/lib"}`, Success: 0,
		ErrorMessage: "file not found",
	})
	if err != nil {
		t.Fatalf("Create run:2: %v", err)
	}

	runs, err := r.ListByTool("analyze", 10)
	if err != nil {
		t.Fatalf("ListByTool: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("len = %d, want 2", len(runs))
	}
	if runs[0].ToolName != "analyze" {
		t.Errorf("tool_name = %q", runs[0].ToolName)
	}
}

// ──────────────────────────────────────────────
// Phase 20: OpenCode Integration
// ──────────────────────────────────────────────

func TestOpenCodeDetectionRepositoryCreateLatest(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewOpenCodeDetectionRepository(db)

	err := r.Create(OpenCodeDetectionRecord{
		ID: "ocd:1", ProjectRoot: "/tmp/test-project",
		Found: 1, ConfigPath: "/tmp/test-project/.opencode",
		IsInitialized: 1, AgentName: "agent-v1", SkillCount: 3,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	latest, err := r.Latest("/tmp/test-project")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest.ID != "ocd:1" {
		t.Errorf("id = %q", latest.ID)
	}
	if latest.AgentName != "agent-v1" {
		t.Errorf("agent_name = %q", latest.AgentName)
	}
	if latest.SkillCount != 3 {
		t.Errorf("skill_count = %d", latest.SkillCount)
	}
}

func TestOpenCodeDetectionRepositoryLatestReturnsEmptyForNone(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewOpenCodeDetectionRepository(db)

	_, err := r.Latest("/nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestOpenCodeIntegrationRepositoryUpsertGet(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewOpenCodeIntegrationRepository(db)

	err := r.Upsert(OpenCodeIntegrationStateRecord{
		ID: "oci:1", ProjectRoot: "/tmp/test-project",
		Mode: "integrated", Enabled: 1, ReadOnly: 1,
		LastDetectedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	state, err := r.Get("/tmp/test-project")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if state.Mode != "integrated" {
		t.Errorf("mode = %q", state.Mode)
	}
	if state.Enabled != 1 {
		t.Errorf("enabled = %d", state.Enabled)
	}

	// Upsert update
	err = r.Upsert(OpenCodeIntegrationStateRecord{
		ID: "oci:1", ProjectRoot: "/tmp/test-project",
		Mode: "standalone", Enabled: 0, ReadOnly: 1,
		LastDetectedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("Upsert update: %v", err)
	}
	state, _ = r.Get("/tmp/test-project")
	if state.Mode != "standalone" {
		t.Errorf("mode after update = %q", state.Mode)
	}
	if state.Enabled != 0 {
		t.Errorf("enabled after update = %d", state.Enabled)
	}
}

func TestOpenCodeDoctorRepositoryCreateListByProject(t *testing.T) {
	db := openStoreTestDB(t)
	r := NewOpenCodeDoctorRepository(db)

	err := r.Create(OpenCodeDoctorCheckRecord{
		ID: "ocdr:1", ProjectRoot: "/tmp/test-project",
		CheckName: "config_exists", Status: "pass", Message: "config found",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = r.Create(OpenCodeDoctorCheckRecord{
		ID: "ocdr:2", ProjectRoot: "/tmp/test-project",
		CheckName: "opencode_version", Status: "fail",
		Message: "version too old",
	})
	if err != nil {
		t.Fatalf("Create check 2: %v", err)
	}

	checks, err := r.LatestByProject("/tmp/test-project")
	if err != nil {
		t.Fatalf("LatestByProject: %v", err)
	}
	if len(checks) != 2 {
		t.Fatalf("len = %d, want 2", len(checks))
	}
}

// ──────────────────────────────────────────────
// Compatibility tables / views (0018 migration)
// ──────────────────────────────────────────────

func TestCompatibilityTablesExist(t *testing.T) {
	db := openStoreTestDB(t)

	checkName := func(kind, name string) error {
		var found int
		return db.QueryRow(
			`SELECT COUNT(*) FROM sqlite_master WHERE type=? AND name=?`,
			kind, name,
		).Scan(&found)
	}

	// These tables exist from earlier migrations
	for _, name := range []string{"snapshots", "change_events", "impact_reports", "opencode_detections"} {
		if err := checkName("table", name); err != nil {
			t.Errorf("table %s check failed: %v", name, err)
		}
	}

	// These views come from migration 0018
	for _, name := range []string{"tool_runs", "tool_audit"} {
		if err := checkName("view", name); err != nil {
			t.Errorf("view %s check failed: %v", name, err)
		}
	}

	// These tables come from migration 0018
	for _, name := range []string{"provider_registry", "skill_registry"} {
		if err := checkName("table", name); err != nil {
			t.Errorf("table %s check failed: %v", name, err)
		}
	}
}

func TestToolRunsViewReflectsMcpRuns(t *testing.T) {
	db := openStoreTestDB(t)
	mcpRunRepo := NewMCPRunRepository(db)

	// Insert into mcp_runs
	mcpRunRepo.Create(MCPRunRecord{
		ID: "run:view:1", ToolName: "plan-tool",
		Arguments: `{"task": "test"}`, Success: 1,
	})
	mcpRunRepo.Create(MCPRunRecord{
		ID: "run:view:2", ToolName: "plan-tool",
		Arguments: `{"task": "fail"}`, Success: 0,
		ErrorMessage: "timeout",
	})

	// Query through tool_runs view
	rows, err := db.Query(`SELECT id, tool_name, success FROM tool_runs ORDER BY id`)
	if err != nil {
		t.Fatalf("query tool_runs: %v", err)
	}
	defer rows.Close()
	var count int
	for rows.Next() {
		var id, toolName string
		var success int
		if err := rows.Scan(&id, &toolName, &success); err != nil {
			t.Fatalf("scan row: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("tool_runs returned %d rows, want 2", count)
	}
}

func TestToolAuditViewQueryable(t *testing.T) {
	db := openStoreTestDB(t)
	toolRepo := NewMCPToolRepository(db)
	runRepo := NewMCPRunRepository(db)

	// Insert a tool and a run
	toolRepo.Upsert(MCPToolRecord{
		ID: "ta:tool:1", Name: "audit-tool",
		Description: "Tool under audit", SchemaDef: `{}`, Enabled: 1,
	})
	runRepo.Create(MCPRunRecord{
		ID: "ta:run:1", ToolName: "audit-tool",
		Arguments: `{"a": 1}`, Success: 1,
	})

	// Query tool_audit view
	var runID, toolName string
	var success int
	err := db.QueryRow(`SELECT run_id, tool_name, success FROM tool_audit WHERE tool_name = ?`, "audit-tool").
		Scan(&runID, &toolName, &success)
	if err != nil {
		t.Fatalf("query tool_audit: %v", err)
	}
	if runID != "ta:run:1" {
		t.Errorf("run_id = %q", runID)
	}
	if success != 1 {
		t.Errorf("success = %d", success)
	}
}

func TestProviderRegistryInsertAndQuery(t *testing.T) {
	db := openStoreTestDB(t)

	_, err := db.Exec(`INSERT INTO provider_registry (id, name, provider_type, endpoint, config, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"pr:1", "anthropic", "api", "https://api.anthropic.com", `{"key": "sk-..."}`, 1,
		time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert provider_registry: %v", err)
	}

	var name, ptype string
	err = db.QueryRow(`SELECT name, provider_type FROM provider_registry WHERE id = ?`, "pr:1").
		Scan(&name, &ptype)
	if err != nil {
		t.Fatalf("query provider_registry: %v", err)
	}
	if name != "anthropic" {
		t.Errorf("name = %q", name)
	}
	if ptype != "api" {
		t.Errorf("provider_type = %q", ptype)
	}
}

func TestSkillRegistryInsertAndQuery(t *testing.T) {
	db := openStoreTestDB(t)

	_, err := db.Exec(`INSERT INTO skill_registry (id, name, source, version, description, checksum, enabled, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"sr:1", "code-review", "builtin", "1.0.0", "Code review skill", "abc123", 1,
		time.Now().UTC().Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert skill_registry: %v", err)
	}

	var name, version string
	var enabled int
	err = db.QueryRow(`SELECT name, version, enabled FROM skill_registry WHERE id = ?`, "sr:1").
		Scan(&name, &version, &enabled)
	if err != nil {
		t.Fatalf("query skill_registry: %v", err)
	}
	if name != "code-review" {
		t.Errorf("name = %q", name)
	}
	if version != "1.0.0" {
		t.Errorf("version = %q", version)
	}
	if enabled != 1 {
		t.Errorf("enabled = %d", enabled)
	}
}

// ──────────────────────────────────────────────
// Helper
// ──────────────────────────────────────────────

func domainID(prefix string, n int) string {
	return prefix + ":" + itoa(n)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
