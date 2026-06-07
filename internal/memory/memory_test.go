package memory_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/memory"
)

func TestNewEntry(t *testing.T) {
	e := memory.NewEntry("proj-1", memory.TypeDecision, "Use SQLite", "Decided to use SQLite for persistence", "", "")
	if e.ProjectID != "proj-1" {
		t.Errorf("ProjectID = %q, want proj-1", e.ProjectID)
	}
	if e.EntryType != memory.TypeDecision {
		t.Errorf("EntryType = %q, want %q", e.EntryType, memory.TypeDecision)
	}
	if e.Title != "Use SQLite" {
		t.Errorf("Title = %q", e.Title)
	}
	if e.Content != "Decided to use SQLite for persistence" {
		t.Errorf("Content = %q", e.Content)
	}
	if e.ID == "" {
		t.Error("ID is empty")
	}
	if e.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
}

func TestNewEntryQuestionAnswer(t *testing.T) {
	e := memory.NewEntryFull("proj-1", memory.TypeQuestionAnswer,
		"How do we handle auth?",
		"How do we handle auth?",
		"Use OAuth 2.0 with PKCE",
		"",
		"Security team recommendation",
		"https://docs.example.com/auth")
	if e.Question != "How do we handle auth?" {
		t.Errorf("Question = %q", e.Question)
	}
	if e.Answer != "Use OAuth 2.0 with PKCE" {
		t.Errorf("Answer = %q", e.Answer)
	}
	if e.Citation != "Security team recommendation" {
		t.Errorf("Citation = %q", e.Citation)
	}
	if e.Source != "https://docs.example.com/auth" {
		t.Errorf("Source = %q", e.Source)
	}
}

func TestNormalizeQuestion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"How do we handle auth?", "how do we handle auth"},
		{"  What is the plan?  ", "what is the plan"},
		{"How-do-we-deploy?", "how do we deploy"},
		{"What's the API?", "whats the api"},
		{"UPPERCASE QUESTION", "uppercase question"},
		{"multiple   spaces   here", "multiple spaces here"},
	}
	for _, tt := range tests {
		got := memory.NormalizeQuestion(tt.input)
		if got != tt.want {
			t.Errorf("NormalizeQuestion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestMatchQuestion(t *testing.T) {
	// Exact match
	if !memory.MatchQuestion("How do we handle auth?", "How do we handle auth?") {
		t.Error("expected exact match")
	}
	// Normalized match
	if !memory.MatchQuestion("How do we handle auth?", "How do we handle auth") {
		t.Error("expected normalized match")
	}
	// Different questions
	if memory.MatchQuestion("How do we handle auth?", "What database?") {
		t.Error("expected no match for different questions")
	}
	// Case insensitive
	if !memory.MatchQuestion("How do we handle auth?", "HOW DO WE HANDLE AUTH?") {
		t.Error("expected case insensitive match")
	}
}

func TestMemoryServiceAskReuse(t *testing.T) {
	store := newMemStore()
	svc := memory.NewService(store)

	// Add a memory entry with question/answer
	_, err := svc.Add(memory.AddInput{
		ProjectID: "proj-1",
		EntryType: memory.TypeQuestionAnswer,
		Title:     "Auth approach",
		Question:  "How do we handle auth?",
		Answer:    "Use OAuth 2.0 with PKCE",
		Citation:  "Security team",
		Source:    "https://docs.example.com/auth",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Ask the same question — should reuse
	result, reused, err := svc.Ask("proj-1", "How do we handle auth?")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if !reused {
		t.Error("expected reused=true for same question")
	}
	if result.Answer != "Use OAuth 2.0 with PKCE" {
		t.Errorf("Answer = %q, want %q", result.Answer, "Use OAuth 2.0 with PKCE")
	}
	if result.Citation != "Security team" {
		t.Errorf("Citation = %q", result.Citation)
	}
}

func TestMemoryServiceAskNoReuse(t *testing.T) {
	store := newMemStore()
	svc := memory.NewService(store)

	// Add one entry
	_, err := svc.Add(memory.AddInput{
		ProjectID: "proj-1",
		EntryType: memory.TypeQuestionAnswer,
		Title:     "Auth approach",
		Question:  "How do we handle auth?",
		Answer:    "Use OAuth 2.0 with PKCE",
	})
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Ask a different question — should NOT reuse
	_, reused, err := svc.Ask("proj-1", "What database should we use?")
	if err != nil {
		t.Fatalf("Ask: %v", err)
	}
	if reused {
		t.Error("expected reused=false for different question")
	}
}

func TestMemoryServiceList(t *testing.T) {
	store := newMemStore()
	svc := memory.NewService(store)

	_, _ = svc.Add(memory.AddInput{ProjectID: "proj-1", EntryType: memory.TypeDecision, Title: "Decision 1", Content: "Content 1"})
	_, _ = svc.Add(memory.AddInput{ProjectID: "proj-1", EntryType: memory.TypeQuestionAnswer, Title: "QA 1", Question: "Q?", Answer: "A!"})
	_, _ = svc.Add(memory.AddInput{ProjectID: "proj-2", EntryType: memory.TypeResearch, Title: "Research 1", Content: "Findings"})

	entries, err := svc.List("proj-1")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(entries))
	}
}

func TestMemoryEntryTypeValid(t *testing.T) {
	valid := []memory.EntryType{
		memory.TypeDecision,
		memory.TypeApproval,
		memory.TypeQuestionAnswer,
		memory.TypeReference,
		memory.TypeResearch,
		memory.TypePlan,
		memory.TypeChange,
	}
	for _, et := range valid {
		if !et.Valid() {
			t.Errorf("%q should be valid", et)
		}
	}
	if memory.EntryType("").Valid() {
		t.Error("empty type should be invalid")
	}
	if memory.EntryType("invalid").Valid() {
		t.Error("invalid type should be invalid")
	}
}

// newMemStore returns an in-memory store for testing.
func newMemStore() memory.Store {
	return &memStore{entries: make([]memory.Entry, 0)}
}

type memStore struct {
	entries []memory.Entry
}

func (s *memStore) Add(e memory.Entry) (memory.Entry, error) {
	s.entries = append(s.entries, e)
	return e, nil
}

func (s *memStore) List(projectID string) ([]memory.Entry, error) {
	var result []memory.Entry
	for _, e := range s.entries {
		if e.ProjectID == projectID {
			result = append(result, e)
		}
	}
	return result, nil
}

func (s *memStore) Search(projectID, query string) ([]memory.Entry, error) {
	var result []memory.Entry
	q := strings.ToLower(query)
	for _, e := range s.entries {
		if e.ProjectID == projectID {
			if strings.Contains(strings.ToLower(e.Title), q) ||
				strings.Contains(strings.ToLower(e.Content), q) ||
				strings.Contains(strings.ToLower(e.Question), q) {
				result = append(result, e)
			}
		}
	}
	return result, nil
}

func (s *memStore) Get(id string) (memory.Entry, error) {
	for _, e := range s.entries {
		if e.ID == id {
			return e, nil
		}
	}
	return memory.Entry{}, memory.ErrNotFound
}

func (s *memStore) Update(e memory.Entry) (memory.Entry, error) {
	for i, existing := range s.entries {
		if existing.ID == e.ID {
			s.entries[i] = e
			return e, nil
		}
	}
	return memory.Entry{}, memory.ErrNotFound
}

var _ time.Time // silence unused import
