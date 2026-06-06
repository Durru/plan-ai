// Package research implements the Research Engine — topic classification,
// entry CRUD, findings/sources/conclusions, reuse (Phase 7), and
// orchestration. It enforces v4 Principle 2: "Nada investigado se
// vuelve a investigar."
//
// Main types: ResearchEntry, ResearchJob, ReuseService, Orchestrator.
// Main entry: Service.CreateResearch for validated research creation.
package research
