package knowledge

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// Repository is the storage surface required by Service. It is satisfied
// by *store.KnowledgeRepository, which is why the interface lives here
// in the knowledge package: the knowledge package owns its storage
// contract, and the store package implements it.
type Repository interface {
	Create(object domain.KnowledgeObject) error
	Update(object domain.KnowledgeObject) error
	GetByID(id string) (domain.KnowledgeObject, error)
	GetByTopic(topic string) (domain.KnowledgeObject, error)
	List() ([]domain.KnowledgeObject, error)
	ListByCategory(category domain.KnowledgeCategory) ([]domain.KnowledgeObject, error)
	Search(query string) ([]domain.KnowledgeObject, error)
	IncrementReuseCount(id string) (domain.KnowledgeObject, error)
	AddTag(knowledgeID, tag string) error
	ListTags(knowledgeID string) ([]Tag, error)
	AddRelation(sourceID, targetID string, relationType domain.KnowledgeRelationType) error
	ListRelations(knowledgeID string) ([]Relation, error)
	AddReference(knowledgeID string, referenceType domain.KnowledgeReferenceType, referenceID string) error
	ListReferences(knowledgeID string) ([]Reference, error)
	Summary() (Summary, error)
}

// Service is the deterministic, AI-free entry point used by the CLI
// and future Planner/Research integration.
type Service struct {
	repo Repository
	now  func() time.Time
}

// NewService wires a knowledge service around a Repository. The clock
// is injectable for tests; production code uses time.Now.UTC.
func NewService(repo Repository) *Service {
	return &Service{repo: repo, now: func() time.Time { return time.Now().UTC() }}
}

// CreateKnowledge validates and persists a KnowledgeObject plus its
// tags. It never investigates, never calls out to the network, never
// runs an LLM, and never blocks the caller.
func (s *Service) CreateKnowledge(input CreateInput) (domain.KnowledgeObject, error) {
	topic := strings.TrimSpace(input.Topic)
	if topic == "" {
		return domain.KnowledgeObject{}, fmt.Errorf("topic is required")
	}

	category := NormalizeCategory(string(input.Category), topic)
	status, err := NormalizeStatus(string(input.Status))
	if err != nil {
		return domain.KnowledgeObject{}, err
	}
	sourceType, err := NormalizeSourceType(string(input.SourceType))
	if err != nil {
		return domain.KnowledgeObject{}, err
	}
	confidence := input.Confidence
	if confidence <= 0 {
		confidence = 0.5
	}
	if confidence > 1 {
		confidence = 1
	}

	now := s.now()
	object := domain.KnowledgeObject{
		ID:         strings.TrimSpace(input.ID),
		Topic:      topic,
		Category:   category,
		Summary:    strings.TrimSpace(input.Summary),
		Content:    input.Content,
		Confidence: confidence,
		SourceType: sourceType,
		ReuseCount: 0,
		Status:     status,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if object.ID == "" {
		object.ID = domain.NewID("knowledge")
	}

	if err := s.repo.Create(object); err != nil {
		return object, err
	}

	seen := make(map[string]struct{}, len(input.Tags))
	for _, raw := range input.Tags {
		tag, err := NormalizeTag(raw)
		if err != nil {
			return object, err
		}
		if _, dup := seen[tag]; dup {
			continue
		}
		seen[tag] = struct{}{}
		if err := s.repo.AddTag(object.ID, tag); err != nil {
			return object, err
		}
	}

	return object, nil
}

// GetKnowledge returns a single KnowledgeObject by ID.
func (s *Service) GetKnowledge(id string) (domain.KnowledgeObject, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.KnowledgeObject{}, fmt.Errorf("id is required")
	}
	return s.repo.GetByID(id)
}

// ListKnowledge returns every KnowledgeObject ordered by creation time.
func (s *Service) ListKnowledge() ([]domain.KnowledgeObject, error) {
	return s.repo.List()
}

// ListByCategory returns KnowledgeObjects filtered by category.
func (s *Service) ListByCategory(category domain.KnowledgeCategory) ([]domain.KnowledgeObject, error) {
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}
	return s.repo.ListByCategory(category)
}

// UpdateKnowledge updates the mutable fields of a KnowledgeObject. The
// ID and creation timestamp are preserved.
func (s *Service) UpdateKnowledge(id string, input UpdateInput) (domain.KnowledgeObject, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.KnowledgeObject{}, fmt.Errorf("id is required")
	}
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return existing, err
	}

	topic := strings.TrimSpace(input.Topic)
	if topic != "" {
		existing.Topic = topic
	}
	if cat := NormalizeCategory(string(input.Category), existing.Topic); cat != "" {
		existing.Category = cat
	}
	if input.Summary != "" {
		existing.Summary = strings.TrimSpace(input.Summary)
	}
	if input.Content != "" {
		existing.Content = input.Content
	}
	if input.Confidence > 0 {
		if input.Confidence > 1 {
			input.Confidence = 1
		}
		existing.Confidence = input.Confidence
	}
	if input.SourceType != "" {
		normalized, err := NormalizeSourceType(string(input.SourceType))
		if err != nil {
			return existing, err
		}
		existing.SourceType = normalized
	}
	if input.Status != "" {
		normalized, err := NormalizeStatus(string(input.Status))
		if err != nil {
			return existing, err
		}
		existing.Status = normalized
	}
	existing.UpdatedAt = s.now()

	if err := s.repo.Update(existing); err != nil {
		return existing, err
	}
	return existing, nil
}

// ReuseKnowledge increments reuse_count and refreshes updated_at. It
// is the cheapest way to record that a knowledge object was referenced
// by another artifact.
func (s *Service) ReuseKnowledge(id string) (domain.KnowledgeObject, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.KnowledgeObject{}, fmt.Errorf("id is required")
	}
	return s.repo.IncrementReuseCount(id)
}

// LinkKnowledge creates a directed relation between two existing
// KnowledgeObjects. Source and target IDs must reference existing
// records; relation type must be one of the official values.
func (s *Service) LinkKnowledge(sourceID, targetID string, relationType domain.KnowledgeRelationType) error {
	sourceID = strings.TrimSpace(sourceID)
	targetID = strings.TrimSpace(targetID)
	if sourceID == "" || targetID == "" {
		return fmt.Errorf("source and target ids are required")
	}
	if sourceID == targetID {
		return fmt.Errorf("source and target must differ")
	}
	normalized, err := NormalizeRelationType(string(relationType))
	if err != nil {
		return err
	}
	if _, err := s.repo.GetByID(sourceID); err != nil {
		return fmt.Errorf("source knowledge %q not found: %w", sourceID, err)
	}
	if _, err := s.repo.GetByID(targetID); err != nil {
		return fmt.Errorf("target knowledge %q not found: %w", targetID, err)
	}
	return s.repo.AddRelation(sourceID, targetID, normalized)
}

// AttachReference links a KnowledgeObject to another artifact
// (plan, decision, research, technology) inside the same project.
func (s *Service) AttachReference(knowledgeID string, referenceType domain.KnowledgeReferenceType, referenceID string) error {
	knowledgeID = strings.TrimSpace(knowledgeID)
	referenceID = strings.TrimSpace(referenceID)
	if knowledgeID == "" || referenceID == "" {
		return fmt.Errorf("knowledge id and reference id are required")
	}
	normalized, err := NormalizeReferenceType(string(referenceType))
	if err != nil {
		return err
	}
	if _, err := s.repo.GetByID(knowledgeID); err != nil {
		return fmt.Errorf("knowledge %q not found: %w", knowledgeID, err)
	}
	return s.repo.AddReference(knowledgeID, normalized, referenceID)
}

// AddTag adds a tag to a KnowledgeObject. Tag is normalized before
// persistence. Duplicates are silently ignored.
func (s *Service) AddTag(knowledgeID, raw string) error {
	knowledgeID = strings.TrimSpace(knowledgeID)
	if knowledgeID == "" {
		return fmt.Errorf("knowledge id is required")
	}
	tag, err := NormalizeTag(raw)
	if err != nil {
		return err
	}
	if _, err := s.repo.GetByID(knowledgeID); err != nil {
		return fmt.Errorf("knowledge %q not found: %w", knowledgeID, err)
	}
	return s.repo.AddTag(knowledgeID, tag)
}

// SearchKnowledge runs a deterministic LIKE search across topic,
// summary, and content. Empty queries list every knowledge object.
func (s *Service) SearchKnowledge(query string) ([]domain.KnowledgeObject, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return s.repo.List()
	}
	return s.repo.Search(query)
}

// Describe returns the KnowledgeObject plus its tags, relations, and
// references in a single call. It is the canonical "show" payload.
func (s *Service) Describe(id string) (domain.KnowledgeObject, []Tag, []Relation, []Reference, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return domain.KnowledgeObject{}, nil, nil, nil, fmt.Errorf("id is required")
	}
	object, err := s.repo.GetByID(id)
	if err != nil {
		return object, nil, nil, nil, err
	}
	tags, err := s.repo.ListTags(id)
	if err != nil {
		return object, nil, nil, nil, err
	}
	relations, err := s.repo.ListRelations(id)
	if err != nil {
		return object, tags, nil, nil, err
	}
	references, err := s.repo.ListReferences(id)
	if err != nil {
		return object, tags, relations, nil, err
	}
	return object, tags, relations, references, nil
}

// GetSummary returns aggregate counts used by the status command.
func (s *Service) GetSummary() (Summary, error) {
	return s.repo.Summary()
}
