# Knowledge Base Architecture

## Purpose

The Knowledge Base is Plan-AI's local, deterministic, AI-free store of reusable project knowledge. It captures what a project has already learned about its stack, conventions, decisions, technologies, integrations, and recurring patterns so that later phases can reuse it instead of rediscovering it.

Knowledge is stored in the same `project.db` introduced in Phase 2. It is local to the project and never leaves the project store.

## Scope of This Phase

This phase introduces the canonical knowledge object, a deterministic classifier, a normalized model for tags, relations, and references, a service layer that enforces invariants, a SQLite adapter, and a CLI surface for managing knowledge objects. It does not introduce:

- A Research Engine.
- A Skill Intelligence layer.
- A Planner.
- A Context Engine.
- MCP or OpenCode integration.
- Engram integration.
- Real agents, sub-agents, or AI calls.
- Embeddings, vector databases, or semantic search.
- FTS5 virtual tables, generated columns, or triggers.
- A web UI.

This phase is intentionally minimal and deterministic. Smart reuse comes later.

## Knowledge Object

A knowledge object is the smallest unit of reusable knowledge Plan-AI tracks. It has the following fields:

| Field         | Type                  | Notes                                                                  |
| ------------- | --------------------- | ---------------------------------------------------------------------- |
| `id`          | string                | Stable identifier of the form `knowledge_<32 hex chars>`.              |
| `topic`       | string                | Short, human-readable title. Required.                                 |
| `category`    | `KnowledgeCategory`   | One of 15 fixed categories. Default `general`.                        |
| `summary`     | string                | One-sentence summary.                                                  |
| `content`     | string                | Optional longer description.                                           |
| `confidence`  | number in [0, 1]      | Author self-rated confidence.                                          |
| `source_type` | `KnowledgeSourceType` | One of `manual`, `research`, `imported`, `generated`.                  |
| `status`      | `KnowledgeStatus`     | Lifecycle stage: `draft`, `reviewed`, `approved`, `archived`.          |
| `reuse_count` | int                   | Incremented every time a downstream phase records reuse.              |
| `created_at`  | RFC3339 timestamp     | UTC.                                                                   |
| `updated_at`  | RFC3339 timestamp     | UTC. Updated on every write and on reuse.                              |

## Categories

The 15 fixed categories are: `database`, `authentication`, `billing`, `frontend`, `backend`, `security`, `deployment`, `architecture`, `testing`, `mcp`, `agents`, `ai`, `devops`, `integration`, `general`.

`general` is the fallback category used when a topic cannot be classified deterministically.

## Lifecycle

A knowledge object moves through one of four statuses:

1. `draft` — captured but not yet validated by the project owner.
2. `reviewed` — read by a human or automated check; still not approved.
3. `approved` — trusted enough to be reused automatically by downstream phases.
4. `archived` — no longer in active use; kept for history and search.

Only `approved` objects are considered "trusted knowledge" by future phases.

## Classification

`Classify(topic)` is a deterministic, keyword-based classifier that inspects the topic string and returns the first matching category. It is implemented as ordered `strings.Contains` rules. The order is fixed and stable. When no rule matches, the result is `general`.

The classifier has no fuzzy matching, no ML model, no embeddings, and no LLM call. Two projects with the same topic always get the same category.

## Tags

Tags are short, lowercase, alphanumeric-with-dashes labels. They are normalized at write time:

- trimmed of leading and trailing whitespace;
- lowercased;
- kept as-is otherwise (a small set of safe characters is allowed: letters, digits, dashes, underscores, dots, colons, slashes).

Duplicate tags on the same knowledge object are silently deduplicated. The `internal/knowledge` package exposes `NormalizeTag` so tests and callers can apply the same rule.

## Relations

Knowledge objects may be linked to other knowledge objects with a typed relation. The four relation types are:

- `related` — loosely connected.
- `depends_on` — the source depends on the target.
- `alternative_to` — the source is a replacement option for the target.
- `extends` — the source builds on the target.

Relation inserts are idempotent (`UNIQUE(source_id, target_id, relation_type)` + `ON CONFLICT DO NOTHING`). A self-link is rejected by the service. A relation whose target does not exist is rejected by the service.

## References

Knowledge objects may reference objects in the domain model — plans, decisions, research, or technologies. The four reference types are: `plan`, `decision`, `research`, `technology`. Reference inserts are idempotent (`UNIQUE(knowledge_id, reference_type, reference_id)` + `ON CONFLICT DO NOTHING`).

References in this phase are recorded but not validated against the referenced object. A reference to a non-existent object is allowed. Cross-store validation is the responsibility of future phases.

## Reuse Tracking

Every reuse event increments `reuse_count` and refreshes `updated_at`. A reuse event is recorded by the service (`ReuseKnowledge`). In future phases, reuse will be triggered by:

- Planner reading an approved knowledge object while drafting a phase.
- Context Engine selecting an approved knowledge object for a prompt.
- Research Engine citing an approved knowledge object.
- A user explicitly invoking `knowledge reuse <id>`.

Reuse is a count, not a list. This phase does not record which consumer reused what. That audit trail will come later.

## Storage

Knowledge data lives in `project.db` only. The Phase 5 migration `0004_knowledge_base` is additive and idempotent:

- adds `category`, `status`, `source_type` columns to `knowledge_objects` with safe `DEFAULT` values so existing rows are not dropped;
- creates `knowledge_tags`, `knowledge_relations`, and `knowledge_references` with `UNIQUE` constraints and indexes;
- uses `ON CONFLICT DO NOTHING` for tag, relation, and reference inserts.

The global store is not touched.

## Service

The `internal/knowledge` package owns a `Repository` interface and a `Service` that enforces invariants. The service is the only entry point that mutations should use. The service:

- generates stable IDs;
- clamps `confidence` to `[0, 1]`;
- fills in `category` from the classifier when empty;
- normalizes tags, relations, and references;
- validates that relation targets exist;
- refreshes `updated_at` on every write;
- returns the updated object on reuse so callers can show the new count.

The store layer implements the `Repository` interface. A compile-time assertion (`var _ knowledge.Repository = KnowledgeRepository{}`) enforces the contract.

## CLI

`plan-ai` exposes a `knowledge` subcommand group with five actions:

- `knowledge add` — create a knowledge object. Required: `--topic`. Optional: `--category`, `--summary`, `--content`, `--confidence`, `--source`, `--status`, `--tag` (repeatable).
- `knowledge list` — list all knowledge objects, optionally filtered by `--category`.
- `knowledge show <id>` — print one knowledge object plus its tags, relations, and references.
- `knowledge search <query>` — substring search on `topic`, `summary`, and `content`. No regex, no fuzzy match, no semantic search.
- `knowledge reuse <id>` — increment the reuse count and print the new value.

`plan-ai dev seed-knowledge` seeds three sample approved objects (`PostgreSQL Multi Tenant`, `OAuth 2.0`, `Stripe Billing`) for local testing.

`plan-ai status` prints a `Knowledge:` block with `total`, per-status counts, and a `reused` sum when a project database exists.

## What This Phase Does Not Promise

- It does not auto-generate knowledge from scans, research, or code.
- It does not suggest knowledge to the user.
- It does not deduplicate semantically similar knowledge.
- It does not validate references against other stores.
- It does not export or import knowledge.
- It does not provide a UI.

All of that belongs to later phases.
