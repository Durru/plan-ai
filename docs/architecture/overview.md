# Plan-AI Architecture Overview

Plan-AI is a modular Go application for continuous implementation planning. It is independent first and integrable later.

Plan-AI does not implement. Plan-AI prepares implementation.

## Active baseline

The active repository contains the old Phase 0–6 foundation:

- Cobra CLI and configuration helpers.
- SQLite global/project persistence.
- Deterministic scanner.
- Initial domain repositories.
- Knowledge Base.
- Research Engine.

This baseline is kept so the project compiles and can be validated while the new roadmap begins.

## New architecture sequence

The next phase is not Skill Intelligence. The next phase is **Phase 7 — Definitive Domain Model**.

After that, Plan-AI should evolve through store, ingestion, vision, approved context, research/knowledge, planning, workflow, model strategy, capability registry, context, change, MCP, OpenCode integration, optional agent, and continuous planning layers.

## Boundaries

- No OpenCode dependency in core.
- No MCP dependency in domain/store.
- No real user config mutation during tests.
- No planning by improvisation; planning state must be stored and reviewable.
