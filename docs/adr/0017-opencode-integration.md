# ADR-0017: OpenCode Integration

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 20

## Context

Plan-AI is developed for and deployed via OpenCode. Users may already
have OpenCode configuration in their projects. Plan-AI should coexist
with OpenCode rather than conflict.

## Decision

1. **Read-only detection** — Plan-AI detects OpenCode configuration but
   never writes to it.
2. **MCP tool bridge** — Plan-AI exposes its tools for OpenCode to call
   via MCP.
3. **Three modes** — standalone (no MCP), tool (default), hybrid.
4. **Doctor checks** — integration health checks for troubleshooting.
5. **Separate tool registry** — tracks which tools OpenCode has invoked.

## Consequences

### Positive
- Safe coexistence — Plan-AI cannot corrupt OpenCode config
- Flexible deployment — can run standalone or as an OpenCode tool
- Diagnostic health checks for integration issues
- Read-only by default prevents accidental changes

### Negative
- Plan-AI cannot auto-configure OpenCode (intentional constraint)
- Detection limited to filesystem scan; runtime detection is out of scope

## Alternatives Considered

**Write OC config from PA**: Risk of corrupting user's OpenCode setup.  
**Deep OpenCode hooking**: Too coupled, maintenance burden.
