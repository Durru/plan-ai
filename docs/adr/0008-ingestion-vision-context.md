# ADR 0008: Ingestion, Vision, and Approved Context

## Status

Accepted

## Context

Plan-AI needs a bounded front half of the planning flow before advanced research
or planning begins. User input must be preserved, normalized, turned into a
vision draft, and only then promoted into durable context after approval.

## Decision

Introduce three separate internal packages:

- `internal/ingestion`: captures raw input and normalized ingested sources.
- `internal/vision`: extracts incomplete vision drafts from ingested sources.
- `internal/context`: stores only explicitly approved context.

Persistence is additive through migration `0008_ingestion_vision_context`, with
runtime SQL inline in `internal/store/store.go` and a mirror SQL file under
`internal/migrations/project/`.

## Consequences

- Ingestion is factual capture, not interpretation.
- Vision drafts may be incomplete and must not invent missing details.
- Approved context is reusable without depending on chat history.
- Advanced research, orchestration, workflow, and model strategy remain outside
  this decision.
