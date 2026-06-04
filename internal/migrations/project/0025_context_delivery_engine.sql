-- Phase 27: Context Delivery Engine
-- Adds budget-aware context delivery, session tracking, usage metrics.

CREATE TABLE IF NOT EXISTS context_delivery_sessions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    level TEXT NOT NULL,
    budget_tokens INT NOT NULL DEFAULT 0,
    tokens_used INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL DEFAULT '',
    metadata TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'delivered',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS context_delivery_usage (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    session_id TEXT,
    level TEXT NOT NULL,
    tokens INT NOT NULL DEFAULT 0,
    source TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (session_id) REFERENCES context_delivery_sessions(id)
);

CREATE TABLE IF NOT EXISTS context_delivery_budgets (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    level TEXT NOT NULL,
    max_tokens INT NOT NULL DEFAULT 4096,
    current_usage INT NOT NULL DEFAULT 0,
    strategy TEXT NOT NULL DEFAULT 'fixed',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_ctx_session_project ON context_delivery_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_ctx_usage_project ON context_delivery_usage(project_id);
CREATE INDEX IF NOT EXISTS idx_ctx_budgets_level ON context_delivery_budgets(level);
