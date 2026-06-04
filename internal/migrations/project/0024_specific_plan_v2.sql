-- Phase 26: Specific Plan Generator v2
-- Adds domain-specific plans, research integration, regeneration tracking.

CREATE TABLE IF NOT EXISTS specific_plan_versions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    version INT NOT NULL,
    plan_id TEXT NOT NULL,
    domain TEXT NOT NULL DEFAULT 'general',
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    tasks TEXT NOT NULL DEFAULT '[]',
    dependencies TEXT NOT NULL DEFAULT '[]',
    risks TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'draft',
    changelog TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id)
);

CREATE TABLE IF NOT EXISTS specific_plan_research_links (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    research_id TEXT NOT NULL,
    section TEXT NOT NULL DEFAULT '',
    relevance REAL NOT NULL DEFAULT 0.0,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id),
    FOREIGN KEY (research_id) REFERENCES research(id)
);

CREATE TABLE IF NOT EXISTS specific_plan_regenerations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    version_from INT NOT NULL,
    version_to INT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL DEFAULT 'full',
    status TEXT NOT NULL DEFAULT 'completed',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id)
);

CREATE INDEX IF NOT EXISTS idx_sp_versions_plan ON specific_plan_versions(plan_id);
CREATE INDEX IF NOT EXISTS idx_sp_research_plan ON specific_plan_research_links(plan_id);
CREATE INDEX IF NOT EXISTS idx_sp_regenerations_plan ON specific_plan_regenerations(plan_id);
