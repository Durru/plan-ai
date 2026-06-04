-- Phase 22 Continuous Planning migration.
-- Runtime source of truth remains the inline migration runner in internal/store/store.go.
-- Migration number 0020 is used because 0018 was already allocated to Phase 18-20 compatibility.

CREATE TABLE IF NOT EXISTS continuous_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  details TEXT NOT NULL DEFAULT '{}',
  source TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS plan_update_proposals (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  affected_plans TEXT NOT NULL DEFAULT '[]',
  affected_tasks TEXT NOT NULL DEFAULT '[]',
  affected_decisions TEXT NOT NULL DEFAULT '[]',
  suggested_updates TEXT NOT NULL DEFAULT '',
  requires_research INTEGER NOT NULL DEFAULT 0,
  requires_approval INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS context_deliveries (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  level TEXT NOT NULL DEFAULT 'L0_Executive',
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_continuous_events_project ON continuous_events(project_id);
CREATE INDEX IF NOT EXISTS idx_continuous_events_type ON continuous_events(event_type);
CREATE INDEX IF NOT EXISTS idx_plan_update_proposals_project ON plan_update_proposals(project_id);
CREATE INDEX IF NOT EXISTS idx_plan_update_proposals_status ON plan_update_proposals(status);
CREATE INDEX IF NOT EXISTS idx_context_deliveries_project ON context_deliveries(project_id);
CREATE INDEX IF NOT EXISTS idx_context_deliveries_level ON context_deliveries(level);
