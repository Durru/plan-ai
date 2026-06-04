# Context System

Plan-AI serves context in levels so humans and agents receive enough information to act without carrying unnecessary project history.

## Context Level 1: Quick Summary

Level 1 is a short orientation. It answers what the project or task is about, what the current objective is, and what the next action likely is.

Use Level 1 when:

- starting a conversation;
- giving a status snapshot;
- routing a request;
- deciding whether deeper context is needed.

## Context Level 2: Operational Summary

Level 2 provides the working context needed to act on a bounded request. It includes objective, relevant decisions, scope boundaries, current phase, known risks, and immediate validation expectations.

Use Level 2 when:

- assigning a task;
- reviewing a small plan change;
- answering a focused project question;
- preparing an implementation agent for a bounded action.

## Context Level 3: Complete Phase

Level 3 provides the full context for one phase, including inputs, outputs, dependencies, validations, tasks, and related decisions.

Use Level 3 when:

- implementing or reviewing an entire phase;
- validating phase readiness;
- analyzing phase-level change impact;
- handing off work between agents or developers.

## Context Level 4: Complete Plan

Level 4 provides the full Master Plan or Specific Plan. It includes objective, scope, non-scope, research, architecture, decisions, risks, dependencies, roadmap, phases, tasks, and success criteria.

Use Level 4 when:

- approving a plan;
- revising scope or architecture;
- onboarding a new lead contributor;
- performing plan-level impact analysis.

## Context Level 5: Complete Document

Level 5 provides the complete source document or documentation set. It is the highest-detail context level and should be used deliberately.

Use Level 5 when:

- auditing documentation;
- resolving contradictions;
- migrating or archiving planning state;
- performing full project review;
- investigating a complex change that crosses several planning objects.

## Selection Rule

Plan-AI should start with the lowest context level that can satisfy the task. If the consumer lacks enough information, the system expands one level at a time.

This approach protects attention, reduces token load for agents, and keeps human review focused.
