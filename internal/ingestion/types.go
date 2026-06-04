package ingestion

import "time"

type SourceType string

const (
	SourceTypeMessage             SourceType = "message"
	SourceTypePrompt              SourceType = "prompt"
	SourceTypeMarkdown            SourceType = "markdown"
	SourceTypeDocument            SourceType = "document"
	SourceTypeRepositoryReference SourceType = "repository_reference"
	SourceTypeWebsiteReference    SourceType = "website_reference"
	SourceTypeImageReference      SourceType = "image_reference"
)

type Classification string

const (
	ClassificationVision      Classification = "vision"
	ClassificationRequirement Classification = "requirement"
	ClassificationConstraint  Classification = "constraint"
	ClassificationPreference  Classification = "preference"
	ClassificationDecision    Classification = "decision"
	ClassificationReference   Classification = "reference"
	ClassificationUnknown     Classification = "unknown"
)

type RawInput struct {
	ID         string
	ProjectID  string
	SourceType SourceType
	RawContent string
	Metadata   map[string]string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type IngestedSource struct {
	ID                string
	ProjectID         string
	RawInputID        string
	SourceType        SourceType
	NormalizedContent string
	Classification    Classification
	Metadata          map[string]string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type InputRequest struct {
	ProjectID  string
	SourceType SourceType
	Content    string
	Metadata   map[string]string
}

type NormalizedInput struct {
	Content   string
	Blocks    []string
	ListItems []string
}

type Repository interface {
	CreateRawInput(RawInput) (RawInput, error)
	CreateIngestedSource(IngestedSource) (IngestedSource, error)
	GetRawInput(id string) (RawInput, error)
	GetIngestedSource(id string) (IngestedSource, error)
	ListIngestedSources(projectID string) ([]IngestedSource, error)
}
