# Research Engine

Phase 12 stores reusable project research so the same topic is not investigated repeatedly.

- Package: `internal/research`
- Store tables: `research_jobs`, `research_entries`, `research_sources`, `research_findings`, `research_recommendations`
- Statuses: `draft`, `running`, `completed`, `archived` plus legacy review statuses kept for compatibility
- Registry methods: `CreateResearch`, `GetResearch`, `ListResearch`

The engine is deterministic storage and registry logic only. It does not crawl the web or call an LLM.
