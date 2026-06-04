# ADR 0009: Research and Knowledge Engines

## Decision
Add project-local research and knowledge registries backed by additive SQLite migrations.

## Rationale
Plan-AI must reuse prior investigation instead of re-researching the same topic.

## Consequences
- Research and knowledge are persisted separately but linked by IDs.
- No LLM, network, or orchestration behavior is introduced in this phase.
