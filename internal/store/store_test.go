package store

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestEnsureGlobalLayoutCreatesExpectedPaths(t *testing.T) {
	home := t.TempDir()

	layout, err := EnsureGlobalLayout(home)
	if err != nil {
		t.Fatalf("ensure global layout: %v", err)
	}

	assertDir(t, layout.Dir)
	assertFileAbsent(t, layout.DBPath)
	for _, dir := range []string{layout.CacheDir, layout.SkillsDir, layout.LogsDir, layout.DataDir, layout.BackupsDir} {
		assertDir(t, dir)
	}
}

func TestEnsureProjectLayoutCreatesExpectedPaths(t *testing.T) {
	project := t.TempDir()

	layout, err := EnsureProjectLayout(project)
	if err != nil {
		t.Fatalf("ensure project layout: %v", err)
	}

	assertDir(t, layout.Dir)
	assertFileAbsent(t, layout.DBPath)
	for _, dir := range []string{layout.CacheDir, layout.SnapshotsDir, layout.ExportsDir, layout.DocsDir, layout.LocksDir, layout.BackupsDir} {
		assertDir(t, dir)
	}
}

func TestGlobalMigrationsAreIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "global.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := RunGlobalMigrations(db); err != nil {
			t.Fatalf("run global migrations pass %d: %v", i+1, err)
		}
	}

	assertTables(t, db, []string{"schema_migrations", "global_metadata", "global_settings", "known_projects", "global_config", "global_tools", "global_integrations", "global_skills", "global_skill_cache", "global_knowledge", "global_research", "global_templates", "global_model_profiles", "global_logs"})
	assertMigrationCount(t, db, 2)
	assertColumns(t, db, "schema_migrations", []string{"id", "name", "applied_at"})
}

func TestProjectMigrationsAreIdempotent(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "project.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	for i := 0; i < 2; i++ {
		if err := RunProjectMigrations(db); err != nil {
			t.Fatalf("run project migrations pass %d: %v", i+1, err)
		}
	}

	assertTables(t, db, []string{
		"schema_migrations", "project_metadata", "project_settings", "project_state",
		"plans", "phases", "tasks", "decisions", "research_entries", "knowledge_objects", "validations", "snapshots",
		"projects", "raw_inputs", "ingested_sources", "visions", "requirements", "constraints", "decision_history",
		"master_plans", "specific_plans", "implementation_documents", "task_steps", "change_requests", "impact_reports",
		"context_views", "context_chunks", "agent_runs", "subagent_outputs", "file_snapshots", "file_change_events",
		"project_scans", "project_scan_languages", "project_scan_frameworks",
		"project_scan_package_managers", "project_scan_dependencies", "project_scan_files",
		"knowledge_tags", "knowledge_relations", "knowledge_references",
		"research_findings", "research_sources", "research_conclusions",
		"research_tags", "research_knowledge_links",
		"research_jobs", "research_recommendations", "knowledge_links", "workflow_runs",
		"approved_requirements", "approved_constraints", "approved_decisions",
		"approved_preferences", "approved_references", "approved_goals",
		"change_events", "change_reports", "snapshots_v2", "entity_states",
		"mcp_tools", "mcp_runs",
		"opencode_detections", "opencode_integration_state", "opencode_doctor_checks",
		"provider_registry", "skill_registry",
		"agent_runs_v2", "agent_messages", "agent_delegated_jobs",
		"continuous_events", "plan_update_proposals", "context_deliveries", "continuous_status",
		"vision_discovery_sessions", "vision_assumptions", "vision_ambiguities", "vision_approvals",
		"master_plan_versions", "master_plan_changes", "master_plan_approvals", "plan_evolution_events",
		"specific_plan_versions", "specific_plan_research_links", "specific_plan_regenerations",
		"context_delivery_sessions", "context_delivery_usage", "context_delivery_budgets",
		"intent_profiles",
		"vision_documents", "approval_records", "requirement_candidates",
		"context_packages_v2", "research_orchestration_runs", "project_references_v2",
		"plan_evolution_blueprints_v3", "implementation_packages_v2",
		"change_impact_reports_v2", "continuous_regenerations_v2",
		"subagent_tasks_v2", "opencode_workflows_v2",
	})
	assertMigrationCount(t, db, 39)
	assertColumns(t, db, "schema_migrations", []string{"id", "name", "applied_at"})
	assertColumns(t, db, "plans", []string{"id", "type", "title", "summary", "status", "version", "parent_plan_id", "created_at", "updated_at"})
	assertColumns(t, db, "phases", []string{"id", "plan_id", "title", "summary", "status", "position", "created_at", "updated_at"})
	assertColumns(t, db, "tasks", []string{"id", "phase_id", "plan_id", "title", "summary", "status", "position", "context_size", "created_at", "updated_at"})
	assertColumns(t, db, "validations", []string{"id", "target_type", "target_id", "status", "summary", "created_at", "updated_at"})
	assertColumns(t, db, "snapshots", []string{"id", "reason", "summary", "created_at"})
	assertColumns(t, db, "project_scans", []string{"id", "project_root", "git_detected", "git_branch", "fingerprint", "summary", "created_at"})
	assertColumns(t, db, "project_scan_languages", []string{"id", "scan_id", "language", "files_count", "created_at"})
	assertColumns(t, db, "project_scan_files", []string{"id", "scan_id", "path", "kind", "size_bytes", "created_at"})
	assertColumns(t, db, "research_entries", []string{"id", "topic", "source", "summary", "conclusion", "confidence", "category", "status", "created_at", "updated_at"})
	assertColumns(t, db, "research_findings", []string{"id", "research_id", "title", "content", "importance", "created_at"})
	assertColumns(t, db, "research_sources", []string{"id", "research_id", "title", "url", "source_type", "created_at"})
	assertColumns(t, db, "research_conclusions", []string{"id", "research_id", "content", "confidence", "created_at"})
	assertColumns(t, db, "research_tags", []string{"id", "research_id", "tag"})
	assertColumns(t, db, "research_knowledge_links", []string{"id", "research_id", "knowledge_id", "created_at"})
	assertColumns(t, db, "knowledge_objects", []string{"id", "project_id", "title", "topic", "category", "summary", "content", "confidence", "source_type", "reuse_count", "status", "research_ids", "related_decisions", "related_requirements", "related_constraints", "created_at", "updated_at"})
	assertColumns(t, db, "research_jobs", []string{"id", "project_id", "topic", "summary", "confidence", "status", "created_at"})
	assertColumns(t, db, "research_recommendations", []string{"id", "research_id", "content", "created_at"})
	assertColumns(t, db, "knowledge_links", []string{"id", "knowledge_id", "link_type", "target_id", "created_at"})
	assertColumns(t, db, "workflow_runs", []string{"id", "workflow_type", "status", "started_at", "finished_at"})
	assertColumns(t, db, "knowledge_tags", []string{"id", "knowledge_id", "tag", "created_at"})
	assertColumns(t, db, "knowledge_relations", []string{"id", "source_id", "target_id", "relation_type", "created_at"})
	assertColumns(t, db, "knowledge_references", []string{"id", "knowledge_id", "reference_type", "reference_id", "created_at"})
	assertColumns(t, db, "raw_inputs", []string{"id", "project_id", "source_type", "content", "raw_content", "metadata", "created_at", "updated_at"})
	assertColumns(t, db, "ingested_sources", []string{"id", "project_id", "raw_input_id", "source_type", "normalized_content", "classification", "metadata", "created_at", "updated_at"})
	assertColumns(t, db, "visions", []string{"id", "project_id", "title", "summary", "target_users", "expected_outcome", "functional_goals", "ux_goals", "business_goals", "constraints", "assumptions", "missing_information", "visual_references", "success_criteria", "approved", "created_at", "updated_at"})
	assertColumns(t, db, "approved_requirements", []string{"id", "project_id", "source_id", "content", "state", "created_at", "updated_at"})
	assertColumns(t, db, "delegated_jobs", []string{"id", "project_id", "intent", "capability", "workflow_type", "job_type", "status", "result_summary", "created_at", "completed_at"})
	assertColumns(t, db, "agent_responses", []string{"id", "run_id", "content", "status", "created_at", "updated_at"})
	assertColumns(t, db, "continuous_status", []string{"id", "project_id", "active_plan", "active_phase", "next_task", "blocked_items", "approvals_needed", "outdated_plans", "created_at", "updated_at"})
	assertColumns(t, db, "context_delivery_logs", []string{"id", "project_id", "level", "content", "created_at"})
}

func assertDir(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat dir %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s should be a directory", path)
	}
}

func assertFileAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s should not exist before opening DB, err=%v", path, err)
	}
}

func assertTables(t *testing.T, db *sql.DB, names []string) {
	t.Helper()
	for _, name := range names {
		var got string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&got)
		if err != nil {
			t.Fatalf("table %s missing: %v", name, err)
		}
	}
}

func assertMigrationCount(t *testing.T, db *sql.DB, want int) {
	t.Helper()
	var got int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&got); err != nil {
		t.Fatalf("count migrations: %v", err)
	}
	if got != want {
		t.Fatalf("migration count = %d, want %d", got, want)
	}
}

func assertColumns(t *testing.T, db *sql.DB, table string, names []string) {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("table info %s: %v", table, err)
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("scan column: %v", err)
		}
		seen[name] = true
	}
	for _, name := range names {
		if !seen[name] {
			t.Fatalf("column %s.%s missing", table, name)
		}
	}
}
