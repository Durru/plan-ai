# Project Lifecycle Workflow

Plan-AI manages projects as a continuous planning lifecycle. The lifecycle begins with an idea and continues through implementation feedback and replanning.

## Lifecycle Overview

```text
Idea -> Context -> Questions -> Research -> Plan -> Approval -> Implementation -> Changes -> Replanning
```

Each stage produces or updates planning objects. The lifecycle is not a one-way tunnel; changes may send the project back to questions, research, planning, or approval.

## 1. Idea

An idea starts the lifecycle. It may be a product request, technical improvement, integration concept, bug-prevention initiative, or operational need.

Plan-AI captures the idea without pretending it is complete. The immediate goal is to identify the intended outcome and the uncertainty around it.

## 2. Context

Plan-AI gathers the context required to reason accurately. Context may include existing documentation, repository structure, constraints, previous decisions, relevant dependencies, target users, and known risks.

The output of this stage is a clear separation between known information and missing information.

## 3. Questions

Plan-AI asks targeted questions for missing information that materially changes the plan. It avoids asking broad questionnaires when a single decision-oriented question would unblock progress.

Questions should clarify objective, scope, constraints, priority, approval boundaries, or risk tolerance.

## 4. Research

Plan-AI researches unresolved technical or product unknowns. Research is scoped to planning needs and captured as reusable knowledge.

Research should identify sources, confidence, recommendations, common errors, and impact on decisions or tasks.

## 5. Plan

Plan-AI builds the master plan or specific plan. The plan connects objective, scope, non-scope, users, architecture, decisions, risks, dependencies, roadmap, phases, tasks, and success criteria.

The plan must be practical enough for implementation and explicit enough for review.

## 6. Approval

Plan-AI presents the plan for human approval. Approval confirms that the proposed scope, decisions, phase structure, and validation criteria are acceptable.

Approved planning state becomes the reference for implementation and future change analysis.

## 7. Implementation

Implementation happens outside Plan-AI's planning responsibility. A developer or implementation agent uses the approved plan, phases, and tasks to modify the system.

Plan-AI may serve context during implementation, but it does not become the executor.

## 8. Changes

During or after implementation, new information may appear: a library does not behave as expected, scope changes, architecture constraints emerge, validation fails, or a stakeholder changes priority.

Plan-AI classifies changes by impact and identifies affected planning objects.

## 9. Replanning

Plan-AI updates plans when project reality changes. Replanning preserves valid work, revises affected work, records decision changes, and requests renewed approval when necessary.

Replanning keeps the project honest: approved plans should reflect the work actually being prepared.
