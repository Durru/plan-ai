# ADR-0018: Compatibility Layer

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 18-20 Compatibility

## Context

Earlier phases used physical table names with `_v2` suffixes (e.g., `snapshots_v2`,
`context_views_v2`) or names that differed from the definitive storage schema documented
in `docs/storage-schema.md`. As the project evolves, consumers (tests, MCP tools, external
scripts) need stable, predictable table and view names that match the definitive schema.

## Decision

Introduce a compatibility layer using SQL `CREATE VIEW IF NOT EXISTS` and additional
tables to bridge the gap between physical implementation names and the definitive schema:

1. **Views over physical tables** — e.g., `tool_runs` over `mcp_runs`,
   `tool_audit` joining `mcp_tools` and `mcp_runs`
2. **New tables** for registry entries not yet covered by earlier migrations:
   `provider_registry`, `skill_registry`
3. **All additions are additive** — existing tables are never altered or renamed
4. **Definitive schema** in `docs/storage-schema.md` reflects the canonical names

## Consequences

### Positive
- Stable names for consumers regardless of internal table naming
- Backward compatible — no data migration required
- Clear separation between physical implementation and logical schema

### Negative
- Additional mental overhead — developers must check both physical and view names
- Views may degrade performance for very large datasets (acceptable for MVP)
