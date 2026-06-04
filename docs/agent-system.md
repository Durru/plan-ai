# Agent System

Phase 21 adds the Plan-AI agent system.

The agent is an orchestration boundary, not an autonomous coding agent. It does not implement code, keep private memory, or contain independent intelligence. It receives a user message, detects intent, loads the minimum derived/project context, selects a workflow and capability, delegates temporary jobs through the orchestrator layer, and returns a clear response.

## Supported intents

- `create_project`
- `create_master_plan`
- `create_specific_plan`
- `research_topic`
- `update_plan`
- `change_request`
- `implementation_help`
- `project_status`
- `approve`
- `reject`
- `validate`
- `next_task`
- `unknown`

## Delegated jobs

Delegated jobs are temporary records, not persistent sub-agents or resident processes.

Supported job types:

- `vision_job`
- `research_job`
- `planning_job`
- `validation_job`
- `impact_job`
- `context_job`

## Persistence

The runtime migration uses the next available store IDs because earlier Phase 18-20 compatibility work already used `0018`.

- `0019_agent_system` creates the physical Phase 21 tables.
- `0021_agent_continuous_compatibility` adds contract-compatible views such as `delegated_jobs` and `agent_responses`.
