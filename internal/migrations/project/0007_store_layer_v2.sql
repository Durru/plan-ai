-- Phase 8 Store Layer v2 project migration.
-- Runtime source of truth currently remains the inline migration runner in internal/store/store.go.
-- This file mirrors the schema for review/documentation and future extraction.

CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY, name TEXT NOT NULL, root_path TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS raw_inputs (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_type TEXT NOT NULL DEFAULT 'manual', content TEXT NOT NULL, metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS ingested_sources (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, raw_input_id TEXT, uri TEXT NOT NULL DEFAULT '', source_type TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS visions (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, expected_outcome TEXT NOT NULL DEFAULT '', approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS requirements (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, type TEXT NOT NULL, statement TEXT NOT NULL, approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS constraints (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, type TEXT NOT NULL, description TEXT NOT NULL, approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS decision_history (id TEXT PRIMARY KEY, decision_id TEXT NOT NULL, from_status TEXT NOT NULL DEFAULT '', to_status TEXT NOT NULL, note TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS master_plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS specific_plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, master_plan_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS implementation_documents (id TEXT PRIMARY KEY, project_id TEXT NOT NULL DEFAULT '', specific_plan_id TEXT NOT NULL, title TEXT NOT NULL, content TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS task_steps (id TEXT PRIMARY KEY, task_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending', position INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS change_requests (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, reason TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL, requester TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS impact_reports (id TEXT PRIMARY KEY, change_request_id TEXT NOT NULL, affected_plans TEXT NOT NULL DEFAULT '[]', affected_phases TEXT NOT NULL DEFAULT '[]', affected_tasks TEXT NOT NULL DEFAULT '[]', affected_decisions TEXT NOT NULL DEFAULT '[]', affected_entities TEXT NOT NULL DEFAULT '[]', summary TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS context_views (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, name TEXT NOT NULL, content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS context_chunks (id TEXT PRIMARY KEY, context_view_id TEXT NOT NULL, chunk_index INTEGER NOT NULL DEFAULT 0, content TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS agent_runs (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, agent_name TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS subagent_outputs (id TEXT PRIMARY KEY, agent_run_id TEXT NOT NULL, subagent_name TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS file_snapshots (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, path TEXT NOT NULL, hash TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS file_change_events (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, path TEXT NOT NULL, event_type TEXT NOT NULL, created_at TEXT NOT NULL);

-- Additive compatibility columns for legacy Phase 0-6 tables. These statements are run by the migration runner once via schema_migrations.
ALTER TABLE decisions ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE decisions ADD COLUMN rationale TEXT NOT NULL DEFAULT '';
ALTER TABLE decisions ADD COLUMN alternatives TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN objective TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN date TEXT NOT NULL DEFAULT '';
ALTER TABLE knowledge_objects ADD COLUMN type TEXT NOT NULL DEFAULT 'reference';
ALTER TABLE phases ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE validations ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE validations ADD COLUMN type TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE validations ADD COLUMN details TEXT NOT NULL DEFAULT '';
ALTER TABLE snapshots ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE snapshots ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_raw_inputs_project ON raw_inputs(project_id);
CREATE INDEX IF NOT EXISTS idx_visions_project ON visions(project_id);
CREATE INDEX IF NOT EXISTS idx_requirements_project ON requirements(project_id);
CREATE INDEX IF NOT EXISTS idx_constraints_project ON constraints(project_id);
CREATE INDEX IF NOT EXISTS idx_decisions_project ON decisions(project_id);
CREATE INDEX IF NOT EXISTS idx_research_entries_project ON research_entries(project_id);
CREATE INDEX IF NOT EXISTS idx_master_plans_project ON master_plans(project_id);
CREATE INDEX IF NOT EXISTS idx_specific_plans_master ON specific_plans(master_plan_id);
CREATE INDEX IF NOT EXISTS idx_phases_plan ON phases(plan_id);
CREATE INDEX IF NOT EXISTS idx_tasks_phase ON tasks(phase_id);
CREATE INDEX IF NOT EXISTS idx_validations_project ON validations(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_project ON snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_change_requests_project ON change_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_impact_reports_change ON impact_reports(change_request_id);

CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_objects_fts USING fts5(id UNINDEXED, topic, summary, content);
CREATE VIRTUAL TABLE IF NOT EXISTS research_entries_fts USING fts5(id UNINDEXED, topic, objective, summary);
CREATE VIRTUAL TABLE IF NOT EXISTS implementation_documents_fts USING fts5(id UNINDEXED, title, content);
CREATE VIRTUAL TABLE IF NOT EXISTS raw_inputs_fts USING fts5(id UNINDEXED, content);
