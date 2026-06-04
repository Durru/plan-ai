# ADR-0019: Agent System

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 21

## Context

Plan-AI needs an agentic layer that can process user messages, detect intent, route
to the appropriate workflow, delegate sub-tasks, and build structured responses. This
enables the system to act as an intelligent planning assistant rather than a passive
tool collection.

## Decision

Implement an Agent System as an internal package (`internal/agent/`) with these
components:

1. **Intent Detection** — classify user messages into known intent kinds
   (create, research, approve, analyze, etc.) using keyword/pattern matching
2. **Workflow Selection** — map detected intents to workflow types
3. **Capability Selection** — map intents to required capabilities
4. **Routing** — orchestrate intent → workflow → response pipeline
5. **Delegation** — spawn sub-agent jobs for long-running or specialized tasks
6. **Context Loading** — load project context (plans, status, events) for the agent
7. **Response Building** — construct structured responses with status, approvals, errors

### Database Schema

New tables in migration 0019 (`agent_system`):
- `agent_runs_v2` — enhanced agent run records with intent, status, response
- `agent_messages` — conversation messages within agent runs
- `agent_delegated_jobs` — delegated sub-agent jobs with capability routing

Compatibility views for definitive schema alignment:
- `agent_runs` — view over `agent_runs_v2`
- `subagent_outputs` — view over `agent_delegated_jobs`

### MCP Tools

- `plan_ai.agent_process` — process a user message through the agent system
- `plan_ai.agent_runs` — list recent agent runs for a project

## Consequences

### Positive
- Structured pipeline for handling diverse user intents
- Sub-agent delegation enables parallel or long-running work
- Full audit trail via agent runs and messages tables
- Clear separation of concerns across intent, routing, and execution

### Negative
- Intent detection is pattern-based, not ML — limited for ambiguous inputs
- Delegation adds complexity for simple requests
- Agent responses are structured JSON — may feel less conversational
