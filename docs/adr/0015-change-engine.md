# ADR-0015: Change Engine

**Status:** Approved  
**Date:** 2026-06-03  
**Phase:** 18

## Context

Planning entities (visions, requirements, plans, etc.) evolve over time.
When one entity changes, it can invalidate others. Without a change engine,
users have no systematic way to know what needs re-review after a change.

## Decision

Implement an invalidation-based change engine with:

1. **Change Types** — enumerated, typed event categories
2. **Invalidation Rules** — matrix mapping change→affected entity types
3. **Entity States** — tracked validity status per entity
4. **Impact Analysis** — automatic determination of blast radius
5. **Snapshots** — point-in-time captures before/after changes
6. **Audit Trail** — full change event history

## Consequences

### Positive
- Clear traceability from changes to affected entities
- Automatic impact analysis reduces human error
- Snapshots enable rollback and comparison
- Entity states guide the user on what needs attention

### Negative
- Requires discipline to register changes rather than modifying state directly
- Invalidation rules are static; complex projects may need customization

## Alternatives Considered

**Event sourcing**: Too heavy for a planning tool context.  
**Optimistic locking**: Doesn't provide impact visibility.  
**Manual tracking**: Error-prone and burdensome.
