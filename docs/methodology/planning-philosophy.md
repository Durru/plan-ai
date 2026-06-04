# Planning Philosophy

Plan-AI treats planning as a structured reasoning process. Its purpose is to transform ambiguity into approved, executable preparation without crossing into implementation.

## How Plan-AI Thinks

Plan-AI thinks by decomposing project intent into explicit planning objects. It avoids treating a conversation as the source of truth. Instead, it extracts durable objects that can be reviewed, revised, and reused.

The system asks:

1. What outcome is desired?
2. What is already known?
3. What remains uncertain?
4. Which decisions constrain the solution?
5. Which plan best satisfies the objective under those constraints?
6. Which phases and tasks prepare implementation safely?
7. What validations prove readiness?

## Idea

An idea is an initial possibility. It may be incomplete, contradictory, or unvalidated. Ideas are allowed to be rough because their role is to start exploration.

Plan-AI should not convert an idea directly into tasks. It first clarifies the objective, constraints, risks, and missing information.

## Objective

An objective is the desired outcome stated in operational terms. It defines what success means before implementation begins.

An objective should be specific enough to guide scope and validation. If the objective is vague, Plan-AI asks clarifying questions before producing an authoritative plan.

## Decision

A decision is a committed choice that affects the plan. It may concern architecture, libraries, scope, workflow, validation strategy, integration boundaries, or product behavior.

Decisions differ from preferences because they carry consequences. If reversing a choice would change tasks, dependencies, risks, or documentation, it is a decision.

## Plan

A plan is the structured path from objective to implementation readiness. It explains scope, non-scope, assumptions, decisions, phases, risks, dependencies, validations, and success criteria.

A plan is not a list of tasks alone. Without rationale and boundaries, tasks become isolated instructions that are easy to execute incorrectly.

## Phase

A phase is a coherent segment of a plan. It groups related work by dependency, risk, reviewability, or integration value.

A phase has inputs, outputs, validations, and dependencies. It should produce a state that can be inspected before moving to the next phase.

## Task

A task is the smallest planning unit intended for execution by a developer or implementation agent. It includes objective, context, files or areas, restrictions, steps, validations, and expected result.

Tasks are implementation-ready, but they are not implementation. They describe what should be done and how to verify it.

## Implementation

Implementation is the act of changing code, configuration, infrastructure, data, or runtime behavior. It is outside Plan-AI's primary responsibility.

Plan-AI prepares implementation by making implementation safer, clearer, and easier to review. It does not replace the implementation phase.

## Boundaries Between Objects

The distinction matters because each object answers a different question:

| Object | Question It Answers |
|---|---|
| Idea | What might we build? |
| Objective | What outcome are we trying to achieve? |
| Decision | Which direction have we chosen? |
| Plan | How will we prepare the work? |
| Phase | In what coherent order should work happen? |
| Task | What should be executed next? |
| Implementation | What changed in the system? |

Plan-AI is effective when these boundaries remain explicit.
