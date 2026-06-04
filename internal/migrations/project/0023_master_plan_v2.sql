-- Phase 25: Master Plan Generator v2
-- Adds versioning, history, change tracking, approvals, and evolution records.

CREATE TABLE IF NOT EXISTS master_plan_versions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    version INT NOT NULL,
    plan_id TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    phases TEXT NOT NULL DEFAULT '[]',
    timeline TEXT NOT NULL DEFAULT '{}',
    risks TEXT NOT NULL DEFAULT '[]',
    dependencies TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'draft',
    changelog TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);

CREATE TABLE IF NOT EXISTS master_plan_changes (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    version_from INT NOT NULL,
    version_to INT NOT NULL,
    change_type TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    author TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);

CREATE TABLE IF NOT EXISTS master_plan_approvals (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    version INT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    approved_by TEXT,
    approved_at TEXT,
    feedback TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);

CREATE TABLE IF NOT EXISTS plan_evolution_events (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    details TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_mp_versions_plan ON master_plan_versions(plan_id);
CREATE INDEX IF NOT EXISTS idx_mp_changes_plan ON master_plan_changes(plan_id);
CREATE INDEX IF NOT EXISTS idx_mp_approvals_plan ON master_plan_approvals(plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_evolution_project ON plan_evolution_events(project_id);
