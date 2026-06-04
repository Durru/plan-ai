# ADR-0020: Continuous Planning

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 22

## Context

Plans become stale when the underlying project evolves — requirements change,
constraints shift, decisions are reversed. Manual re-planning is error-prone and
often forgotten. Plan-AI needs a continuous planning subsystem that detects changes,
proposes plan updates, and regenerates context automatically.

## Decision

Implement a Continuous Planning system as an internal package (`internal/continuous/`)
with these components:

1. **Event Detection** — monitor for events that may trigger re-planning:
   - File changes in the project
   - Research findings added or approved
   - Requirements, constraints, or decisions modified
   - Plans or tasks updated
   - External triggers (MCP, user request, schedule)
2. **Plan Update Proposal** — when an event is detected, generate a proposal
   describing what plans are affected and what updates are suggested
3. **Approval Workflow** — proposals require approval before they are applied
   (configurable via `requires_approval` flag)
4. **Context Generation** — produce context at multiple levels (L0–L4) on demand:
   - L0: Executive summary
   - L1: Planning status
   - L2: Specific plan details
   - L3: Task-level breakdown
   - L4: Implementation details

### Database Schema

New tables in migration 0020 (`continuous_planning`):
- `continuous_events` — detected events for continuous planning
- `plan_update_proposals` — proposed plan updates with affected entities
- `context_deliveries` — delivered context at various levels

### MCP Tools

- `plan_ai.continuous_status` — get continuous planning status for a project
- `plan_ai.continuous_events` — list recent continuous planning events
- `plan_ai.continuous_proposals` — list plan update proposals
- `plan_ai.continuous_context` — generate context at a specified level

## Consequences

### Positive
- Plans stay fresh as the project evolves
- Automatic detection reduces human oversight burden
- Approval gate prevents unwanted plan mutations
- Multi-level context generation serves different audiences (executive, planner, implementer)

### Negative
- False positives from event detection may generate noise
- Proposal approval adds process overhead for trivial updates
- Context generation quality depends on completeness of project data
