package knowledge

import "testing"

func TestKnowledgeRegistryCreatesRelatesAndSearchesKnowledge(t *testing.T) {
	repo := NewMemoryRegistryRepository()
	registry := NewRegistry(repo)
	object, err := registry.CreateKnowledge(CreateKnowledgeRequest{ProjectID: "project:test", Title: "SQLite WAL", Category: CategoryDatabase, Summary: "Use WAL for concurrent local writes", ResearchIDs: []string{"research-1"}, RelatedDecisions: []string{"decision-1"}, Confidence: 0.9})
	if err != nil {
		t.Fatalf("CreateKnowledge error: %v", err)
	}
	if len(object.ResearchIDs) != 1 || len(object.RelatedDecisions) != 1 {
		t.Fatalf("unexpected object: %+v", object)
	}
	got, err := registry.GetKnowledge(object.ID)
	if err != nil {
		t.Fatalf("GetKnowledge error: %v", err)
	}
	if got.Title != object.Title {
		t.Fatalf("unexpected object: %+v", got)
	}
	matches, err := registry.SearchKnowledge("wal")
	if err != nil {
		t.Fatalf("SearchKnowledge error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}
