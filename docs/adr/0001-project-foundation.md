# ADR 0001: Project Foundation

## Status

Accepted

## Context

Plan-AI needs a clean Phase 0 foundation before any planner, research, storage, MCP, skills, integrations, or business logic is implemented.

## Decision

Use a Go modular monolith with a Cobra CLI, private `internal/` packages, reserved `pkg/`, and documented global and project `.plan-ai` directories.

Plan-AI does not implement. Plan-AI prepares implementation.

## Consequences

- The CLI can expose stable command names early without implementing behavior.
- Future modules have explicit package homes.
- Configuration paths are standardized before persistence or integrations are added.
- Business logic remains intentionally absent from Phase 0.
