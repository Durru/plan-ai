# Architecture

## Overview

Plan-AI is a **local-first, continuous implementation planning engine** for AI-assisted software projects. It prepares plans, tracks approved context, manages research and knowledge, and detects project changes — all in local SQLite stores.

### Core principle

**Plan-AI prepares plans. It does not implement code.**

This separation ensures clean ownership: Plan-AI owns what should be built and why. Implementation tools own how to build it.

## Layer architecture

```
┌─────────────────────────────────────────────────────┐
│                     CLI Layer                        │
│      cmd/plan-ai/ (Cobra, 20+ commands)             │
├─────────────────────────────────────────────────────┤
│                   MCP Server Layer                   │
│      cmd/plan-ai mcp serve (stdio, 30 tools)        │
├─────────────────────────────────────────────────────┤
│              Integration & Agent Layer               │
│  ┌─────────┐ ┌──────────┐ ┌───────────────────┐    │
│  │ OpenCode│ │ Workflows│ │ Agent System       │    │
│  └─────────┘ └──────────┘ └───────────────────┘    │
├─────────────────────────────────────────────────────┤
│                 Domain Engine Layer                  │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌───────────┐  │
│  │Vision  │ │Context │ │Research  │ │Knowledge  │  │
│  └────────┘ └────────┘ └──────────┘ └───────────┘  │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌───────────┐  │
│  │Planning│ │Change  │ │Continuous│ │Ingestion  │  │
│  └────────┘ └────────┘ └──────────┘ └───────────┘  │
├─────────────────────────────────────────────────────┤
│                    Scanner Layer                     │
│       internal/scanner/ (deterministic scan)        │
├─────────────────────────────────────────────────────┤
│                     Store Layer                      │
│   internal/store/ (SQLite, migrations, repos)       │
├─────────────────────────────────────────────────────┤
│                     Domain Model                     │
│          internal/domain/ (canonical entities)       │
└─────────────────────────────────────────────────────┘
```

## Key flows

### Initial setup

```
install → init → scan → ingest → vision → approved → plan
```

### Continuous planning loop

```
detect change → analyze impact → propose plan update → approve
```

### Agent processing

```
receive intent → classify → route → process → complete
```

## Storage architecture

Two SQLite databases:

**Global store** (`~/.plan-ai/global.db`):
- `known_projects` — registered project metadata

**Project store** (`<project>/.plan-ai/project.db`):
- `vision_drafts` — vision artifacts
- `approved_context` — approved context items
- `research_entries` — research records
- `research_findings` — findings
- `research_sources` — source URLs
- `research_conclusions` — conclusions
- `knowledge_entries` — knowledge base
- `master_plans` — master planning artifacts
- `specific_plans` — specific planning artifacts
- `implementation_docs` — implementation documents
- `scan_results` — scan outputs
- `ingested_inputs` — ingested input records
- `change_events` — detected changes
- `snapshots` — project snapshots
- `agent_tasks` — agent task records
- `continuous_events` — continuous planning events
- `continuous_proposals` — plan update proposals
- `workflow_executions` — workflow execution records
- `capabilities` — registered capabilities

## Domain isolation

Each domain package is structurally similar:
- Thin data types in `internal/domain/`
- Repository interfaces in `internal/store/`
- Business logic in domain engine packages
- Commands and flags in `internal/*/cli/` (where applicable)

## Integration surface

Plan-AI exposes two integration surfaces:

1. **CLI** (20+ commands) — for shell scripting, CI/CD, and manual use
2. **MCP server** (30 tools) — for AI tool discovery and automated workflows

The OpenCode integration is optional and generates static artifacts only.

## Capability registry

The capability registry (`internal/capabilities/`) allows Plan-AI to advertise which features are available. This supports:
- Agent routing decisions
- Tool discovery in MCP
- Feature detection in continuous planning
