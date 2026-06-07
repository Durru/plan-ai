package store_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/Durru/plan-ai/internal/research"
	"github.com/Durru/plan-ai/internal/store"
)

func openResearchDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "research.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := store.RunProjectMigrations(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestResearchRepositoryCreateAndGetEntry(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	entry := research.ResearchEntry{
		Topic:      "Test Research",
		Summary:    "A test research entry",
		Category:   research.CategoryArchitecture,
		Confidence: 85,
	}

	if err := r.CreateEntry(entry); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	// Retrieve by ID — need to find it
	entries, err := r.ListEntries()
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	got, err := r.GetEntry(entries[0].ID)
	if err != nil {
		t.Fatalf("GetEntry: %v", err)
	}

	if got.Topic != "Test Research" {
		t.Errorf("topic = %q, want %q", got.Topic, "Test Research")
	}
	if got.Category != research.CategoryArchitecture {
		t.Errorf("category = %q, want %q", got.Category, research.CategoryArchitecture)
	}
	if got.Confidence != 85 {
		t.Errorf("confidence = %d, want 85", got.Confidence)
	}
}

func TestResearchRepositoryListEntries(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	entries := []research.ResearchEntry{
		{Topic: "Alpha", Summary: "First entry"},
		{Topic: "Beta", Summary: "Second entry"},
	}
	for _, e := range entries {
		if err := r.CreateEntry(e); err != nil {
			t.Fatalf("CreateEntry: %v", err)
		}
	}

	got, err := r.ListEntries()
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}
}

func TestResearchRepositorySearchEntries(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	entries := []research.ResearchEntry{
		{Topic: "Go Performance", Summary: "Optimizing Go programs"},
		{Topic: "Python Web", Summary: "Building web apps in Python"},
	}
	for _, e := range entries {
		if err := r.CreateEntry(e); err != nil {
			t.Fatalf("CreateEntry: %v", err)
		}
	}

	// Search by topic keyword
	got, err := r.SearchEntries("go")
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}

	// Search by summary keyword
	got, err = r.SearchEntries("Python")
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 match, got %d", len(got))
	}

	// No match
	got, err = r.SearchEntries("zzzznonexistent")
	if err != nil {
		t.Fatalf("SearchEntries: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(got))
	}
}

func TestResearchRepositoryUpdateEntryStatus(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}

	entries, _ := r.ListEntries()
	id := entries[0].ID

	if entries[0].Status != research.ResearchStatusDraft {
		t.Errorf("expected draft, got %q", entries[0].Status)
	}

	if err := r.UpdateEntryStatus(id, research.ResearchStatusApproved); err != nil {
		t.Fatalf("UpdateEntryStatus: %v", err)
	}

	got, err := r.GetEntry(id)
	if err != nil {
		t.Fatalf("GetEntry: %v", err)
	}
	if got.Status != research.ResearchStatusApproved {
		t.Errorf("status = %q, want %q", got.Status, research.ResearchStatusApproved)
	}
}

func TestResearchRepositoryDeleteEntryCascades(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "ToDelete"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	// Add sub-entities
	if err := r.CreateFinding(research.ResearchFinding{ResearchID: id, Title: "Finding1"}); err != nil {
		t.Fatalf("CreateFinding: %v", err)
	}
	if err := r.CreateSource(research.ResearchSource{ResearchID: id, Title: "Source1"}); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}
	if err := r.AddTag(id, "tag1"); err != nil {
		t.Fatalf("AddTag: %v", err)
	}

	// Verify they exist
	findings, _ := r.ListFindings(id)
	if len(findings) != 1 {
		t.Fatalf("expected 1 finding before delete")
	}

	if err := r.DeleteEntry(id); err != nil {
		t.Fatalf("DeleteEntry: %v", err)
	}

	// Entry should be gone
	_, err := r.GetEntry(id)
	if err == nil {
		t.Fatal("expected error after delete")
	}

	// Sub-entities should be gone
	findings, _ = r.ListFindings(id)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings after cascade, got %d", len(findings))
	}
}

func TestResearchRepositoryFindings(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Findings Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	f1 := research.ResearchFinding{ResearchID: id, Title: "Important", Content: "Critical", Importance: 5}
	f2 := research.ResearchFinding{ResearchID: id, Title: "Minor", Content: "Trivial", Importance: 1}

	if err := r.CreateFinding(f1); err != nil {
		t.Fatalf("CreateFinding f1: %v", err)
	}
	if err := r.CreateFinding(f2); err != nil {
		t.Fatalf("CreateFinding f2: %v", err)
	}

	findings, err := r.ListFindings(id)
	if err != nil {
		t.Fatalf("ListFindings: %v", err)
	}
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	// Should be ordered by importance DESC
	if findings[0].Title != "Important" {
		t.Errorf("first finding should be Important, got %q", findings[0].Title)
	}
}

func TestResearchRepositorySources(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Sources Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	s := research.ResearchSource{
		ResearchID: id,
		Title:      "Source Title",
		URL:        "https://example.com",
		SourceType: research.SourceTypeManual,
	}
	if err := r.CreateSource(s); err != nil {
		t.Fatalf("CreateSource: %v", err)
	}

	sources, err := r.ListSources(id)
	if err != nil {
		t.Fatalf("ListSources: %v", err)
	}
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
	if sources[0].Title != "Source Title" {
		t.Errorf("title = %q", sources[0].Title)
	}
	if sources[0].SourceType != research.SourceTypeManual {
		t.Errorf("source_type = %q", sources[0].SourceType)
	}
}

func TestResearchRepositoryConclusions(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Conclusions Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	c1 := research.ResearchConclusion{ResearchID: id, Content: "Main conclusion", Confidence: 90}
	c2 := research.ResearchConclusion{ResearchID: id, Content: "Secondary conclusion", Confidence: 50}

	if err := r.CreateConclusion(c1); err != nil {
		t.Fatalf("CreateConclusion c1: %v", err)
	}
	if err := r.CreateConclusion(c2); err != nil {
		t.Fatalf("CreateConclusion c2: %v", err)
	}

	conclusions, err := r.ListConclusions(id)
	if err != nil {
		t.Fatalf("ListConclusions: %v", err)
	}
	if len(conclusions) != 2 {
		t.Fatalf("expected 2 conclusions, got %d", len(conclusions))
	}
	// Ordered by confidence DESC
	if conclusions[0].Confidence != 90 {
		t.Errorf("first conclusion should have confidence 90, got %d", conclusions[0].Confidence)
	}
}

func TestResearchRepositoryTags(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Tags Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	if err := r.AddTag(id, "tag1"); err != nil {
		t.Fatalf("AddTag tag1: %v", err)
	}
	if err := r.AddTag(id, "tag2"); err != nil {
		t.Fatalf("AddTag tag2: %v", err)
	}

	tags, err := r.ListTags(id)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	// Duplicate should be silently ignored
	if err := r.AddTag(id, "tag1"); err != nil {
		t.Fatalf("AddTag duplicate: %v", err)
	}
	tags, err = r.ListTags(id)
	if err != nil {
		t.Fatalf("ListTags after dup: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags after duplicate, got %d", len(tags))
	}
}

func TestResearchRepositoryKnowledgeLinks(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	if err := r.CreateEntry(research.ResearchEntry{Topic: "Links Test"}); err != nil {
		t.Fatalf("CreateEntry: %v", err)
	}
	entries, _ := r.ListEntries()
	id := entries[0].ID

	if err := r.LinkKnowledge(id, "k1"); err != nil {
		t.Fatalf("LinkKnowledge k1: %v", err)
	}
	if err := r.LinkKnowledge(id, "k2"); err != nil {
		t.Fatalf("LinkKnowledge k2: %v", err)
	}

	links, err := r.ListKnowledgeLinks(id)
	if err != nil {
		t.Fatalf("ListKnowledgeLinks: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}

	// Duplicate should be silently ignored
	if err := r.LinkKnowledge(id, "k1"); err != nil {
		t.Fatalf("LinkKnowledge dup: %v", err)
	}
	links, err = r.ListKnowledgeLinks(id)
	if err != nil {
		t.Fatalf("ListKnowledgeLinks after dup: %v", err)
	}
	if len(links) != 2 {
		t.Errorf("expected 2 links after duplicate, got %d", len(links))
	}
}

func TestResearchRepositorySummary(t *testing.T) {
	db := openResearchDB(t)
	r := store.NewResearchRepository(db)

	// Initially all zeros
	s, err := r.Summary()
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if s.Total != 0 {
		t.Errorf("expected Total=0, got %d", s.Total)
	}

	// Create entries with different statuses
	create := func(topic string, status research.ResearchStatus) {
		entry := research.ResearchEntry{Topic: topic}
		if err := r.CreateEntry(entry); err != nil {
			t.Fatalf("CreateEntry(%q): %v", topic, err)
		}
		entries, _ := r.ListEntries()
		// Find the last one
		for _, e := range entries {
			if e.Topic == topic {
				r.UpdateEntryStatus(e.ID, status)
				break
			}
		}
	}

	create("Draft1", research.ResearchStatusDraft)
	create("Draft2", research.ResearchStatusDraft)
	create("Approved1", research.ResearchStatusApproved)
	create("Rejected1", research.ResearchStatusRejected)

	s, err = r.Summary()
	if err != nil {
		t.Fatalf("Summary after inserts: %v", err)
	}
	if s.Total != 4 {
		t.Errorf("Total = %d, want 4", s.Total)
	}
	if s.Draft != 2 {
		t.Errorf("Draft = %d, want 2", s.Draft)
	}
	if s.Approved != 1 {
		t.Errorf("Approved = %d, want 1", s.Approved)
	}
	if s.Rejected != 1 {
		t.Errorf("Rejected = %d, want 1", s.Rejected)
	}
}
