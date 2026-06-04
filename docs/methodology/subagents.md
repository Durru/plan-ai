# Temporary Subagent Architecture

Plan-AI may use temporary specialist subagents to improve planning quality. Subagents are not the product. Plan-AI is the product. Subagents are disposable reasoning workers that support Plan-AI's planning process.

## Core Rule

A subagent exists only while it is useful for a planning question. It is born for a bounded investigation, produces a structured planning deliverable, and disappears after its output is absorbed into Plan-AI's planning state.

Subagents do not own project truth, approval, or implementation. Plan-AI remains responsible for synthesizing their findings into plans, decisions, tasks, and knowledge objects.

## When Subagents Are Born

Plan-AI may create a subagent when:

- a planning question requires specialist knowledge;
- parallel investigation would reduce uncertainty;
- a decision needs adversarial review;
- a plan needs domain-specific validation;
- the main planning thread would become overloaded by detailed research.

Subagents should not be created for routine summarization, simple clarification, or work that Plan-AI can handle directly.

## What Subagents Deliver

Every subagent should deliver a bounded report containing:

- question investigated;
- assumptions;
- findings;
- recommendation;
- risks;
- confidence level;
- planning impact;
- sources or evidence;
- follow-up questions if needed.

Plan-AI then decides what becomes reusable knowledge, what becomes a decision candidate, and what changes the plan.

## When Subagents Disappear

A subagent disappears after delivering its report and after Plan-AI has absorbed the useful output. The subagent should not persist as an independent authority or runtime component.

If the same specialty is needed again later, Plan-AI may create a new temporary subagent with fresh scope and current context.

## Research Specialist

The Research Specialist investigates external or project-local unknowns that affect planning. It compares sources, identifies confidence, and converts findings into planning recommendations.

Use when a plan depends on facts that are not already captured in project knowledge.

## Database Specialist

The Database Specialist evaluates data modeling, persistence boundaries, migration risk, query patterns, storage constraints, and data ownership from a planning perspective.

Use when decisions involve schema design, persistence technology, migration sequencing, or data integrity.

## Security Specialist

The Security Specialist evaluates threat models, permission boundaries, secret handling, trust assumptions, authentication, authorization, data exposure, and secure validation needs.

Use when plans affect sensitive data, execution permissions, external integrations, user identity, or operational trust.

## Frontend Specialist

The Frontend Specialist evaluates user flows, interface boundaries, accessibility implications, state management, integration with APIs, and frontend validation strategy.

Use when planning work affects visible user experience or frontend architecture.

## Backend Specialist

The Backend Specialist evaluates service boundaries, APIs, domain logic placement, concurrency, reliability, operational behavior, and integration contracts.

Use when planning work affects backend modules, APIs, background work, or server-side architecture.

## Testing Specialist

The Testing Specialist evaluates validation strategy, test boundaries, regression risks, fixtures, acceptance criteria, and verification commands.

Use when a plan needs confidence that implementation can be verified objectively.

## MCP Specialist

The MCP Specialist evaluates Model Context Protocol planning concerns, tool boundaries, resources, prompts, permissions, context serving, and integration contracts.

Use when future Plan-AI capabilities involve MCP behavior or agent-facing context delivery.

## Synthesis Responsibility

Subagent output is raw specialist input. Plan-AI must synthesize it into the official planning model. A subagent recommendation becomes authoritative only when Plan-AI records it in an approved plan, decision, or knowledge object.
