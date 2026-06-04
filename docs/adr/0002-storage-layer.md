# ADR 0002: Storage Layer

## Status

Accepted

## Context

Plan-AI needs persistent local state for installation configuration and project initialization. The storage layer must support future planning data without implementing planner, research, MCP, agent, Engram, or skills behavior in Phase 2.

## Decision

Use SQLite for persistence with separate global and project databases:

- Global database: `~/.plan-ai/global.db`
- Project database: `<project>/.plan-ai/project.db`

Use `modernc.org/sqlite` as the SQLite driver to avoid CGO. Use a simple internal migration runner rather than adding a heavy migration framework.

## Rationale

SQLite is a good fit for a local-first CLI because it is durable, queryable, transactional, and does not require a server. The pure Go `modernc.org/sqlite` driver keeps installation simple on Ubuntu 22.04 and avoids native compiler requirements.

Separating global and project stores keeps installation preferences and reusable data independent from repository-local project state. This also makes project data easier to inspect, back up, and eventually version or export.

The internal migration runner is intentionally small. Phase 2 needs only idempotent table creation and a migration ledger with `id`, `name`, and `applied_at`.

## Consequences

- `plan-ai install` creates and migrates the global store.
- `plan-ai init` creates and migrates the project store and registers the project in `known_projects` when the global store exists.
- Commands are idempotent and never delete existing content.
- Future planning tables are not created yet; they belong to later phases.
