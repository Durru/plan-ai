# ADR 0006: Definitive Domain Model v2

## Status

Accepted

## Context

The initial domain model (ADR 0003) was defined as a minimal set of entities
to support Phases 3–6: plans, phases, tasks, decisions, research, knowledge,
validations, and snapshots. All structs lived in a single `domain.go` file.

Over Phases 3–6, the codebase grew to:

- 14 internal packages importing domain types
- A CLI surface with 30+ sub-commands
- A research engine with its own `Repository` interface (Phase 6)
- A knowledge base with its own `Repository` interface (Phase 5)
- A store layer with concrete SQLite repositories
- Multiple test files constructing domain entities by field name

The original monolithic `domain.go` became difficult to navigate and extend.
Adding new entities (Vision, Requirement, Constraint, ChangeRequest,
ImpactReport) would have made it worse. The model also lacked:

- Typed statuses per entity (everything used the generic `Status` string)
- Explicit `ProjectID` ownership on all entities
- Formal lifecycle transition rules
- A `ProjectStatus` type distinct from the generic `Status`
- An `ImplementationDocument` entity (derived from a SpecificPlan, not a plan)
- A `ChangeRequest` + `ImpactReport` pair for scope management
- Repository interfaces that Phase 8 can implement

## Decision

Redesign the domain model into a definitive, per-file, documented model with:

### One entity per file

| File | Entity |
|---|---|
| `project.go` | Project with `ProjectStatus` |
| `vision.go` | Vision |
| `requirement.go` | Requirement with `RequirementType` |
| `constraint.go` | Constraint with `ConstraintType` |
| `decision.go` | Decision with rationale/alternatives |
| `research.go` | Research with objective/sources |
| `knowledge.go` | KnowledgeObject with tags/relations/references |
| `planning.go` | MasterPlan, SpecificPlan, Phase, Task, ImplementationDocument |
| `validation.go` | Validation with typed outcome |
| `snapshot.go` | Snapshot with ProjectID |
| `change.go` | ChangeRequest, ImpactReport |
| `repositories.go` | Repository interfaces |

### Explicit typed statuses

- `ProjectStatus` — draft, active, paused, completed, archived
- `DecisionProposed` / `DecisionDeprecated` — shared `Status` extensions
- `PlanStatusReview` / `PlanStatusPending` / `PlanStatusDone` / `PlanStatusActive`
- `ValidationType` — manual, automatic
- `ValidationStatus` — pending, passed, failed
- `ChangeRequestStatus` — draft, submitted, approved, rejected, applied

### Lifecycle transition helpers

- `ValidProjectTransitions(from, to ProjectStatus) bool`
- `ValidDecisionTransitions(from, to Status) bool`
- `ValidMasterPlanTransitions(from, to Status) bool`
- `ValidSpecificPlanTransitions(from, to Status) bool`
- `ValidPhaseTransitions(from, to Status) bool`
- `ValidTaskTransitions(from, to Status) bool`

### Backward compatibility

All existing struct fields are preserved. New fields are additive.
Existing consumers (store layer, research package, knowledge package, CLI)
continue to compile and function without changes.

### `domain.go` as shared types only

The legacy `domain.go` now contains only shared enum types (`Status`,
`PlanType`, `ContextSize`, `KnowledgeCategory`, etc.) and the `NewID()`
utility. Entity structs have been moved to their dedicated files.

## Consequences

### Positive

1. **Navigability**: Each entity has a single file. Find any entity by filename.
2. **Documentability**: Lifecycle rules and invariants are documented alongside
   the entities that define them.
3. **Extensibility**: Adding a new entity means adding one file, not bloating `domain.go`.
4. **Repository interfaces**: Phase 8 can implement `domain.ProjectRepository`,
   `domain.VisionRepository`, etc. against any store backend.
5. **Typed statuses**: Compile-time safety for entity-specific statuses.
6. **Transition helpers**: Validate state changes before persisting them.

### Negative

1. **Dual hierarchy**: `MasterPlan` and `SpecificPlan` now both have `ProjectID`
   and the legacy `Plan` struct still exists for store compatibility.
2. **Some duplicated fields**: `SpecificPlan` has both `MasterPlanID` and
   `ParentPlanID` (legacy alias) for backward compatibility.
3. **Documentation surface**: Four new documentation files must be kept in sync
   with code changes.

### Migration

Existing tables (migrations 0001–0005) are untouched. New entities
(Vision, Requirement, Constraint, ChangeRequest, ImpactReport,
ImplementationDocument) will get tables in a future migration (Phase 8).
The domain model is the specification for that migration.

## Non-Scope (Deferred to Phase 8+)

- Actual store implementation of the new repository interfaces
- New SQL migration for new entity tables
- Validation engine (beyond transition helpers)
- Enforcement engine for approval invariants
- Vision Engine, Research Engine (advanced), Planner Engine, Context Engine
- MCP or OpenCode integration
- Agents or sub-agents
- Model strategy or LLM integration
