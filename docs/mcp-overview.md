# MCP Overview

MCP is a future integration layer for Plan-AI, not the core architecture.

## Current state

`internal/mcp` is a placeholder only. No stdio MCP server is active yet.

## Direction

Phase 19 will implement the definitive MCP layer after domain, store, context, change, and workflow layers exist.

Expected MCP responsibilities:

- expose approved context
- expose plans/tasks/status
- expose change and impact information
- serve right-sized context to external tools

## Non-goals before Phase 19

- no real MCP server
- no OpenCode MCP config mutation
- no tool execution through MCP
- no agent-specific assumptions in the core model
