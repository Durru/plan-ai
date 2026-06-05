# MCP Server

The MCP (Model Context Protocol) Server exposes Plan-AI functionality as callable tools that AI agents can invoke. Each tool has a name, description, schema-like argument contract, validation, and a handler.

## Architecture

Located in `internal/mcp/` and `cmd/mcp-server/`:

- **Server** — tool registry, execution, validation, and audit.
- **Types** — `ToolDefinition`, `ToolResult`, `JSONSchema`, `RunRecord`.
- **Validation** — argument validation against schema-like structs with type and enum checks.
- **Tools** — default tool set registration.
- **Protocol server** — `cmd/mcp-server/main.go` serves MCP JSON-RPC over stdio when run with no arguments or `stdio`.
- **CLI compatibility** — `list-tools` and `call-tool` remain available for local validation.

## Default tools

Core tools:

- `plan_ai.init_project`
- `plan_ai.project_status`
- `plan_ai.create_master_plan`
- `plan_ai.create_specific_plan`
- `plan_ai.research_topic`
- `plan_ai.approve_plan`
- `plan_ai.reject_plan`
- `plan_ai.analyze_impact`
- `plan_ai.get_next_task`
- `plan_ai.mark_task_done`
- `plan_ai.create_snapshot`
- `plan_ai.list_plans`
- `plan_ai.list_tasks`

Phase 21/22 tools:

- `plan_ai.agent_message`
- `plan_ai.agent_status`
- `plan_ai.continuous_status`
- `plan_ai.propose_plan_update`
- `plan_ai.approve_plan_update`
- `plan_ai.reject_plan_update`
- `plan_ai.get_context_level`

Backward-compatible aliases may also exist for earlier internal names such as `plan_ai.agent_process`, `plan_ai.agent_runs`, and `plan_ai.continuous_context`.

## CLI usage

```sh
plan-ai-mcp-server list-tools
plan-ai-mcp-server call-tool plan_ai.project_status '{"project_root": "/path/to/project"}'
plan-ai-mcp-server call-tool plan_ai.agent_message '{"message": "What is next?"}'
plan-ai-mcp-server call-tool plan_ai.get_context_level '{"level": "L0_Executive"}'
```

For OpenCode or any MCP client, configure the local command as:

```json
{
  "mcp": {
    "plan-ai": {
      "type": "local",
      "enabled": true,
      "command": ["plan-ai-mcp-server"]
    }
  }
}
```

## Tool execution lifecycle

1. Tool name is resolved from the registry.
2. Arguments are validated against the schema-like contract.
3. The handler function executes through service-layer dependencies.
4. The result and execution record are available for audit.
