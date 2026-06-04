# ADR 0003: Introduce the Project Domain Model Before Planner Logic

## Status

Accepted

## Context

Plan-AI needs to reason about plans, phases, tasks, decisions, research, reusable knowledge, validations, and snapshots. It would be tempting to start with planner generation logic, but generated planning behavior without durable domain records would be difficult to inspect, test, validate, or evolve.

Phase 3 follows the project principle that Plan-AI prepares implementation. Before any Planner, Research Engine, MCP integration, OpenCode integration, Engram integration, skills scanner, real agents, or automatic plan generation exists, the project needs a small trustworthy data model.

## Decision

Add project-local domain tables and repositories before implementing planner behavior.

The project database now stores:

- `plans`
- `phases`
- `tasks`
- `decisions`
- `research_entries`
- `knowledge_objects`
- `validations`
- `snapshots`

Plans are structured data with explicit `type`, `status`, `version`, and optional `parent_plan_id`. The system distinguishes master plans from specific plans instead of relying on document naming conventions.

Research and knowledge are separated:

- Research entries preserve source-bound findings and confidence.
- Knowledge objects preserve reusable distilled project knowledge.

Snapshots are introduced early to support future state preservation before replanning, validation, or major planning changes.

## Consequences

### Positive

- Future planner logic has stable persistence boundaries.
- CLI and tests can inspect domain state without invoking generation.
- Decisions become auditable project records instead of hidden chat context.
- Research can be retained without prematurely turning every finding into reusable knowledge.
- Snapshots provide a future-safe checkpoint mechanism.

### Negative

- Phase 3 adds tables and repositories before higher-level behavior exists.
- Some relationships remain conceptual until future planner/replanning phases use them.
- The project has more persistence surface to maintain.

## Non-Scope

This ADR does not approve or implement:

- real Planner behavior,
- automatic plan generation,
- Research Engine behavior,
- MCP or OpenCode integration,
- Engram integration,
- skills scanning,
- real agent execution,
- advanced search or FTS.

Phase 3 only establishes the durable domain substrate.
