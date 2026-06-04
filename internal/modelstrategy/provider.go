package modelstrategy

// ModelProvider is the interface that each provider must implement.
type ModelProvider interface {
	// Generate sends a prompt to the model and returns the response content.
	Generate(request GenerateRequest) (GenerateResponse, error)
	// ProviderType returns which provider this instance represents.
	ProviderType() ProviderType
}

type GenerateRequest struct {
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	MaxTokens int    `json:"max_tokens"`
	System    string `json:"system,omitempty"`
	Schema    string `json:"schema,omitempty"` // JSON schema for structured output
}

type GenerateResponse struct {
	Content      string `json:"content"`
	Model        string `json:"model"`
	FinishReason string `json:"finish_reason"`
	Usage        Usage  `json:"usage,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
