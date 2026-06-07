package research

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

type RegistryRepository interface {
	CreateResearchJob(ResearchJob) (ResearchJob, error)
	GetResearchJob(string) (ResearchJob, error)
	ListResearchJobs(projectID string) ([]ResearchJob, error)
}

type Registry struct {
	repo RegistryRepository
	now  func() time.Time
}

func NewRegistry(repo RegistryRepository) *Registry {
	return &Registry{repo: repo, now: time.Now().UTC}
}

func (r *Registry) CreateResearch(req CreateResearchRequest) (ResearchJob, error) {
	if strings.TrimSpace(req.ProjectID) == "" {
		return ResearchJob{}, fmt.Errorf("project id is required")
	}
	if strings.TrimSpace(req.Topic) == "" {
		return ResearchJob{}, fmt.Errorf("topic is required")
	}
	now := r.now()
	job := ResearchJob{ID: domain.NewID("research"), ProjectID: req.ProjectID, Topic: strings.TrimSpace(req.Topic), Summary: strings.TrimSpace(req.Summary), Confidence: clampConfidence(req.Confidence), Status: ResearchStatusDraft, CreatedAt: now}
	for _, f := range req.Findings {
		if f.ID == "" {
			f.ID = domain.NewID("finding")
		}
		f.ResearchID = job.ID
		if f.CreatedAt.IsZero() {
			f.CreatedAt = now
		}
		job.Findings = append(job.Findings, f)
	}
	for _, rec := range req.Recommendations {
		if rec.ID == "" {
			rec.ID = domain.NewID("recommendation")
		}
		rec.ResearchID = job.ID
		if rec.CreatedAt.IsZero() {
			rec.CreatedAt = now
		}
		job.Recommendations = append(job.Recommendations, rec)
	}
	for _, src := range req.Sources {
		if src.ID == "" {
			src.ID = domain.NewID("source")
		}
		src.ResearchID = job.ID
		if src.CreatedAt.IsZero() {
			src.CreatedAt = now
		}
		job.Sources = append(job.Sources, src)
	}
	return r.repo.CreateResearchJob(job)
}

func (r *Registry) GetResearch(id string) (ResearchJob, error) {
	if strings.TrimSpace(id) == "" {
		return ResearchJob{}, fmt.Errorf("id is required")
	}
	return r.repo.GetResearchJob(id)
}
func (r *Registry) ListResearch(projectID string) ([]ResearchJob, error) {
	return r.repo.ListResearchJobs(projectID)
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

type MemoryRegistryRepository struct{ jobs []ResearchJob }

func NewMemoryRegistryRepository() *MemoryRegistryRepository { return &MemoryRegistryRepository{} }
func (m *MemoryRegistryRepository) CreateResearchJob(job ResearchJob) (ResearchJob, error) {
	m.jobs = append(m.jobs, job)
	return job, nil
}
func (m *MemoryRegistryRepository) GetResearchJob(id string) (ResearchJob, error) {
	for _, job := range m.jobs {
		if job.ID == id {
			return job, nil
		}
	}
	return ResearchJob{}, fmt.Errorf("research %q not found", id)
}
func (m *MemoryRegistryRepository) ListResearchJobs(projectID string) ([]ResearchJob, error) {
	var out []ResearchJob
	for _, job := range m.jobs {
		if projectID == "" || job.ProjectID == projectID {
			out = append(out, job)
		}
	}
	return out, nil
}
