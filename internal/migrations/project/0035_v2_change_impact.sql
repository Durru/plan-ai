CREATE TABLE IF NOT EXISTS change_impact_reports_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  change_type TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  architecture_concerns TEXT NOT NULL DEFAULT '[]',
  backend_concerns TEXT NOT NULL DEFAULT '[]',
  migration_concerns TEXT NOT NULL DEFAULT '[]',
  docs_concerns TEXT NOT NULL DEFAULT '[]',
  api_concerns TEXT NOT NULL DEFAULT '[]',
  plan_concerns TEXT NOT NULL DEFAULT '[]',
  validation_commands TEXT NOT NULL DEFAULT '[]',
  rollback_strategy TEXT NOT NULL DEFAULT '[]',
  affected_plans TEXT NOT NULL DEFAULT '[]',
  affected_tasks TEXT NOT NULL DEFAULT '[]',
  severity TEXT NOT NULL DEFAULT 'medium',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_change_impact_reports_v2_project ON change_impact_reports_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_change_impact_reports_v2_type ON change_impact_reports_v2(change_type);
