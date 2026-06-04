package research

import "testing"

func TestResearchRegistryCreatesListsAndRegistersFindings(t *testing.T) {
	repo := NewMemoryRegistryRepository()
	registry := NewRegistry(repo)
	job, err := registry.CreateResearch(CreateResearchRequest{ProjectID: "project:test", Topic: "SQLite research", Summary: "Evaluate SQLite", Findings: []ResearchFinding{{Title: "WAL", Content: "Use WAL"}}, Recommendations: []ResearchRecommendation{{Content: "Enable busy timeout"}}, Sources: []ResearchSource{{Title: "SQLite docs", URL: "https://sqlite.org"}}, Confidence: 0.8})
	if err != nil {
		t.Fatalf("CreateResearch error: %v", err)
	}
	if job.Status != ResearchStatusDraft || len(job.Findings) != 1 || len(job.Recommendations) != 1 {
		t.Fatalf("unexpected job: %+v", job)
	}
	got, err := registry.GetResearch(job.ID)
	if err != nil {
		t.Fatalf("GetResearch error: %v", err)
	}
	if got.Topic != "SQLite research" {
		t.Fatalf("unexpected topic: %s", got.Topic)
	}
	items, err := registry.ListResearch("project:test")
	if err != nil {
		t.Fatalf("ListResearch error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 research job, got %d", len(items))
	}
}
