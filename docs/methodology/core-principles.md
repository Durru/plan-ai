# Core Principles

Plan-AI exists to prepare implementation before code is written. These principles define how it behaves across projects, workflows, prompts, and future integrations.

## Independent First

Plan-AI must be useful before it is connected to implementation agents, editors, task runners, MCP servers, or external memory systems. Its core value is disciplined planning, not integration complexity.

Operationally, this means every plan, decision, phase, and task should be understandable as a document. Integrations may consume those documents later, but the planning system must not depend on an integration to be meaningful.

## Integrable Later

Plan-AI outputs should be structured enough for future tools to consume. A master plan, specific plan, decision record, phase, task, or knowledge object should have stable fields and predictable meaning.

Integrability does not mean premature automation. The product should first define accurate planning objects, then expose them to automation when their semantics are stable.

## Planning Before Implementation

Plan-AI separates preparation from execution. Before implementation begins, the project should have a clear objective, boundaries, decisions, risks, dependencies, and validation path.

This principle prevents agents or developers from coding their way into unresolved product and architecture questions. Implementation should execute a plan; it should not discover the entire plan accidentally.

## Reusable Knowledge

Research, decisions, best practices, and recurring errors are project assets. Plan-AI should capture them in reusable form so future planning does not restart from zero.

Reusable knowledge must include enough context to be applied safely: what was learned, where it came from, when it applies, how confident the system is, and which plans or decisions it affects.

## Continuous Planning

A plan is not a static artifact. It remains valid only while its assumptions, decisions, and constraints remain valid. Plan-AI must detect meaningful changes and update the affected planning objects.

Continuous planning does not mean constant churn. It means the planning state remains synchronized with the project reality.

## Human Approval

Plan-AI proposes; humans approve. The system may draft decisions and plans, but project-defining choices should become authoritative only after explicit human acceptance.

Approval creates a boundary. Before approval, a plan is a candidate. After approval, it becomes the reference point for implementation and future change analysis.

## Context On Demand

Plan-AI should provide the smallest useful context for the current consumer. A developer implementing one task usually needs the task, relevant phase, constraints, and affected decisions, not every document in the project.

When more context is needed, the system expands progressively from summary to full document.

## Non-Invasive Planning

Plan-AI prepares implementation without mutating implementation code. It may inspect project structure, documentation, and configuration to understand context, but it should not change application behavior as part of planning.

This keeps planning safe, reviewable, and reversible. Code changes belong to a separate implementation phase.
