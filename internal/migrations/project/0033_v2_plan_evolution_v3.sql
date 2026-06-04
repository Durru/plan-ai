CREATE TABLE IF NOT EXISTS plan_evolution_blueprints_v3 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  objective TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT '[]',
  exclusions TEXT NOT NULL DEFAULT '[]',
  dependencies TEXT NOT NULL DEFAULT '[]',
  stack TEXT NOT NULL DEFAULT '[]',
  versions TEXT NOT NULL DEFAULT '[]',
  libraries TEXT NOT NULL DEFAULT '[]',
  folders TEXT NOT NULL DEFAULT '[]',
  files TEXT NOT NULL DEFAULT '[]',
  validations TEXT NOT NULL DEFAULT '[]',
  tests TEXT NOT NULL DEFAULT '[]',
  risks TEXT NOT NULL DEFAULT '[]',
  rollback TEXT NOT NULL DEFAULT '[]',
  approved_inputs TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_plan_evolution_blueprints_v3_project ON plan_evolution_blueprints_v3(project_id);
