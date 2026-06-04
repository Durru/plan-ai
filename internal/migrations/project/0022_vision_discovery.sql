-- Phase 24: Vision Discovery Engine
-- Adds discovery sessions, assumptions, ambiguities, and approval flow.

CREATE TABLE IF NOT EXISTS vision_discovery_sessions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    summary TEXT NOT NULL DEFAULT '',
    raw_context TEXT NOT NULL DEFAULT '',
    findings TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS vision_assumptions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL DEFAULT 0.5,
    status TEXT NOT NULL DEFAULT 'unvalidated',
    validated_by TEXT,
    validated_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS vision_ambiguities (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    resolution TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    resolved_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS vision_approvals (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    vision_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    approved_by TEXT,
    approved_at TEXT,
    feedback TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_vision_discovery_project ON vision_discovery_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_vision_assumptions_session ON vision_assumptions(session_id);
CREATE INDEX IF NOT EXISTS idx_vision_ambiguities_session ON vision_ambiguities(session_id);
CREATE INDEX IF NOT EXISTS idx_vision_approvals_vision ON vision_approvals(vision_id);
