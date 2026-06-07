package research

import (
	"fmt"
	"strings"
	"time"

	"github.com/Durru/plan-ai/internal/domain"
)

// ResearchStatus constants.
type ResearchStatus string

const (
	ResearchStatusDraft     ResearchStatus = "draft"
	ResearchStatusRunning   ResearchStatus = "running"
	ResearchStatusCompleted ResearchStatus = "completed"
	ResearchStatusInReview  ResearchStatus = "in_review"
	ResearchStatusApproved  ResearchStatus = "approved"
	ResearchStatusRejected  ResearchStatus = "rejected"
	ResearchStatusArchived  ResearchStatus = "archived"
)

// ResearchSourceType constants.
type ResearchSourceType string

const (
	SourceTypeManual        ResearchSourceType = "manual"
	SourceTypeDocumentation ResearchSourceType = "documentation"
	SourceTypeArticle       ResearchSourceType = "article"
	SourceTypeRepository    ResearchSourceType = "repository"
	SourceTypeSpecification ResearchSourceType = "specification"
	SourceTypeBenchmark     ResearchSourceType = "benchmark"
	SourceTypeInternal      ResearchSourceType = "internal"
)

// ResearchCategory re-exports domain.KnowledgeCategory for callers that
// prefer the research package as the import surface.
type ResearchCategory = domain.KnowledgeCategory

const (
	CategoryDatabase       ResearchCategory = "database"
	CategoryAuthentication ResearchCategory = "authentication"
	CategoryBilling        ResearchCategory = "billing"
	CategoryFrontend       ResearchCategory = "frontend"
	CategoryBackend        ResearchCategory = "backend"
	CategorySecurity       ResearchCategory = "security"
	CategoryDeployment     ResearchCategory = "deployment"
	CategoryArchitecture   ResearchCategory = "architecture"
	CategoryTesting        ResearchCategory = "testing"
	CategoryMCP            ResearchCategory = "mcp"
	CategoryAgents         ResearchCategory = "agents"
	CategoryAI             ResearchCategory = "ai"
	CategoryDevops         ResearchCategory = "devops"
	CategoryIntegration    ResearchCategory = "integration"
	CategoryGeneral        ResearchCategory = "general"
)

// ResearchEntry is the canonical research entity.
type ResearchEntry struct {
	ID         string           `json:"id"`
	ProjectID  string           `json:"project_id"`
	Topic      string           `json:"topic"`
	Category   ResearchCategory `json:"category"`
	Summary    string           `json:"summary"`
	Status     ResearchStatus   `json:"status"`
	Confidence int              `json:"confidence"` // 0-100
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

// ResearchJob is the Phase 12 aggregate used by the research registry.
type ResearchJob struct {
	ID              string                   `json:"id"`
	ProjectID       string                   `json:"project_id"`
	Topic           string                   `json:"topic"`
	Summary         string                   `json:"summary"`
	Findings        []ResearchFinding        `json:"findings"`
	Recommendations []ResearchRecommendation `json:"recommendations"`
	Sources         []ResearchSource         `json:"sources"`
	Confidence      float64                  `json:"confidence"`
	Status          ResearchStatus           `json:"status"`
	CreatedAt       time.Time                `json:"created_at"`
}

type ResearchRecommendation struct {
	ID         string    `json:"id"`
	ResearchID string    `json:"research_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreateResearchRequest struct {
	ProjectID       string
	Topic           string
	Summary         string
	Findings        []ResearchFinding
	Recommendations []ResearchRecommendation
	Sources         []ResearchSource
	Confidence      float64
}

// ResearchFinding is a single finding within a research entry.
type ResearchFinding struct {
	ID         string    `json:"id"`
	ResearchID string    `json:"research_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	Importance int       `json:"importance"` // 1-5
	CreatedAt  time.Time `json:"created_at"`
}

// ResearchSource is a reference/source for a research entry.
type ResearchSource struct {
	ID         string             `json:"id"`
	ResearchID string             `json:"research_id"`
	Title      string             `json:"title"`
	URL        string             `json:"url"`
	SourceType ResearchSourceType `json:"source_type"`
	CreatedAt  time.Time          `json:"created_at"`
}

// ResearchConclusion is a conclusion for a research entry.
type ResearchConclusion struct {
	ID         string    `json:"id"`
	ResearchID string    `json:"research_id"`
	Content    string    `json:"content"`
	Confidence int       `json:"confidence"` // 0-100
	CreatedAt  time.Time `json:"created_at"`
}

// ResearchTag is a tag on a research entry.
type ResearchTag struct {
	ID         string `json:"id"`
	ResearchID string `json:"research_id"`
	Tag        string `json:"tag"`
}

// ResearchKnowledgeLink links a research entry to a knowledge object.
type ResearchKnowledgeLink struct {
	ID          string    `json:"id"`
	ResearchID  string    `json:"research_id"`
	KnowledgeID string    `json:"knowledge_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// ValidationError is a single validation failure.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
