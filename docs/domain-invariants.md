# Domain Invariants

Invariants are constraints that must always hold true for the domain model.
Some are enforced by the Go type system, some by helper functions, and some
are documented expectations for future enforcement layers.

## Structural Invariants

### Entity Identity

Every entity has a non-empty `ID` string. IDs are generated via `domain.NewID(prefix)`:

```
NewID("plan")   → "plan_<16-hex-bytes>"
NewID("vision") → "vision_<16-hex-bytes>"
```

The `NewID` function guarantees non-empty, unique IDs within a 2^128 space.

### Ownership Chain

Every entity is reachable from a Project through a chain of foreign keys:

```
Project
├── Vision (project_id)
├── Requirement (project_id)
├── Constraint (project_id)
├── Decision (project_id)
├── MasterPlan (project_id)
│   └── SpecificPlan (project_id, master_plan_id)
│       └── Phase (plan_id)
│           └── Task (phase_id, plan_id)
├── ImplementationDocument (specific_plan_id)
├── Research (project_id)
├── KnowledgeObject (no direct project_id — cross-project)
├── Snapshot (project_id)
├── ChangeRequest (project_id)
│   └── ImpactReport (change_request_id)
└── Validation (target_id references any entity with a target_type)
```

A Validation's `target_id` + `target_type` must reference an existing entity.
This is a documented expectation; Phase 8 should add referential enforcement.

### Timestamps

Every entity has `CreatedAt` and `UpdatedAt` fields. `CreatedAt` must not be zero.
`UpdatedAt` is set to `CreatedAt` on initial creation, and updated on every mutation.

### Typed Statuses

Every mutable entity carries a `Status` field. The valid values depend on the
entity type (see [domain-lifecycle.md](domain-lifecycle.md)). The type system
enforces the value space for entity-specific status types (e.g., `ProjectStatus`,
`ChangeRequestStatus`). For entities using the shared `Status` type, canonical
values are documented but not type-enforced.

## Lifecycle Invariants

### Terminal States

The following states are terminal — no transitions out:

| Entity | Terminal status |
|---|---|
| Project | `archived` |
| Decision | `deprecated` |
| MasterPlan | `archived` |
| SpecificPlan | `archived` |
| Phase | `completed` |
| Task | `validated` |
| ChangeRequest | `applied`, `rejected` |

### Prohibited Transitions

From [domain-lifecycle.md](domain-lifecycle.md):

- **Project**: `completed` → `draft` / `active` ❌; `archived` → anything ❌
- **Decision**: `approved` → `rejected` ❌; `deprecated` → anything ❌
- **MasterPlan**: `draft` → `approved` ❌ (must go through `review`)
- **SpecificPlan**: `approved` → `draft` ❌ (must go through `blocked`)
- **Phase**: `completed` → anything ❌
- **Task**: `validated` → anything ❌
- **ChangeRequest**: `rejected` → anything ❌

## Approval Invariants

### Vision Approval

A `Vision` can only be referenced for planning if `Approved == true`.
No formal enforcement exists yet — this is a documented expectation for Phase 8+.

### Requirement Approval

A `Requirement` must be `Approved == true` before it can be mapped to a
SpecificPlan or Task. Exception: `RequirementTypeTechnical` may appear in
Research as a proposal without being approved.

### Decision Approval

A `Decision` must have `Status == StatusApproved` before it can be used
as a `KnowledgeReference`. Draft and proposed decisions are provisional.

## Knowledge Invariants

### Reuse Counter Integrity

`KnowledgeObject.ReuseCount` is monotonic: it only increases via
`IncrementReuseCount()`. It must never be decremented or reset.

### Relation Uniqueness

A `KnowledgeRelation` must be unique on `(source_id, target_id, relation_type)`.
Duplicate relation inserts are silently ignored (`ON CONFLICT DO NOTHING`).

### Tag Uniqueness

A `KnowledgeTag` must be unique on `(knowledge_id, tag)`.
Duplicate tag inserts are silently ignored.

## Research Invariants

### Source Requirement for Approval

A `Research` entry should have at least one `ResearchSource` before being
promoted to `KnowledgeObject`. This is enforced by the research engine's
`ApprovalChecker` gateway (in Phase 6), not by the domain model itself.

## Change Invariants

### ImpactReport Existence

Every `ChangeRequest` that transitions to `approved` or `applied` must have
an associated `ImpactReport`. This is a documented expectation for Phase 8+.

### Immutable ImpactReport

Once created, an `ImpactReport` is immutable. If the change scope changes,
a new `ChangeRequest` must be created.

## Migration Invariants

### Non-Destructive Evolution

Existing database tables (migrations 0001–0005) must never be dropped or
destructively altered by domain model changes. New entities may add tables,
but existing data must remain readable. See [0006-domain-model-v2.md](adr/0006-domain-model-v2.md).

### Backward Compatible Structs

Existing entity structs retain all original fields. New fields are additive
and either optional (zero-value safe) or have defaults. Field types are
preserved unless the change is transparent (e.g., adding a new field).
