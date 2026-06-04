# Change Engine

The Change Engine is the core of Plan-AI's invalidation-based planning.
It tracks every change to planning entities (visions, requirements, plans,
phases, tasks, decisions, research, knowledge) and determines what must
be reviewed, re-approved, or rebuilt as a result.

## Core Concepts

### Change Types

| Type | Severity | Description |
|------|----------|-------------|
| `vision_changed` | high | Project vision or high-level direction changed |
| `requirement_added` | medium | A new requirement was added |
| `requirement_removed` | high | An existing requirement was removed |
| `constraint_changed` | medium | A project constraint was modified |
| `decision_changed` | medium | A design/technical decision was changed |
| `research_updated` | medium | Research findings or conclusions were updated |
| `knowledge_updated` | low | Knowledge base content was updated |
| `plan_changed` | high | A plan structure or content changed |
| `technology_changed` | low | Technology choices or dependencies changed |
| `implementation_feedback` | low | Feedback from implementation affecting future plans |

### Invalidation Rules

Each entity type has specific change types that invalidate it:

```
vision       ← vision_changed, constraint_changed, decision_changed
requirement  ← requirement_added/removed, constraint_changed, vision_changed
constraint   ← constraint_changed, vision_changed, decision_changed
decision     ← decision_changed, research_updated, knowledge_updated
research     ← research_updated, knowledge_updated
knowledge    ← knowledge_updated
master_plan  ← vision_changed, requirement_added/removed, constraint_changed,
               decision_changed, research_updated, knowledge_updated, plan_changed
specific_plan← requirement_added/removed, constraint_changed, decision_changed,
               research_updated, knowledge_updated, plan_changed, implementation_feedback
phase        ← plan_changed, requirement_added/removed, implementation_feedback
task         ← plan_changed, requirement_removed, decision_changed, implementation_feedback
validation   ← plan_changed, requirement_removed, decision_changed, implementation_feedback
```

### Statuses

- **current** — valid, no action needed
- **outdated** — may be stale, refresh recommended
- **needs_review** — must be reviewed before reuse
- **blocked** — cannot proceed until resolved

## Architecture

The Change Engine lives in `internal/change/` and provides:

- **Registry** — enumerates all valid change types with metadata
- **Service** — registers changes, analyses impact, creates snapshots
- **Analyzer** — determines affected entity types from change types
- **ImpactBuilder** — builds detailed impact analyses
- **VersionManager** — tracks entity invalidation states
- **SnapshotManager** — captures point-in-time project snapshots

## Usage

### Via Go API

```go
srv := change.NewService(changeStore, snapshotStore)
report, err := srv.RegisterChange(change.ChangeEvent{
    ProjectID:  "proj_1",
    ChangeType: change.RequirementAdded,
    Summary:    "Add user authentication",
    Severity:   change.SeverityMedium,
})
```

### Via CLI

```sh
plan-ai impact                          # list recent changes and their impact
plan-ai snapshot                        # create a snapshot
plan-ai snapshot list                   # list recent snapshots
```

### Via MCP

The Change Engine tools are exposed as MCP tools:

- `plan_ai.analyze_impact` — runs impact analysis on a change
- `plan_ai.create_snapshot` — creates a point-in-time snapshot
