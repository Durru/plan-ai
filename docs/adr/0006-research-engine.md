# ADR 0006: Research Engine

## Status

Accepted

## Context

Plan-AI's project planning flow needs a structured way to track research — investigations into technologies, patterns, architectures, and approaches that inform planning decisions. Before Phase 6, the only research persistence was a flat `research_entries` table with `topic`, `source`, `summary`, and `conclusion` columns — no findings, no sources, no conclusions, no tags, no cross-references to the Knowledge Base.

Without a structured research model, every phase that needs research context would either re-derive it or store it ad-hoc. The Knowledge Base (ADR 0005) stores reusable, approved knowledge. Research is the *process* of generating that knowledge — provisional findings, evaluated sources, tentative conclusions — before it is mature enough to promote to the Knowledge Base.

Phase 6 introduces a self-contained Research Engine alongside the existing data model, using the same patterns as the Knowledge Base: deterministic classification, gateway-based lifecycle gates, idempotent storage, and a `Repository` interface with a SQLite adapter.

## Decision

Implement a deterministic, local, AI-free Research Engine as Phase 6, using the same architectural patterns as the Knowledge Base.

Concretely:

- Replace the flat `ResearchEntry` struct in the domain model with a rich entity that carries `Category` (re-uses Knowledge Base's 15 categories) and `Status` (5 values: draft, in-review, approved, rejected, archived).
- Extend `research_entries` with `category` and `status` columns via an idempotent `ALTER TABLE` migration (0005), keeping legacy `source` and `conclusion` columns for backward compatibility.
- Add five new sub-tables in migration 0005: `research_findings`, `research_sources`, `research_conclusions`, `research_tags`, and `research_knowledge_links`.
- Define a `research.Repository` interface in the `research` package with 16 methods covering all CRUD + summary. The store package provides a concrete `*ResearchRepository` that satisfies the interface, verified by a compile-time assertion.
- Classify research topics deterministically using the same keyword rules as the Knowledge Base (`research.Classify`); fall back to `general`.
- Implement an `ApprovalChecker` gateway that requires at least one finding, one source, and one conclusion before transitioning a research entry from draft to approved.
- Expose a `plan-ai research` CLI subcommand group with 11 sub-commands: `add`, `list`, `show`, `search`, `approve`, `reject`, `archive`, `finding`, `source`, `conclusion`, `link`.
- Add a `dev seed-research` helper that seeds two sample research entries with findings, sources, and conclusions.
- Add a `Research:` block to `plan-ai status` with total, per-status counts, and findings/sources/conclusions counts.
- Remove the old `store.ResearchRepository` from `domain_repositories.go` and replace its usage with the new `*ResearchRepository` from `research_repositories.go`.

The Research Engine does not crawl the web, call LLMs, generate content, train models, or integrate with external APIs. Those capabilities are explicitly deferred.

## Rationale

A gateway-based approval check (findings + sources + conclusions) before approving research enforces a quality bar without requiring human review in the loop. The checker verifies existence only — it does not evaluate quality, accuracy, or relevance. Those judgements belong to human reviewers once a CLI or UI layer invites them.

Re-using Knowledge Base categories ensures that research and knowledge share the same vocabulary, making cross-references (`research_knowledge_links`) meaningful without schema translation.

Legacy `source` and `conclusion` columns on `research_entries` are kept (not dropped) because dropping columns in SQLite requires a full table rebuild — an unnecessary cost when the new sub-tables supersede those fields. The new code writes empty strings to the legacy columns and ignores them on read.

Separating the `research` package from `store` behind a `Repository` interface follows the same contract as the Knowledge Base: the research package owns its storage contract, and the store package implements it. A compile-time assertion (`var _ research.Repository = (*ResearchRepository)(nil)`) prevents interface drift.

## Consequences

- The Planner phase can query research by status and category to find approved research that should inform planning decisions.
- The Knowledge Base can track which research entries were promoted to knowledge objects via the `research_knowledge_links` table.
- A future Phase can add a `promote` command that copies an approved ResearchEntry's findings+conclusions into a KnowledgeObject, optionally linking them.
- The CLI surface grows by 11 research sub-commands plus 1 seed-research command.
- The database gains 5 new tables and 7 new indexes. Migration count moves from 4 to 5.
- The test suite grows by 13 new test functions covering the research repository.

## Alternatives Considered

### Research as part of the Knowledge Base

Rejected. Research is provisional by nature; Knowledge is curated and approved. Mixing them would force research entries to carry Knowledge lifecycle semantics (reuse count, reference types) that they do not need, and would prevent the approval gate from being a research-specific invariant.

### Research as free-form text

Rejected. The flat model (single `source`, single `conclusion`) was the pre-Phase-6 state. Structured findings/sources/conclusions enable search, filtering, and the approval gate that free-form text cannot support deterministically.

### AI-generated research summaries

Deferred. An AI layer could auto-generate findings, sources, and conclusions from a topic string. That capability belongs to a dedicated Phase after the research storage model is stable and tested.

### Web crawling for sources

Deferred. Fetching and validating sources from the web is a separate capability that depends on the MCP layer or an external integration. Phase 6 treats sources as human-entered metadata only.
