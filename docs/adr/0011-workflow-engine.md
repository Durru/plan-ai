# ADR 0011: Workflow Engine

## Decision
Add a small workflow registry and `workflow_runs` persistence for core project workflows.

## Rationale
Plan-AI should formalize process boundaries while leaving intelligence to the LLM/human.

## Consequences
- Workflows are named step sequences.
- Execution records status and timing only.
- Phase 15+ engines remain pending.
