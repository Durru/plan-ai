# Plan-AI MVP Audit

## Scope

This audit covers the implemented Plan-AI system from Phase 0 through Phase 22, plus the Phase 23-27 functionalization block.

## Ready components

- Store Layer: SQLite global/project stores, additive migrations, repository coverage, sandbox verification.
- Ingestion/Vision/Approved Context: input capture, normalized sources, vision drafts, approved facts.
- Research/Knowledge: reusable research and knowledge objects with link compatibility.
- Planning/Workflow: master/specific/implementation planning primitives and workflow runs.
- Model Strategy/Capabilities/Orchestrator: model tier selection, capability registry, job orchestration.
- Context Engine: derived context views and chunks with L0-L4 final delivery support.
- Change/MCP/OpenCode: change impact/snapshots, MCP tool registry, read-only OpenCode detection.
- Agent/Continuous Planning: intent routing, delegated job records, plan update proposals, context levels.
- Phase 24-27 engines: vision discovery, versioned master/specific plan generation, budget-aware context delivery.

## Incomplete or intentionally minimal components

- MCP handlers are service-layer stubs for some operations until real provider/workflow wiring is added.
- Vision discovery uses deterministic scaffolding and structured outputs; LLM intelligence is still expected to provide semantic depth.
- Plan generators create structured artifacts and version metadata, but final strategic judgment remains LLM-driven.
- OpenCode integration is read-only detection and registry preparation, not full advanced integration.

## Architecture-only or documentation-heavy areas

- Advanced external provider integrations.
- Deep Engram integration.
- Advanced Skill Intelligence.
- TUI.

## Functional components

Every major package compiles and is covered by unit and/or integration sandbox checks. The sandbox exercises storage, CLI flows, MCP tools, safety markers, and Phase 23-27 validation commands.

## Non-functional components

No real external LLM calls, real OpenCode mutation, or real user config writes are implemented. This is intentional and preserves sandbox safety.

## Future dependencies

- Real model provider adapters.
- Real MCP transport integration beyond current CLI/server primitives.
- Optional OpenCode advanced integration.
- Optional Engram deep sync.

## Discardable components

- Legacy placeholder commands and compatibility views can eventually be removed after a migration/backfill policy exists.
- Internal alias MCP tool names can remain for compatibility but are not the canonical contract.
