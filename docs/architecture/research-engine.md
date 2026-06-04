# Research Engine Architecture

## Purpose

The Research Engine is Plan-AI's local, deterministic, AI-free tracker of structured research investigations. It captures what a project has researched — findings, sources, conclusions, tags — and how that research maps to knowledge objects, so that the Planner and other phases can consume approved research without re-doing the investigation.

Research data lives in the same `project.db` introduced in Phase 2. It is local to the project and never leaves the project store.

## Scope of This Phase

This phase introduces a structured research entity, a deterministic classifier, sub-tables for findings/sources/conclusions/tags/knowledge-links, a gateway-based approval checker, a service layer, a SQLite adapter, and a CLI surface for managing research entries. It does not introduce:

- Web crawling or source fetching.
- AI-generated findings, sources, or conclusions.
- Skill Intelligence.
- A Planner.
- A Context Engine.
- MCP or OpenCode integration.
- Engram integration.
- Real agents, sub-agents, or AI calls.
- Embeddings, vector databases, or semantic search.
- FTS5 virtual tables, generated columns, or triggers.
- A web UI.

This phase is intentionally minimal and deterministic. Smart research consumption comes later.

## Research Entry

A research entry represents an investigation into a specific topic. It has the following fields:

| Field        | Type               | Notes                                                               |
|-------------|--------------------|---------------------------------------------------------------------|
| `id`        | string             | Stable identifier of the form `research_<32 hex chars>`.            |
| `topic`     | string             | Short, human-readable title. Required.                              |
| `category`  | `ResearchCategory` | Re-uses Knowledge Base's 15 fixed categories. Default `general`.    |
| `summary`   | string             | One-sentence summary of the investigation.                          |
| `status`    | `ResearchStatus`   | Lifecycle stage: `draft`, `in-review`, `approved`, `rejected`, `archived`. |
| `confidence`| int [0-100]        | Self-rated confidence in the research quality.                      |
| `created_at`| RFC3339 timestamp  | UTC.                                                                |
| `updated_at`| RFC3339 timestamp  | UTC. Updated on status changes.                                     |

## Categories

Research uses the same 15 fixed categories as the Knowledge Base: `database`, `authentication`, `billing`, `frontend`, `backend`, `security`, `deployment`, `architecture`, `testing`, `mcp`, `agents`, `ai`, `devops`, `integration`, `general`.

`general` is the fallback category.

## Lifecycle

A research entry moves through one of five statuses:

1. `draft` — created but not yet reviewed. Findings, sources, and conclusions may be empty.
2. `in-review` — ready for human review. Not yet enforced by Phase 6.
3. `approved` — has at least one finding, one source, and one conclusion. Gate-checked by `ApprovalChecker`.
4. `rejected` — review determined the research is not useful.
5. `archived` — no longer active; kept for history and search.

The approval gate (`CanApprove`) is the only automated status transition. It requires existence of ≥1 finding, ≥1 source, and ≥1 conclusion. It does not evaluate quality — that is a human responsibility.

## Classification

`Classify(topic)` is identical to the Knowledge Base classifier. It is a deterministic, keyword-based function returning the first matching category. No fuzzy matching, no ML model, no embeddings, no LLM call.

## Sub-Entities

### Findings

A finding is a discrete discovery made during research:

| Field        | Type    | Notes                                        |
|-------------|---------|----------------------------------------------|
| `id`        | string  | Stable identifier of the form `finding_...`. |
| `research_id`| string | Foreign key to `research_entries(id)`.       |
| `title`     | string  | Short label. Required.                       |
| `content`   | string  | Detailed description.                        |
| `importance`| int [1-5]| Subjective importance rating. Default 3.    |
| `created_at`| RFC3339 | UTC.                                         |

Ordered by `importance DESC` in list views.

### Sources

A source is a reference that informed the research:

| Field        | Type                | Notes                                        |
|-------------|---------------------|----------------------------------------------|
| `id`        | string              | Stable identifier of the form `source_...`.  |
| `research_id`| string             | Foreign key to `research_entries(id)`.       |
| `title`     | string              | Short label. Required.                       |
| `url`       | string              | Optional URL.                                |
| `source_type`| `ResearchSourceType`| One of `manual`, `documentation`, `article`, `repository`, `specification`, `benchmark`, `internal`. |
| `created_at`| RFC3339             | UTC.                                         |

Ordered by `created_at` in list views.

### Conclusions

A conclusion is a synthesized insight drawn from findings and sources:

| Field        | Type    | Notes                                        |
|-------------|---------|----------------------------------------------|
| `id`        | string  | Stable identifier of the form `conclusion_...`. |
| `research_id`| string | Foreign key to `research_entries(id)`.       |
| `content`   | string  | The conclusion text. Required.               |
| `confidence`| int [0-100]| Self-rated confidence. Default 0.          |
| `created_at`| RFC3339 | UTC.                                         |

Ordered by `confidence DESC` in list views.

### Tags

Tags are short, lowercase labels. The same normalization rules apply as in the Knowledge Base. Duplicates on the same research entry are silently deduplicated (`UNIQUE(research_id, tag)` + `ON CONFLICT DO NOTHING`).

### Knowledge Links

A knowledge link connects research to an approved knowledge object:

| Field          | Type    | Notes                                        |
|---------------|---------|----------------------------------------------|
| `id`          | string  | Stable identifier of the form `rlink_...`.   |
| `research_id` | string  | Foreign key to `research_entries(id)`.       |
| `knowledge_id`| string  | Foreign key to `knowledge_objects(id)`.      |
| `created_at`  | RFC3339 | UTC.                                         |

Duplicates are silently deduplicated (`UNIQUE(research_id, knowledge_id)` + `ON CONFLICT DO NOTHING`). In Phase 6, links are recorded but not validated against the knowledge store. Cross-store validation is the responsibility of future phases.

## Storage

Research data lives in `project.db` only. The Phase 6 migration `0005_research_engine` is additive and idempotent:

- adds `category` and `status` columns to `research_entries` with safe `DEFAULT` values;
- creates `research_findings`, `research_sources`, `research_conclusions`, `research_tags`, and `research_knowledge_links` with `UNIQUE` constraints and indexes;
- uses `ON CONFLICT DO NOTHING` for tag and knowledge-link inserts;
- legacy `source` and `conclusion` columns on `research_entries` are kept (not dropped) to avoid a full table rebuild.

The global store is not touched.

## Service

The `internal/research` package owns a `Repository` interface and a `Service` that enforce invariants. The service is the only entry point that mutations should use. The service:

- generates stable IDs;
- clamps `confidence` to [0, 100];
- fills in `category` from the classifier when empty;
- normalizes tags;
- validates that the research entry exists before adding sub-entities;
- enforces the approval gate (`CanApprove`) before transitioning to approved;
- refreshes `updated_at` on status changes;
- provides `Describe` to load a research entry with all its sub-entities in one call.

The store layer implements the `Repository` interface. A compile-time assertion (`var _ research.Repository = (*ResearchRepository)(nil)`) enforces the contract.

## CLI

`plan-ai` exposes a `research` subcommand group with 11 actions:

- `research add --topic <topic> [--category ...] [--summary ...] [--confidence ...] [--tag ...]` — create a new research entry.
- `research list` — list all research entries with ID, topic, and status.
- `research show <id>` — print one research entry plus its findings, sources, conclusions, tags, and knowledge links.
- `research search <query>` — substring search on topic and summary.
- `research approve <id>` — approve a research entry (gate-checked: requires findings + sources + conclusions).
- `research reject <id>` — reject a research entry.
- `research archive <id>` — archive a research entry.
- `research finding <id> <title> [content]` — add a finding to an entry.
- `research source <id> <title> [url] [--type ...]` — add a source to an entry.
- `research conclusion <id> <content>` — add a conclusion to an entry.
- `research link <id> <knowledge-id>` — link research to a knowledge object.

`plan-ai dev seed-research` seeds two sample entries (`LLM Token Optimization`, `SQLite Performance Limits`) with findings, sources, and conclusions.

`plan-ai status` prints a `Research:` block with `total`, per-status counts, and findings/sources/conclusions counts when a project database exists.

## What This Phase Does Not Promise

- It does not crawl the web or fetch sources automatically.
- It does not generate findings, sources, or conclusions.
- It does not validate knowledge links against the knowledge store.
- It does not provide a UI.
- It does not integrate with the Planner, Context Engine, or any other phase.

All of that belongs to later phases.
