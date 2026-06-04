# ADR 0004: Deterministic Project Scanner

## Status

Accepted

## Context

Plan-AI needs to understand a project's local repository before later phases can produce useful plans, research prompts, knowledge records, or context slices. That understanding must be available offline, must be testable, and must not depend on AI behavior.

Phase 4 therefore needs a local scanner that records the repository's stack evidence in `project.db`.

## Decision

Implement a deterministic, rule-based scanner before any AI-assisted analysis.

The scanner detects Git, languages, package managers, frameworks, dependencies, important files, and source files using local file names, extensions, manifests, lockfiles, and simple parsers. It stores scan summaries and child records in project-local SQLite tables.

It does not read or store full file contents. It does not execute Planner, Research Engine, Skill Intelligence, MCP, OpenCode, Engram, agents, or AI workflows.

## Rationale

A deterministic scanner first is the correct substrate because it is:

- reproducible: the same repository state produces the same scan facts;
- testable: rules can be validated with fixtures and table-driven tests;
- cheap: no LLM calls, network access, or heavyweight external tooling;
- offline-first: it works in local repositories without cloud services;
- safe: it indexes metadata only and avoids storing full source contents.

Persisting results in `project.db` makes the scan available to future phases without repeatedly walking the repository.

## Consequences

Future phases can layer richer behavior on top of this stable baseline:

- Planner can adapt plans to detected stacks.
- Research can focus on unknowns instead of rediscovering obvious dependencies.
- Knowledge can reuse scan facts as durable project context.
- Context Engine can use file counts and kinds to estimate context budgets.

The tradeoff is that deterministic rules will miss some frameworks or dependencies that require deeper semantic analysis. That is acceptable for Phase 4 because correctness and reproducibility matter more than exhaustive detection.

## Alternatives Considered

### Pure AI scanner

Rejected. It would be expensive, non-deterministic, harder to test, and premature before Plan-AI has a stable local project model.

### Hybrid scanner with AI fallback

Rejected for this phase. Hybrid behavior can be added later once deterministic storage, fingerprints, and scan history are reliable.

### Third-party stack scanners

Deferred. External scanners may add coverage later, but Phase 4 should stay dependency-light and under Plan-AI's control.