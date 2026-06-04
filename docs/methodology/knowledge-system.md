# Knowledge System

Plan-AI uses reusable knowledge to prevent repeated investigation and to keep planning decisions consistent across phases.

## Knowledge Object

A Knowledge Object is a durable record of information that may affect future planning. It captures a finding, recommendation, best practice, common error, or source-backed fact in a structured form.

A Knowledge Object should include:

- title;
- summary;
- category;
- source or origin;
- confidence level;
- applicability conditions;
- planning impact;
- related decisions, plans, phases, or tasks;
- date or event of capture when available.

## Storing Research

Research should be stored when it resolves an uncertainty, supports a decision, invalidates an assumption, or identifies a risk. Research that does not affect planning should not be promoted into durable knowledge.

Stored research should state what was investigated, what was learned, which sources were used, and how the finding changes or confirms the plan.

## Best Practices

Best practices are guidance patterns that are likely to apply across multiple plans. Plan-AI should store them only when they are relevant to the project context.

A best practice should include where it applies and where it does not. Generic advice without applicability rules is not useful planning knowledge.

## Common Errors

Common errors are recurring mistakes that future planning or implementation should avoid. They may come from prior implementation feedback, research, dependency behavior, architecture constraints, or validation failures.

Common error records should include the symptom, cause, prevention strategy, and affected planning areas.

## Recommendations

Recommendations are suggested directions derived from research or analysis. A recommendation is not the same as an approved decision.

Recommendations should identify confidence, tradeoffs, and the decision they may inform. Once approved, the chosen direction should be recorded as a decision.

## Sources

Plan-AI should prefer project-local documents, official documentation, primary sources, accepted decisions, and verified implementation feedback. Informal sources may be useful for exploration but should carry lower confidence unless corroborated.

Source notes should be specific enough for a future reader to understand why the knowledge was trusted.

## Confidence

Confidence expresses how strongly Plan-AI should rely on a Knowledge Object.

| Confidence | Meaning |
|---|---|
| High | Supported by project evidence, primary sources, or direct validation |
| Medium | Supported by reliable but indirect evidence |
| Low | Plausible but requires confirmation before driving major decisions |

Low-confidence knowledge may guide questions or research, but it should not silently drive approved plans.

## Reuse

Before researching a familiar topic, Plan-AI should check existing Knowledge Objects. If existing knowledge is sufficient and current, it should be reused and cited in the plan.

Reuse should not become blind copying. Plan-AI must confirm that the knowledge applies to the current objective, constraints, stack, and decision context.

## Avoiding Reinvestigation

Plan-AI avoids reinvestigation by linking knowledge to decisions and plans. When a future task raises the same question, the system should first inspect prior knowledge, then research only the delta: what changed, what is uncertain, or what no longer applies.

This keeps planning fast without sacrificing correctness.
