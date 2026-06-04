# ADR 0007: Store Layer v2

## Status

Accepted for Phase 8.

## Context

Phase 7 introduced the definitive domain model under `internal/domain`. The old store package already had inline migrations and repositories used by CLI flows and tests. Phase 8 must connect the domain model to SQLite persistence without breaking legacy behavior.

## Decision

Use additive SQLite migrations for a two-store architecture:

- Global store at `~/.plan-ai/global.db`.
- Project store at `<project>/.plan-ai/project.db`.

The implementation keeps legacy tables and adds definitive tables. Repositories for the Phase 7 domain interfaces live in `internal/store/repositories`, while existing store repositories remain available for current CLI behavior.

`PLAN_AI_HOME` remains a home root, not the final `.plan-ai` directory.

## Consequences

- Existing CLI commands and tests keep working.
- Future engines can persist canonical entities without needing a schema rewrite.
- The schema contains both legacy and definitive planning tables during migration.
- FTS5 tables are present where supported; repository search still has a simple fallback.

## Deferred

Phase 9 may add ingestion logic, engine-owned write patterns, richer FTS synchronization, vector search, advanced research flows, Context Engine, MCP/OpenCode integration, and external service coordination.
