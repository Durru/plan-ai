# ADR 0012: Model Strategy

## Decision
Add a model strategy layer that manages LLM model profiles, prompt contracts, and output schemas with selection, retry, and budget logic.

## Rationale
Plan-AI needs a centralised way to configure and select LLM models, enforce prompt contracts, and validate output schemas without scattering provider-specific logic across the codebase.

## Consequences
- Model profiles are stored in the project database and managed via `ModelProfileRepository`.
- Prompt contracts and output schemas enable typed, validated LLM interactions.
- The selector picks profiles by capability and cost budget; retry logic handles transient failures.
- Phase 16+ (orchestrator) depends on this layer for model-aware job dispatch.
