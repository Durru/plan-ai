-- Phase 20: OpenCode Integration
-- Tracks OpenCode detection results and integration health.

CREATE TABLE IF NOT EXISTS opencode_detections (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL,
  found INTEGER NOT NULL DEFAULT 0,
  config_path TEXT NOT NULL DEFAULT '',
  is_initialized INTEGER NOT NULL DEFAULT 0,
  agent_name TEXT NOT NULL DEFAULT '',
  skill_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS opencode_integration_state (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL UNIQUE,
  mode TEXT NOT NULL DEFAULT 'standalone',
  enabled INTEGER NOT NULL DEFAULT 1,
  read_only INTEGER NOT NULL DEFAULT 1,
  last_detected_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS opencode_doctor_checks (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL,
  check_name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pass',
  message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_opencode_detections_project ON opencode_detections(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_integration_project ON opencode_integration_state(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_doctor_project ON opencode_doctor_checks(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_doctor_name ON opencode_doctor_checks(check_name);
