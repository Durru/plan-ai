# ADR 0005: Project Knowledge Base

## Status

Accepted

## Context

Plan-AI's later phases — Planner, Research Engine, Context Engine, Skill Intelligence, and integrations — all need a stable, project-local model of "what the project already knows." Without it, every phase re-derives the same facts: which database is used, which auth strategy is approved, which billing provider is preferred, which technologies are forbidden, which decisions are settled, which patterns are reusable.

Phase 5 introduces a local, deterministic Knowledge Base that captures reusable project knowledge in `project.db`. The Knowledge Base is the substrate that later phases will read from, but it is intentionally simple in this phase: it does not generate, infer, or curate knowledge on its own. A human (or future automated phase) must add knowledge objects explicitly.

## Decision

Implement a deterministic, local, AI-free Knowledge Base before any phase that would consume it.

Concretely:

- Add a `KnowledgeObject` entity with stable ID, topic, category, summary, content, confidence, source type, status, reuse count, and timestamps.
- Define 15 fixed categories, 4 lifecycle statuses, 4 source types, 4 relation types, and 4 reference types. The set is closed and stable.
- Classify topics deterministically by ordered keyword rules; fall back to `general`. No ML, no embeddings, no LLM call.
- Persist knowledge objects, tags, relations, and references in `project.db` via a Phase 5 migration that is additive and idempotent with Phases 2–4.
- Enforce invariants in a `Service` layer (`internal/knowledge`) over a `Repository` interface. The SQLite adapter satisfies the interface and is verified by a compile-time assertion.
- Expose a `plan-ai knowledge` CLI subcommand group plus a `dev seed-knowledge` helper for local validation.
- Add a `Knowledge:` block to `plan-ai status` with `total`, per-status counts, and a `reused` sum.

The Knowledge Base does not generate, infer, deduplicate semantically, validate references against other stores, expose a UI, or talk to AI. Those capabilities are explicitly deferred.

## Rationale

A deterministic local Knowledge Base first is the correct foundation because it is:

- testable: classification rules, normalization, and invariants are unit-tested in isolation;
- reproducible: the same topic, tags, and relations always produce the same stored state;
- cheap: no LLM calls, no network access, no embeddings, no vector database;
- offline-first: it works in local projects without cloud services;
- safe: it stores human-curated strings only — no code, no secrets, no prompts;
- composable: a `Repository` interface lets future storage engines (Postgres, in-memory, remote) replace SQLite without touching service code.

A closed enum set for categories, statuses, source types, relation types, and reference types keeps the model auditable. Every stored value can be reasoned about by reading the code, not by running an LLM.

`UNIQUE` constraints plus `ON CONFLICT DO NOTHING` make tag, relation, and reference inserts idempotent. Combined with a migration that uses `ALTER TABLE ... ADD COLUMN ... NOT NULL DEFAULT ...`, this lets Phase 5 land without dropping or rewriting existing data from Phases 2–4.

Reuse is tracked as a counter, not a list, to keep Phase 5 minimal. The audit trail of "who reused what, when" belongs to later phases once the consumers exist.

## Consequences

Later phases can read from the Knowledge Base without re-deriving facts:

- Planner can pre-fill phase templates with approved architecture, framework, and integration knowledge.
- Research can mark a topic as "already known" and skip it.
- Context Engine can pre-load high-confidence knowledge into prompts.
- Skill Intelligence can correlate knowledge with detected skills.
- A future agent layer can suggest reuse events when a phase clearly uses existing knowledge.

The Knowledge Base's strict scope (no auto-generation, no semantic search, no cross-store validation, no UI) keeps Phase 5 small and verifiable. Adding any of those capabilities belongs to a dedicated later phase, not a stealth extension of Phase 5.

## Alternatives Considered

### AI-curated Knowledge Base

Rejected for this phase. AI curation would be non-deterministic, harder to test, and premature before Plan-AI has a stable local project model and a real consumer. The Knowledge Base is the substrate that an AI layer would later write to, not the place where AI lives.

### Embeddings + vector search

Rejected. Embeddings require a model, a vector store, and operational complexity. Phase 5 only needs substring search across `topic`, `summary`, and `content`. Substring search is testable, fast, and good enough for a closed, human-curated set.

### Knowledge as part of the Domain Model

Rejected. Knowledge and Domain entities have different lifecycles. Domain entities (plans, phases, tasks, decisions, research, validations, snapshots) are project execution artifacts; knowledge is reusable context. Mixing them would force the Domain Model to carry lifecycle semantics (status, reuse count) it does not need.

### External knowledge store (Notion, Confluence, GitHub Wiki)

Deferred. Local-first SQLite keeps the Knowledge Base offline, private, and version-controllable. An external sync layer can be added later without changing the on-disk schema.

### FTS5 virtual tables

Deferred. Substring `LIKE` search is sufficient for a closed, human-curated set in Phase 5. FTS5 can be added when the corpus grows large enough that substring search becomes a bottleneck.
