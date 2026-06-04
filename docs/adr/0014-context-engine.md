# ADR 0014: Context Engine

## Decision
Add a context engine that builds composite context views (executive, planning, implementation, research) from approved context, domain data, visions, and research entries, persisted as `context_views_v2`.

## Rationale
Different phases and roles need different slices of project knowledge. A unified builder pattern avoids duplicating context-assembly logic across the orchestrator, planner, and research subsystems.

## Consequences
- Context views are versioned and persisted for auditing and replay.
- The builder accepts pluggable queriers (vision, planning) for extensibility.
- CLI integration (`plan-ai context`) exposes the executive view to users.
- Phase 18+ (agent system) will consume context views as session input.
