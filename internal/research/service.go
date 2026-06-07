package research

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// Repository is the storage surface required by Service. It is satisfied
// by *store.ResearchRepository, which is why the interface lives here
// in the research package: the research package owns its storage contract,
// and the store package implements it.
type Repository interface {
	// ResearchEntry CRUD
	CreateEntry(entry ResearchEntry) error
	GetEntry(id string) (ResearchEntry, error)
	ListEntries() ([]ResearchEntry, error)
	SearchEntries(query string) ([]ResearchEntry, error)
	UpdateEntryStatus(id string, status ResearchStatus) error
	DeleteEntry(id string) error

	// Findings
	CreateFinding(finding ResearchFinding) error
	ListFindings(researchID string) ([]ResearchFinding, error)

	// Sources
	CreateSource(source ResearchSource) error
	ListSources(researchID string) ([]ResearchSource, error)

	// Conclusions
	CreateConclusion(conclusion ResearchConclusion) error
	ListConclusions(researchID string) ([]ResearchConclusion, error)

	// Tags
	AddTag(researchID, tag string) error
	ListTags(researchID string) ([]ResearchTag, error)

	// Knowledge Links
	LinkKnowledge(researchID, knowledgeID string) error
	ListKnowledgeLinks(researchID string) ([]ResearchKnowledgeLink, error)

	// Reuse
	IncrementReuseCount(id string) error
	EnsureFTS() error

	// Knowledge promotion
	PromoteToKnowledge(researchID string) (knowledgeID string, err error)

	// Summary
	Summary() (ResearchSummary, error)
}

// ResearchSummary is a flat projection used by status / dashboards.
type ResearchSummary struct {
	Total       int
	Draft       int
	InReview    int
	Approved    int
	Rejected    int
	Archived    int
	Findings    int
	Sources     int
	Conclusions int
}

// Service is the deterministic, AI-free entry point used by the CLI
// and future Planner integration.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService wires a research service around a Repository. The clock
// is injectable for tests; production code uses time.Now.UTC.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: time.Now().UTC}
}

// CreateOption is a functional option for CreateResearch.
type CreateOption func(*createConfig)

type createConfig struct {
	category   ResearchCategory
	summary    string
	confidence int
	tags       []string
}

// WithCategory sets a category override for the research entry.
func WithCategory(c ResearchCategory) CreateOption {
	return func(cfg *createConfig) {
		cfg.category = c
	}
}

// WithSummary sets the summary for the research entry.
func WithSummary(s string) CreateOption {
	return func(cfg *createConfig) {
		cfg.summary = s
	}
}

// WithConfidence sets the confidence level (0-100) for the research entry.
func WithConfidence(c int) CreateOption {
	return func(cfg *createConfig) {
		cfg.confidence = c
	}
}

// WithTags adds tags to the research entry.
func WithTags(tags ...string) CreateOption {
	return func(cfg *createConfig) {
		cfg.tags = tags
	}
}

// CreateResearch validates and persists a ResearchEntry plus its tags.
func (s *Service) CreateResearch(topic string, opts ...CreateOption) (ResearchEntry, error) {
	topic = strings.TrimSpace(topic)
	if topic == "" {
		return ResearchEntry{}, fmt.Errorf("topic is required")
	}

	cfg := &createConfig{
		confidence: 50,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	category := cfg.category
	if category == "" {
		category = Classify(topic)
	}
	confidence := cfg.confidence
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}

	now := s.now()
	entry := ResearchEntry{
		ID:         domain.NewID("research"),
		Topic:      topic,
		Category:   category,
		Summary:    strings.TrimSpace(cfg.summary),
		Status:     ResearchStatusDraft,
		Confidence: confidence,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.repo.CreateEntry(entry); err != nil {
		return entry, err
	}

	seen := make(map[string]struct{}, len(cfg.tags))
	for _, raw := range cfg.tags {
		tag := normalizeTag(raw)
		if tag == "" {
			continue
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		if err := s.repo.AddTag(entry.ID, tag); err != nil {
			return entry, err
		}
	}

	return entry, nil
}

// AddFinding adds a finding to a research entry.
func (s *Service) AddFinding(researchID, title, content string, importance int) (ResearchFinding, error) {
	researchID = strings.TrimSpace(researchID)
	if researchID == "" {
		return ResearchFinding{}, fmt.Errorf("research id is required")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ResearchFinding{}, fmt.Errorf("title is required")
	}
	if importance < 1 {
		importance = 1
	}
	if importance > 5 {
		importance = 5
	}

	if _, err := s.repo.GetEntry(researchID); err != nil {
		return ResearchFinding{}, fmt.Errorf("research entry %q not found: %w", researchID, err)
	}

	finding := ResearchFinding{
		ID:         domain.NewID("finding"),
		ResearchID: researchID,
		Title:      title,
		Content:    content,
		Importance: importance,
		CreatedAt:  s.now(),
	}
	if err := s.repo.CreateFinding(finding); err != nil {
		return ResearchFinding{}, err
	}
	return finding, nil
}

// AddSource adds a source to a research entry.
func (s *Service) AddSource(researchID, title, url string, sourceType ResearchSourceType) (ResearchSource, error) {
	researchID = strings.TrimSpace(researchID)
	if researchID == "" {
		return ResearchSource{}, fmt.Errorf("research id is required")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return ResearchSource{}, fmt.Errorf("title is required")
	}
	if sourceType == "" {
		sourceType = SourceTypeManual
	}

	if _, err := s.repo.GetEntry(researchID); err != nil {
		return ResearchSource{}, fmt.Errorf("research entry %q not found: %w", researchID, err)
	}

	source := ResearchSource{
		ID:         domain.NewID("source"),
		ResearchID: researchID,
		Title:      title,
		URL:        url,
		SourceType: sourceType,
		CreatedAt:  s.now(),
	}
	if err := s.repo.CreateSource(source); err != nil {
		return ResearchSource{}, err
	}
	return source, nil
}

// AddConclusion adds a conclusion to a research entry.
func (s *Service) AddConclusion(researchID, content string, confidence int) (ResearchConclusion, error) {
	researchID = strings.TrimSpace(researchID)
	if researchID == "" {
		return ResearchConclusion{}, fmt.Errorf("research id is required")
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return ResearchConclusion{}, fmt.Errorf("content is required")
	}
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}

	if _, err := s.repo.GetEntry(researchID); err != nil {
		return ResearchConclusion{}, fmt.Errorf("research entry %q not found: %w", researchID, err)
	}

	conclusion := ResearchConclusion{
		ID:         domain.NewID("conclusion"),
		ResearchID: researchID,
		Content:    content,
		Confidence: confidence,
		CreatedAt:  s.now(),
	}
	if err := s.repo.CreateConclusion(conclusion); err != nil {
		return ResearchConclusion{}, err
	}
	return conclusion, nil
}

// ApproveResearch transitions a research entry to approved after validating
// that it has at least one finding, one source, and one conclusion.
func (s *Service) ApproveResearch(id string) (ResearchEntry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ResearchEntry{}, fmt.Errorf("id is required")
	}

	if err := NewApprovalChecker(s.repo).CanApprove(id); err != nil {
		return ResearchEntry{}, err
	}

	if err := s.repo.UpdateEntryStatus(id, ResearchStatusApproved); err != nil {
		return ResearchEntry{}, err
	}
	return s.repo.GetEntry(id)
}

// RejectResearch transitions a research entry to rejected.
func (s *Service) RejectResearch(id string) (ResearchEntry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ResearchEntry{}, fmt.Errorf("id is required")
	}
	if _, err := s.repo.GetEntry(id); err != nil {
		return ResearchEntry{}, fmt.Errorf("research entry %q not found: %w", id, err)
	}
	if err := s.repo.UpdateEntryStatus(id, ResearchStatusRejected); err != nil {
		return ResearchEntry{}, err
	}
	return s.repo.GetEntry(id)
}

// ArchiveResearch transitions a research entry to archived.
func (s *Service) ArchiveResearch(id string) (ResearchEntry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ResearchEntry{}, fmt.Errorf("id is required")
	}
	if _, err := s.repo.GetEntry(id); err != nil {
		return ResearchEntry{}, fmt.Errorf("research entry %q not found: %w", id, err)
	}
	if err := s.repo.UpdateEntryStatus(id, ResearchStatusArchived); err != nil {
		return ResearchEntry{}, err
	}
	return s.repo.GetEntry(id)
}

// GetResearch returns a single ResearchEntry by ID.
func (s *Service) GetResearch(id string) (ResearchEntry, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ResearchEntry{}, fmt.Errorf("id is required")
	}
	return s.repo.GetEntry(id)
}

// ListResearch returns every ResearchEntry ordered by creation time.
func (s *Service) ListResearch() ([]ResearchEntry, error) {
	return s.repo.ListEntries()
}

// SearchResearch runs a deterministic LIKE search across topic and summary.
func (s *Service) SearchResearch(query string) ([]ResearchEntry, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.repo.ListEntries()
	}
	return s.repo.SearchEntries(query)
}

// LinkToKnowledge links a research entry to a knowledge object.
func (s *Service) LinkToKnowledge(researchID, knowledgeID string) error {
	researchID = strings.TrimSpace(researchID)
	knowledgeID = strings.TrimSpace(knowledgeID)
	if researchID == "" || knowledgeID == "" {
		return fmt.Errorf("research id and knowledge id are required")
	}
	if _, err := s.repo.GetEntry(researchID); err != nil {
		return fmt.Errorf("research entry %q not found: %w", researchID, err)
	}
	return s.repo.LinkKnowledge(researchID, knowledgeID)
}

// GetSummary returns aggregate counts used by the status command.
func (s *Service) GetSummary() (ResearchSummary, error) {
	return s.repo.Summary()
}

// Describe returns the ResearchEntry plus its findings, sources, conclusions,
// tags, and knowledge links in a single call.
func (s *Service) Describe(id string) (ResearchEntry, []ResearchFinding, []ResearchSource, []ResearchConclusion, []ResearchTag, []ResearchKnowledgeLink, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ResearchEntry{}, nil, nil, nil, nil, nil, fmt.Errorf("id is required")
	}
	entry, err := s.repo.GetEntry(id)
	if err != nil {
		return entry, nil, nil, nil, nil, nil, err
	}
	findings, err := s.repo.ListFindings(id)
	if err != nil {
		return entry, nil, nil, nil, nil, nil, err
	}
	sources, err := s.repo.ListSources(id)
	if err != nil {
		return entry, findings, nil, nil, nil, nil, err
	}
	conclusions, err := s.repo.ListConclusions(id)
	if err != nil {
		return entry, findings, sources, nil, nil, nil, err
	}
	tags, err := s.repo.ListTags(id)
	if err != nil {
		return entry, findings, sources, conclusions, nil, nil, err
	}
	links, err := s.repo.ListKnowledgeLinks(id)
	if err != nil {
		return entry, findings, sources, conclusions, tags, nil, err
	}
	return entry, findings, sources, conclusions, tags, links, nil
}

// normalizeTag normalizes a tag string.
func normalizeTag(raw string) string {
	tag := strings.TrimSpace(raw)
	tag = strings.ToLower(tag)
	return tag
}
