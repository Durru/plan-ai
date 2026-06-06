// Package mcp implements Plan-AI's MCP server — tool definitions,
// handlers, protocol (stdio/HTTP), SDK integration, and validation.
// It is the transport adapter for Plan-AI services, not an alternate
// business logic path (Phase 15).
//
// Main types: ToolDefinition, ToolDependencies, Server, SDKServer.
// Main entry: RegisterSDKDefaultTools for tool registration.
package mcp
