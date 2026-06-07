package memory

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/Durru/plan-ai/internal/domain"
)

// ──────────────────────────────────────────────
// Entry types
// ──────────────────────────────────────────────

// EntryType categorizes a memory entry.
type EntryType string

const (
	TypeDecision       EntryType = "decision"
	TypeApproval       EntryType = "approval"
	TypeQuestionAnswer EntryType = "question_answer"
	TypeReference      EntryType = "reference"
	TypeResearch       EntryType = "research"
	TypePlan           EntryType = "plan"
	TypeChange         EntryType = "change"
)

// Valid returns true if the entry type is known.
func (et EntryType) Valid() bool {
	switch et {
	case TypeDecision, TypeApproval, TypeQuestionAnswer,
		TypeReference, TypeResearch, TypePlan, TypeChange:
		return true
	default:
		return false
	}
}

// ──────────────────────────────────────────────
// Entry
// ──────────────────────────────────────────────

// Entry is a single durable project memory record.
type Entry struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	EntryType EntryType `json:"entry_type"`
	Title     string    `json:"title"`
	Question  string    `json:"question,omitempty"`
	Answer    string    `json:"answer,omitempty"`
	Content   string    `json:"content,omitempty"`
	Citation  string    `json:"citation,omitempty"`
	Source    string    `json:"source,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewEntry creates a new memory entry with generated ID.
func NewEntry(projectID string, entryType EntryType, title, content, citation, source string) Entry {
	return NewEntryFull(projectID, entryType, title, "", "", content, citation, source)
}

// NewEntryFull creates a memory entry with all fields including question/answer.
func NewEntryFull(projectID string, entryType EntryType, title, question, answer, content, citation, source string) Entry {
	now := time.Now().UTC()
	return Entry{
		ID:        domain.NewID("mem"),
		ProjectID: projectID,
		EntryType: entryType,
		Title:     title,
		Question:  question,
		Answer:    answer,
		Content:   content,
		Citation:  citation,
		Source:    source,
		Status:    "active",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ──────────────────────────────────────────────
// Sentinal errors
// ──────────────────────────────────────────────

var ErrNotFound = errors.New("memory entry not found")

// ──────────────────────────────────────────────
// Store interface
// ──────────────────────────────────────────────

// Store defines persistence operations for memory entries.
type Store interface {
	Add(e Entry) (Entry, error)
	List(projectID string) ([]Entry, error)
	Search(projectID, query string) ([]Entry, error)
	Get(id string) (Entry, error)
	Update(e Entry) (Entry, error)
}

// ──────────────────────────────────────────────
// AddInput
// ──────────────────────────────────────────────

// AddInput specifies the fields for creating a new memory entry.
type AddInput struct {
	ProjectID string    `json:"project_id"`
	EntryType EntryType `json:"entry_type"`
	Title     string    `json:"title"`
	Question  string    `json:"question,omitempty"`
	Answer    string    `json:"answer,omitempty"`
	Content   string    `json:"content,omitempty"`
	Citation  string    `json:"citation,omitempty"`
	Source    string    `json:"source,omitempty"`
}

// ──────────────────────────────────────────────
// Service
// ──────────────────────────────────────────────

// Service provides memory CRUD and question reuse.
type Service struct {
	store Store
}

// NewService creates a memory service backed by the given store.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Add creates a new memory entry.
func (s *Service) Add(input AddInput) (Entry, error) {
	if !input.EntryType.Valid() {
		return Entry{}, fmt.Errorf("invalid entry type: %q", input.EntryType)
	}
	e := NewEntryFull(input.ProjectID, input.EntryType, input.Title, input.Question, input.Answer, input.Content, input.Citation, input.Source)
	return s.store.Add(e)
}

// Ask searches for a matching question and reuses the answer if found.
// Returns the entry, reused=true if a match was found, or an empty entry with reused=false.
func (s *Service) Ask(projectID, question string) (Entry, bool, error) {
	entries, err := s.store.List(projectID)
	if err != nil {
		return Entry{}, false, err
	}
	for _, e := range entries {
		if e.EntryType == TypeQuestionAnswer && e.Question != "" {
			if MatchQuestion(e.Question, question) {
				return e, true, nil
			}
		}
	}
	return Entry{}, false, nil
}

// List returns all memory entries for a project.
func (s *Service) List(projectID string) ([]Entry, error) {
	return s.store.List(projectID)
}

// Search searches memory entries by keyword.
func (s *Service) Search(projectID, query string) ([]Entry, error) {
	return s.store.Search(projectID, query)
}

// Get returns a single memory entry by ID.
func (s *Service) Get(id string) (Entry, error) {
	return s.store.Get(id)
}

// ──────────────────────────────────────────────
// Question normalization and matching
// ──────────────────────────────────────────────

// NormalizeQuestion normalizes a question for comparison:
// lowercase, trim, remove punctuation, normalize whitespace.
// Hyphens, slashes, and similar separators are treated as whitespace.
func NormalizeQuestion(q string) string {
	q = strings.ToLower(strings.TrimSpace(q))
	// Treat separators as spaces
	var b strings.Builder
	for _, r := range q {
		if r == '-' || r == '/' || r == '_' || r == '\\' || r == '|' {
			b.WriteRune(' ')
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			b.WriteRune(r)
		}
		// else drop the character (punctuation removal)
	}
	// Normalize whitespace
	fields := strings.Fields(b.String())
	return strings.Join(fields, " ")
}

// MatchQuestion returns true if two questions match after normalization.
func MatchQuestion(a, b string) bool {
	return NormalizeQuestion(a) == NormalizeQuestion(b)
}
