# Storage Overview

Plan-AI uses SQLite for local durable state.

## Current state

- Global layout: `~/.plan-ai/` or sandboxed `PLAN_AI_HOME` root.
- Project layout: `<project>/.plan-ai/` or sandboxed `PLAN_AI_PROJECT_ROOT`.
- Global DB: `global.db`.
- Project DB: `project.db`.
- Migrations are currently inline in `internal/store/store.go`.

## Current active project migrations

- `0001_project_base`
- `0002_project_domain`
- `0003_project_scan`
- `0004_knowledge_base`
- `0005_research_engine`

The old `0006_skill_intelligence` migration was archived because the roadmap changed.

## Direction

Phase 8 will define the definitive Store Layer after Phase 7 defines the definitive domain model. FTS5 should be introduced where it supports real query needs, not as decoration.
