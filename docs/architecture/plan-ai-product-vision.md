# Plan-AI Product Vision

**Status:** Proposed
**Scope:** UX, storage model, installation model, agent interaction, and planning workflow

## Goal

Turn Plan-AI into a permanent planning layer that lives outside any single model session or project repo, with conversation-first usage and global persistence.

## Core Principles

### 1. Install Once

Plan-AI should be installed once on the machine.

After installation it should:
- detect OpenCode and other compatible tools
- register its MCP automatically
- configure required integrations
- create global storage
- remain available for all projects

The user should not have to manage complex config files manually.

### 2. Conversation First

Commands should exist, but they are secondary.

Primary commands:
- `install`
- `update`
- `uninstall`
- `doctor`
- `setup`

Everything else should be possible through conversation:
- analyze this project
- create a SaaS
- tell me what is next
- create the database plan
- analyze the impact of this change

### 3. External Storage

Plan-AI should not write inside the project by default.

It should store everything in its own space:
- vision
- decisions
- research
- plans
- tasks
- snapshots
- approved context

Each project should be associated with a stable identifier.
The repository should stay clean.

### 4. Permanent Layer

Models change.
Chats end.
Sessions disappear.

Plan-AI should preserve:
- knowledge
- decisions
- research
- vision
- approved context

The memory lives in Plan-AI, not in the model.

### 5. Discovery Before Planning

Plan-AI should not create plans immediately.

First it should:
- understand the idea
- discover the expected outcome
- identify features
- detect constraints
- understand what the user does not want

Only after that should it create plans.

### 6. Approved Context Is Never Re-Asked

Everything the user approves becomes persistent context.

Plan-AI should consult existing knowledge first.
It should only ask for new or ambiguous information.

### 7. Structured Research

Every important topic should be researched once.

The result should become reusable knowledge that future plans can reuse without repeating work.

### 8. Continuous Planning

Plan-AI should not generate a plan and disappear.

It should stay with the full cycle:
- idea
- vision
- research
- planning
- implementation
- changes
- replanning

It should keep the project coherent over time.

### 9. Optional Agent

Plan-AI should work without an agent.

The agent should improve the experience by:
- guiding the user
- detecting opportunities
- checking project state
- proposing next steps
- coordinating research

But the source of truth must remain Plan-AI, not the agent.

### 10. Main Mission

Plan-AI is not just a plan generator.

Its mission is to reduce the distance between the idea in the user’s head and the product that gets built.

## Current Architectural Gaps

The current codebase already has strong pieces:
- MCP integration
- OpenCode setup
- research and knowledge systems
- approved context
- continuous planning pieces

But the default persistence model still leans too much toward project-local storage.

Current issues to address:
- project data still defaults to `<project>/.plan-ai`
- project identity is path-based instead of registry-based
- the conversation gateway still feels command-first in many flows
- discovery is present, but not always the first step
- approved context exists, but should become a stronger authority boundary

## Proposed Direction

### Storage Model

Use a global Plan-AI home as the durable source of truth:

```txt
~/.plan-ai/
  global.db
  projects/
    <project-id>/
      project.db
      snapshots/
      exports/
```

### Project Identity

Each project should have a stable identifier managed by Plan-AI, not just a path-derived identity.

### Experience Model

The agent should become the conversation gateway.
Commands should remain available for power users and automation, but they should not define the primary product experience.

## First Implementation Step

The first architectural step should be moving default persistence from project-local storage to a global registry plus external project store.

Without that, the rest of the vision stays layered on top of the repo instead of around it.
