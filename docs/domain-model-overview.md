# Domain Model Overview

## Purpose

The definitive Plan-AI domain model describes the durable project truth that
Plan-AI manages before, during, and after implementation. Every entity,
relationship, status, and invariant is defined here.

## Entities

| Entity | File | Responsibility |
|---|---|---|
| Project | `project.go` | Top-level container; every operation belongs to one project. |
| Vision | `vision.go` | Strategic direction: what problem, what success looks like. |
| Requirement | `requirement.go` | Typed capability/constraint the project must satisfy. |
| Constraint | `constraint.go` | Boundary: budget, stack, time, compliance, resource. |
| Decision | `decision.go` | Architectural or design choice with rationale and alternatives. |
| Research | `research.go` | Provisional investigation with sources and findings. |
| KnowledgeObject | `knowledge.go` | Curated, reusable project knowledge for later phases. |
| MasterPlan | `planning.go` | Top-level strategic plan with high-level phases. |
| SpecificPlan | `planning.go` | Concrete implementation plan derived from a MasterPlan. |
| Phase | `planning.go` | Major stage within a SpecificPlan; groups tasks. |
| Task | `planning.go` | Smallest unit of planned work. |
| ImplementationDocument | `planning.go` | Deliverable guide derived from a SpecificPlan (not a plan). |
| Validation | `validation.go` | Single check against a target entity with pass/fail outcome. |
| Snapshot | `snapshot.go` | Point-in-time capture of decisions, plans, and context. |
| ChangeRequest | `change.go` | Proposal to modify plans, decisions, or scope. |
| ImpactReport | `change.go` | Entities/plans/decisions affected by a ChangeRequest. |

## Phase 9–11 Operational Models

The Phase 7 domain model remains the frozen core. Phases 9–11 add operational
models around that core without changing `internal/domain`:

- `internal/ingestion` stores `RawInput` and `IngestedSource` records.
- `internal/vision` creates incomplete vision drafts from ingested sources.
- `internal/context` stores only approved requirements, constraints, decisions,
  preferences, goals, and references.

This keeps the flow explicit: input is captured first, vision is drafted second,
and durable reusable context is created only after approval.

## Phase 12–14 Operational Models

The Phase 7 domain model remains frozen. Phases 12–14 add operational services
around that core without changing `internal/domain`:

- `internal/research` stores project research jobs, findings, recommendations, and sources.
- `internal/knowledge` stores reusable knowledge linked to research and approved context.
- `internal/planning` stores master plans, specific plans, and implementation documents.
- `internal/workflows` stores registered workflow definitions and workflow run records.

Plans are derived from Vision, Approved Context, Research, and Knowledge—not from raw messages.

## Relations

See [domain-relations.md](domain-relations.md) for the full relationship tree.

## Lifecycle

See [domain-lifecycle.md](domain-lifecycle.md) for valid/prohibited status transitions.

## Invariants

See [domain-invariants.md](domain-invariants.md) for constraints enforced by the model.

## Design Decisions

See [adr/0006-domain-model-v2.md](adr/0006-domain-model-v2.md) for the rationale
behind the definitive domain model redesign.

## Package Layout

```
internal/domain/
  domain.go           — shared types (Status, KnowledgeCategory, NewID, …)
  project.go          — Project, ProjectStatus
  vision.go           — Vision
  requirement.go      — Requirement, RequirementType
  constraint.go       — Constraint, ConstraintType
  decision.go         — Decision, DecisionStatus transitions
  research.go         — Research, ResearchSource
  knowledge.go        — KnowledgeObject, KnowledgeTag, KnowledgeRelation, …
  planning.go         — MasterPlan, SpecificPlan, Phase, Task, ImplementationDocument
  validation.go       — Validation, ValidationType, ValidationStatus
  snapshot.go         — Snapshot
  change.go           — ChangeRequest, ChangeRequestStatus, ImpactReport
  repositories.go     — Repository interfaces for all entities
  domain_test.go      — Tests for shared types and entities
```
