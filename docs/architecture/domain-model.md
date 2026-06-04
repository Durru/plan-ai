# Domain Model

Phase 3 introduces Plan-AI's structured planning records as project-local SQLite data. This phase only models and stores records; it does not generate plans, run agents, scan skills, integrate MCP/OpenCode, or execute research workflows.

## Core Entities

### Project

A project represents the repository or workspace being planned. Project identity is currently managed by Phase 2 project state and config; the domain struct exists so future planning code can reason about projects consistently.

### Plan

A plan is the highest-level planning artifact stored in `plans`.

Plan types are explicit:

- `master`: the top-level plan for a project or major objective.
- `specific`: a scoped plan that can point back to a master plan through `parent_plan_id`.

Plans have status and version fields so they can move from draft thinking to approved implementation readiness without losing history.

### Phase

A phase groups related work under a plan. Phases are ordered with `position` and carry their own status so a plan can be partially blocked, approved, implemented, or validated.

### Task

A task is the smallest implementation-ready planning unit. It belongs to both a phase and a plan, has an ordered `position`, and declares a `context_size`:

- `short`: minimal task context.
- `medium`: operational context.
- `full`: complete planning context.

This supports Plan-AI's context-on-demand principle.

### Decision

Decisions are first-class because plans are only trustworthy when their architectural and product commitments are explicit. A decision records:

- the context that forced the choice,
- the chosen direction,
- its status,
- and its expected impact.

Future replanning can trace from a changed decision to affected plans, phases, tasks, validations, and documentation.

### Research Entry

A research entry captures a source-aware investigation result. It records topic, source, summary, conclusion, and confidence.

Research entries are evidence records: what was looked up, where it came from, and what it appeared to mean at the time.

### Knowledge Object

A knowledge object is reusable project knowledge distilled from research, decisions, implementation discoveries, or documentation. It includes `reuse_count` because Plan-AI should prefer known project knowledge before repeating research.

Research Entry vs Knowledge Object:

- Research Entry = source-bound finding.
- Knowledge Object = reusable project memory.

### Validation

A validation records whether a target is ready or verified. Validation target types are explicit:

- `plan`
- `phase`
- `task`
- `decision`

This keeps validation attached to the planning object it evaluates.

### Snapshot

A snapshot records a lightweight checkpoint with a reason and summary. Snapshots exist early so future phases can preserve planning state before major changes, replanning, or validation runs.

## Statuses

Official statuses are Go constants, not loose strings:

- `draft`
- `in_review`
- `approved`
- `rejected`
- `blocked`
- `implemented`
- `validated`
- `archived`

The same status vocabulary is shared by plans, phases, tasks, decisions, and validations where applicable.

## Relationships

```text
Project
└── Master Plan
    └── Specific Plan
        └── Phase
            └── Task

Decision ── influences plans/phases/tasks conceptually
Research Entry ── evidence source
Knowledge Object ── reusable distilled knowledge
Validation ── targets plan/phase/task/decision
Snapshot ── project-local checkpoint
```

Phase 3 intentionally keeps relationships simple. The database stores core references such as `parent_plan_id`, `plan_id`, `phase_id`, `target_type`, and `target_id`, but does not add advanced graph traversal, FTS, or automatic impact analysis.
