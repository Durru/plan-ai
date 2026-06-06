# ADR 0023: Conversation Gateway

**Status:** Accepted
**Date:** 2026-06-06
**Phase:** 4

## Context

Plan-AI had two separate user interaction paths: a CLI `agent process` command and an MCP `plan_ai.agent_message` tool. The CLI worked correctly, constructing an `agent.Service` with intent detection, routing, context loading, delegation, and run persistence. The MCP handler was a stub (`"Agent processing is a stub. Connect real agent services."`).

Phase 4 unified them through a single `conversation.Gateway`.

## Decision

1. **Single entry point**. Both CLI and MCP route through `conversation.Gateway.ProcessMessage(projectID, message)`.
2. **Intent-based routing**. The gateway wraps `agent.Service` which detects intent (create_master_plan, research_topic, analyze_project, etc.) and routes to the appropriate workflow.
3. **Run persistence**. Every message pair (user + agent) is persisted to `agent_runs_v2` and `agent_messages` tables.
4. **Conversation intents**. The gateway supports project analysis, product creation, next step, database plan, and impact analysis in addition to the original 12 intents.
5. **Response guidance**. Responses carry `suggested_next_action` instead of command instructions.

## Rationale

Two separate code paths that produce different results for the same user input is a correctness bug. The conversation gateway makes CLI and MCP behavior identical.

Persisting runs and messages gives Plan-AI a durable conversation history independent of any model session. This is a prerequisite for memory recording (Phase 8).

## Consequences

- `plan_ai.agent_message` no longer returns stub text.
- `plan-ai agent process <message>` and the MCP equivalent return identical structured responses.
- Conversation runs are queryable via `plan-ai agent runs`.
- The planning guard (Phase 5) is wired into the gateway to block plan creation without approved intent.
- All existing agent tests pass; Phase 4+ tests verify the gateway.
