# ADR-0016: MCP Server

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 19

## Context

AI agents need a way to invoke Plan-AI functionality from their own
tool-calling workflows. The Model Context Protocol (MCP) provides a
standard for exposing tools that agents can discover and call.

## Decision

Implement a lightweight MCP-compatible server within Plan-AI, exposing
core planning operations as callable tools.

Key design choices:
1. **In-process** — no separate daemon process for MVP
2. **JSON Schema validation** — arguments validated before dispatch
3. **Audit trail** — every tool execution is recorded
4. **13 default tools** — covering the complete planning lifecycle
5. **CLI interface** — `mcp-server` binary for list-tools and call-tool

## Consequences

### Positive
- AI agents can interact with Plan-AI in a standardised way
- Each tool has documented schema and validation
- Execution audit provides debugging and replay capability
- No external dependencies for MCP transport

### Negative
- Not a full MCP server with streaming/SSE transport
- Tool discovery requires running the binary
- Stub handlers until real services are wired

## Alternatives Considered

**Full HTTP server with SSE**: More complex, adds infra overhead for MVP.  
**gRPC service**: Too heavy for a CLI tool's integration surface.
