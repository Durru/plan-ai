-- Phase 18: Change Engine
-- Tracks planning entity changes and their impact on the project.

CREATE TABLE IF NOT EXISTS change_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  change_type TEXT NOT NULL,
  summary TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  severity TEXT NOT NULL DEFAULT 'medium',
  status TEXT NOT NULL DEFAULT 'pending',
  source TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS change_reports (
  id TEXT PRIMARY KEY,
  change_event_id TEXT NOT NULL,
  project_id TEXT NOT NULL,
  analysis TEXT NOT NULL DEFAULT '{}',
  affected_entities TEXT NOT NULL DEFAULT '[]',
  review_required INTEGER NOT NULL DEFAULT 0,
  summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (change_event_id) REFERENCES change_events(id)
);

CREATE TABLE IF NOT EXISTS snapshots_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  entity_snapshot TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS entity_states (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'valid',
  last_change_id TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, entity_type, entity_id)
);

CREATE INDEX IF NOT EXISTS idx_change_events_project ON change_events(project_id);
CREATE INDEX IF NOT EXISTS idx_change_events_type ON change_events(change_type);
CREATE INDEX IF NOT EXISTS idx_change_events_status ON change_events(status);
CREATE INDEX IF NOT EXISTS idx_change_reports_event ON change_reports(change_event_id);
CREATE INDEX IF NOT EXISTS idx_change_reports_project ON change_reports(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_v2_project ON snapshots_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_project ON entity_states(project_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_type ON entity_states(entity_type);
CREATE INDEX IF NOT EXISTS idx_entity_states_status ON entity_states(status);
