# ADR 0010: Planning Framework

## Decision
Create `MasterPlan`, `SpecificPlan`, and `ImplementationDocument` services that derive from Vision, Approved Context, Research, and Knowledge.

## Rationale
Plans must be grounded in approved/stored state instead of transient messages.

## Consequences
- Planning is deterministic artifact creation.
- Future orchestration remains out of scope.
