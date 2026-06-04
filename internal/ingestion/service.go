package ingestion

import "fmt"

type Service struct{ repo Repository }

func NewService(repo Repository) Service { return Service{repo: repo} }

func (s Service) Ingest(req InputRequest) (RawInput, IngestedSource, error) {
	if req.ProjectID == "" {
		return RawInput{}, IngestedSource{}, fmt.Errorf("project id is required")
	}
	if req.SourceType == "" {
		req.SourceType = SourceTypePrompt
	}
	normalized := Normalize(req.Content)
	raw, err := s.repo.CreateRawInput(RawInput{ProjectID: req.ProjectID, SourceType: req.SourceType, RawContent: req.Content, Metadata: req.Metadata})
	if err != nil {
		return RawInput{}, IngestedSource{}, err
	}
	metadata := map[string]string{"blocks": fmt.Sprint(len(normalized.Blocks)), "list_items": fmt.Sprint(len(normalized.ListItems))}
	for k, v := range req.Metadata {
		metadata[k] = v
	}
	source, err := s.repo.CreateIngestedSource(IngestedSource{ProjectID: req.ProjectID, RawInputID: raw.ID, SourceType: req.SourceType, NormalizedContent: normalized.Content, Classification: Classify(normalized.Content), Metadata: metadata})
	if err != nil {
		return RawInput{}, IngestedSource{}, err
	}
	return raw, source, nil
}
