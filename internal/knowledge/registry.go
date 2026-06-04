package knowledge

import (
	"fmt"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/domain"
)

type RegistryRepository interface {
	CreateKnowledgeObject(KnowledgeObject) (KnowledgeObject, error)
	GetKnowledgeObject(string) (KnowledgeObject, error)
	ListKnowledgeObjects(projectID string) ([]KnowledgeObject, error)
	SearchKnowledgeObjects(query string) ([]KnowledgeObject, error)
}

type Registry struct {
	repo RegistryRepository
	now  func() time.Time
}

func NewRegistry(repo RegistryRepository) *Registry {
	return &Registry{repo: repo, now: time.Now().UTC}
}

func (r *Registry) CreateKnowledge(req CreateKnowledgeRequest) (KnowledgeObject, error) {
	if strings.TrimSpace(req.ProjectID) == "" {
		return KnowledgeObject{}, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(req.Title) == "" {
		return KnowledgeObject{}, fmt.Errorf("title is required")
	}
	category := req.Category
	if category == "" {
		category = domain.KnowledgeCategoryGeneral
	}
	now := r.now().Format(time.RFC3339)
	object := KnowledgeObject{ID: domain.NewID("knowledge"), ProjectID: req.ProjectID, Title: strings.TrimSpace(req.Title), Category: category, Summary: strings.TrimSpace(req.Summary), ResearchIDs: req.ResearchIDs, RelatedDecisions: req.RelatedDecisions, RelatedRequirements: req.RelatedRequirements, RelatedConstraints: req.RelatedConstraints, Confidence: clampConfidence(req.Confidence), CreatedAt: now, UpdatedAt: now}
	return r.repo.CreateKnowledgeObject(object)
}

func (r *Registry) GetKnowledge(id string) (KnowledgeObject, error) {
	if strings.TrimSpace(id) == "" {
		return KnowledgeObject{}, fmt.Errorf("id is required")
	}
	return r.repo.GetKnowledgeObject(id)
}
func (r *Registry) ListKnowledge(projectID string) ([]KnowledgeObject, error) {
	return r.repo.ListKnowledgeObjects(projectID)
}
func (r *Registry) SearchKnowledge(query string) ([]KnowledgeObject, error) {
	return r.repo.SearchKnowledgeObjects(query)
}

func clampConfidence(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

type MemoryRegistryRepository struct{ objects []KnowledgeObject }

func NewMemoryRegistryRepository() *MemoryRegistryRepository { return &MemoryRegistryRepository{} }
func (m *MemoryRegistryRepository) CreateKnowledgeObject(object KnowledgeObject) (KnowledgeObject, error) {
	m.objects = append(m.objects, object)
	return object, nil
}
func (m *MemoryRegistryRepository) GetKnowledgeObject(id string) (KnowledgeObject, error) {
	for _, object := range m.objects {
		if object.ID == id {
			return object, nil
		}
	}
	return KnowledgeObject{}, fmt.Errorf("knowledge %q not found", id)
}
func (m *MemoryRegistryRepository) ListKnowledgeObjects(projectID string) ([]KnowledgeObject, error) {
	var out []KnowledgeObject
	for _, object := range m.objects {
		if projectID == "" || object.ProjectID == projectID {
			out = append(out, object)
		}
	}
	return out, nil
}
func (m *MemoryRegistryRepository) SearchKnowledgeObjects(query string) ([]KnowledgeObject, error) {
	q := strings.ToLower(strings.TrimSpace(query))
	var out []KnowledgeObject
	for _, object := range m.objects {
		if q == "" || strings.Contains(strings.ToLower(object.Title), q) || strings.Contains(strings.ToLower(object.Summary), q) {
			out = append(out, object)
		}
	}
	return out, nil
}
