package reference

import (
	"strings"
	"time"
)

type Status string
type Category string
type SourceType string

const (
	StatusNeedsReview Status = "needs_review"
	StatusApproved    Status = "approved"
	StatusRejected    Status = "rejected"

	CategoryVisual     Category = "visual"
	CategoryUX         Category = "ux"
	CategoryFunctional Category = "functional"
	CategoryTechnical  Category = "technical"
	CategoryBusiness   Category = "business"

	SourceURL        SourceType = "url"
	SourceImage      SourceType = "image"
	SourceDocument   SourceType = "document"
	SourceRepository SourceType = "repository"
	SourceScreenshot SourceType = "screenshot"
	SourceExample    SourceType = "example"
)

type Reference struct {
	ID        string
	ProjectID string
	Source    SourceType
	URI       string
	Title     string
	Category  Category
	Notes     string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Repository interface {
	SaveReference(Reference) (Reference, error)
	ListReferences(projectID string) ([]Reference, error)
	ApproveReference(id string) (Reference, error)
	RejectReference(id string) (Reference, error)
}

type Service struct{ repo Repository }

func NewService(repo Repository) Service { return Service{repo: repo} }

func (s Service) Add(projectID string, source SourceType, uri, title string, category Category) (Reference, error) {
	return s.repo.SaveReference(Build(projectID, source, uri, title, category))
}

func (s Service) List(projectID string) ([]Reference, error) { return s.repo.ListReferences(projectID) }

func (s Service) Approve(id string) (Reference, error) { return s.repo.ApproveReference(id) }

func (s Service) Reject(id string) (Reference, error) { return s.repo.RejectReference(id) }

func Build(projectID string, source SourceType, uri, title string, category Category) Reference {
	if source == "" {
		source = SourceURL
	}
	if category == "" {
		category = classify(uri + " " + title)
	}
	if strings.TrimSpace(title) == "" {
		title = uri
	}
	return Reference{ProjectID: projectID, Source: source, URI: uri, Title: title, Category: category, Notes: "Reference requires approval before planning use.", Status: StatusNeedsReview}
}

func classify(value string) Category {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(lower, "stripe"):
		return CategoryUX
	case strings.Contains(lower, "github"):
		return CategoryTechnical
	default:
		return CategoryFunctional
	}
}
