# Domain Relations

## Ownership Tree

Every entity is owned (directly or transitively) by a Project.
The tree below shows the parent–child relationships.

```
Project
│
├── Vision              [project_id]        — 0..N per project
│
├── Requirement         [project_id]        — 0..N per project
│
├── Constraint          [project_id]        — 0..N per project
│
├── Decision            [project_id]        — 0..N per project
│
├── MasterPlan          [project_id]        — 0..N per project
│   │
│   ├── SpecificPlan    [master_plan_id]    — 0..N per master
│   │   │
│   │   ├── Phase       [plan_id]           — 1..N per plan
│   │   │   │
│   │   │   └── Task    [phase_id, plan_id] — 0..N per phase
│   │   │
│   │   └── ImplementationDocument [specific_plan_id] — 0..1 per plan
│   │
│   └── (recursive: SpecificPlan → Plan store table)
│
├── Research            [project_id]        — 0..N per project
│   │
│   └── ResearchSource  [research_id]       — 0..N per research
│
├── KnowledgeObject     (cross-project via category/type)
│   │
│   ├── KnowledgeTag    [knowledge_id]      — 0..N per object
│   ├── KnowledgeRelation [source_id/target_id] — 0..N per object
│   └── KnowledgeReference [knowledge_id]   — 0..N per object
│
├── Snapshot            [project_id]        — 0..N per project
│
└── ChangeRequest       [project_id]        — 0..N per project
    │
    └── ImpactReport    [change_request_id] — 1 per request
```

## Cross-Cutting Relationships

### Research ↔ Knowledge

Research findings can be promoted to KnowledgeObjects. The link is recorded
in the `research_knowledge_links` table (migration 0005). This is a 0..N
relationship: one Research can link to many KnowledgeObjects, and one
KnowledgeObject can be linked from many Research entries.

### Decision ↔ Knowledge

Approved Decisions can be referenced by KnowledgeObjects via
`KnowledgeReference` with `ReferenceType = "decision"`. This is how settled
architectural choices become reusable knowledge.

### Validation → Target

A `Validation` targets exactly one entity (plan, phase, task, or decision)
via `target_type` + `target_id`. This is an N:1 relationship — one entity
can have many validations over time.

### ImplementationDocument → SpecificPlan

An `ImplementationDocument` is derived from exactly one `SpecificPlan`.
It is not a plan itself — it is a downstream deliverable guide. The
relationship is 0..1: a plan may or may not have been documented yet.

### ImpactReport → ChangeRequest

An `ImpactReport` is always derived from a single `ChangeRequest`. They
have a 1:1 relationship: each approved/requested change has exactly one
impact report, and each impact report belongs to exactly one change request.

## Relationship Table Summary

| Source | Target | Cardinality | Via |
|---|---|---|---|
| Vision | Project | N:1 | vision.project_id |
| Requirement | Project | N:1 | requirement.project_id |
| Constraint | Project | N:1 | constraint.project_id |
| Decision | Project | N:1 | decision.project_id |
| MasterPlan | Project | N:1 | master_plan.project_id |
| SpecificPlan | MasterPlan | N:1 | specific_plan.master_plan_id |
| Phase | SpecificPlan | N:1 | phase.plan_id |
| Task | Phase | N:1 | task.phase_id |
| ImplementationDocument | SpecificPlan | 0..1 | impl_doc.specific_plan_id |
| Research | Project | N:1 | research.project_id |
| ResearchSource | Research | N:1 | research_source.research_id |
| KnowledgeTag | KnowledgeObject | N:1 | knowledge_tag.knowledge_id |
| KnowledgeRelation | KnowledgeObject | N:N | relation.source_id / target_id |
| KnowledgeReference | KnowledgeObject | N:1 | knowledge_ref.knowledge_id |
| Snapshot | Project | N:1 | snapshot.project_id |
| ChangeRequest | Project | N:1 | change_request.project_id |
| ImpactReport | ChangeRequest | 1:1 | impact_report.change_request_id |
| Validation | (any entity) | N:1 | validation.target_id + target_type |
