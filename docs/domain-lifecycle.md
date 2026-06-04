# Domain Lifecycle — Status Transitions

Each entity type has a canonical set of status values and valid transitions.
This document defines them.

## Shared Status Values

These are defined in `domain.go` and used across multiple entity types:

| Constant | Value |
|---|---|
| `StatusDraft` | `draft` |
| `StatusInReview` | `in_review` |
| `StatusApproved` | `approved` |
| `StatusRejected` | `rejected` |
| `StatusBlocked` | `blocked` |
| `StatusImplemented` | `implemented` |
| `StatusValidated` | `validated` |
| `StatusArchived` | `archived` |
| `StatusCompleted` | `completed` |
| `DecisionProposed` | `proposed` |
| `DecisionDeprecated` | `deprecated` |
| `PlanStatusReview` | `review` |
| `PlanStatusPending` | `pending` |
| `PlanStatusDone` | `done` |
| `PlanStatusValidated` | `validated` |
| `PlanStatusActive` | `active` |
| `PlanStatusCompleted` | `completed` |

## Project

- **Type**: `ProjectStatus` (`project.go`)
- **Helper**: `ValidProjectTransitions(from, to ProjectStatus) bool`

```
draft ──→ active ──→ paused ──→ active
  │         │          │
  │         └──→ completed ──→ archived
  │                        │
  └──→ archived            │
               ↑ ←─────────┘
```

**Prohibited**:
- archived → *anything* (terminal)
- completed → draft, active
- draft → paused, completed

## Decision

- **Field type**: `Status` (backward compat)
- **Canonical values**: `proposed`, `approved`, `rejected`, `deprecated`
- **Helper**: `ValidDecisionTransitions(from, to Status) bool`

```
proposed ──→ approved ──→ deprecated
    │
    └──→ rejected ──→ proposed (reconsider)
```

**Prohibited**:
- approved → rejected, proposed
- rejected → approved, deprecated
- deprecated → *anything* (terminal)

## MasterPlan

- **Field type**: `Status` (backward compat)
- **Canonical values**: `draft`, `review`, `approved`, `archived`
- **Helper**: `ValidMasterPlanTransitions(from, to Status) bool`

```
draft ──→ review ──→ approved ──→ archived
```

**Prohibited**:
- draft → approved, archived (must go through review)
- review → archived (must be approved first)
- archived → *anything* (terminal)

## SpecificPlan

- **Field type**: `Status` (backward compat)
- **Canonical values**: `draft`, `review`, `approved`, `blocked`, `archived`
- **Helper**: `ValidSpecificPlanTransitions(from, to Status) bool`

```
draft ──→ review ──→ approved ──→ archived
  ↑                  │
  └── blocked ←──────┘
```

**Prohibited**:
- draft → approved (must go through review)
- approved → draft (must go through blocked)
- archived → *anything* (terminal)

## Phase

- **Field type**: `Status` (backward compat)
- **Canonical values**: `pending`, `active`, `completed`, `blocked`
- **Helper**: `ValidPhaseTransitions(from, to Status) bool`

```
pending ──→ active ──→ completed
  ↑           │
  └── blocked ←┘
```

**Prohibited**:
- pending → completed, blocked
- completed → *anything* (terminal)
- active → pending

## Task

- **Field type**: `Status` (backward compat)
- **Canonical values**: `pending`, `active`, `done`, `validated`, `blocked`
- **Helper**: `ValidTaskTransitions(from, to Status) bool`

```
pending ──→ active ──→ done ──→ validated
  ↑           │
  └── blocked ←┘
```

**Prohibited**:
- pending → done, validated
- validated → *anything* (terminal)
- active → pending (must go through blocked)

## Validation

- **Field type**: `Status` (backward compat, uses `ValidationStatus` constants)
- **Canonical values**: `pending`, `passed`, `failed`
- **Transitions**:

```
pending ──→ passed
pending ──→ failed
```

No transitions out of `passed` or `failed`. A new Validation record must be created for re-validation.

## Research

Legacy Research uses `ResearchStatus` (`draft`, `in_review`, `approved`, `rejected`, `archived`).
The canonical definitive Research entity uses no explicit status; it uses `Date` instead.
The `Status` field is retained for backward compatibility.

## ChangeRequest

- **Type**: `ChangeRequestStatus` (`change.go`)
- **Canonical values**: `draft`, `submitted`, `approved`, `rejected`, `applied`

```
draft ──→ submitted ──→ approved ──→ applied
              │
              └──→ rejected
```

**Prohibited**:
- draft → approved, applied (must be submitted first)
- rejected → *anything* (closed)
- applied → *anything* (terminal)
