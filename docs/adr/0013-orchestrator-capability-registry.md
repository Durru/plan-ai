# ADR 0013: Orchestrator & Capability Registry

## Decision
Add an orchestrator that manages job lifecycle (create, dispatch, execute, complete/fail) and a capability registry that maps workflow types to their handlers.

## Rationale
Multi-step workflows (planning, research, implementation) need a job queue, run tracking, and capability-based dispatch rather than ad-hoc execution.

## Consequences
- Jobs and job runs are persisted in the project database.
- The capability registry provides a default set of workflow-to-capability mappings.
- Model strategy integration happens at dispatch time (selecting capability-appropriate profiles).
- Phase 17 (context engine) feeds into the orchestrator for job-scoped context.
