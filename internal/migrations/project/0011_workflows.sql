CREATE TABLE IF NOT EXISTS workflow_runs (id TEXT PRIMARY KEY, workflow_type TEXT NOT NULL, status TEXT NOT NULL, started_at TEXT NOT NULL, finished_at TEXT NOT NULL DEFAULT '');
CREATE INDEX IF NOT EXISTS idx_workflow_runs_type ON workflow_runs(workflow_type);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs(status);
