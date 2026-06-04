# ADR 0018: Agent System

## Status

Accepted

## Context

Plan-AI needs a user-facing agent boundary after the workflow, orchestrator, MCP, and OpenCode layers exist. The agent must not become a hidden implementation engine or independent memory system.

## Decision

Add Phase 21 Agent System as a thin orchestration layer:

- detect user intent;
- load minimum context;
- select workflow, capability, and model strategy;
- create temporary delegated jobs;
- return clear responses;
- request approval when needed.

The agent does not implement code, maintain private memory, or run resident sub-agents.

## Persistence

The requested ADR number is `0018`, but migration `0018` was already used by the Phase 18-20 compatibility layer. Therefore Phase 21 runtime persistence uses `0019_agent_system`, while this ADR remains `0018-agent-system` to match the roadmap phase documentation.

## Consequences

Plan-AI now has a user-facing orchestration layer without violating the separation between LLM intelligence and Plan-AI state/workflow coordination.
