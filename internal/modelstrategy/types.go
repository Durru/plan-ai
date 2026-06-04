package modelstrategy

import "time"

// ──────────────────────────────────────────────
// Provider types
// ──────────────────────────────────────────────

type ProviderType string

const (
	ProviderOpenAI           ProviderType = "openai"
	ProviderAnthropic        ProviderType = "anthropic"
	ProviderGemini           ProviderType = "gemini"
	ProviderOpenRouter       ProviderType = "openrouter"
	ProviderLocal            ProviderType = "local"
	ProviderDeepSeek         ProviderType = "deepseek"
	ProviderQwen             ProviderType = "qwen"
	ProviderOpenAICompatible ProviderType = "openai_compatible"
)

// ──────────────────────────────────────────────
// Task classification
// ──────────────────────────────────────────────

type TaskClassification string

const (
	TaskExtraction     TaskClassification = "extraction"
	TaskClassify       TaskClassification = "classification"
	TaskResearch       TaskClassification = "research"
	TaskPlanning       TaskClassification = "planning"
	TaskValidation     TaskClassification = "validation"
	TaskImpactAnalysis TaskClassification = "impact_analysis"
	TaskSummarization  TaskClassification = "summarization"
)

// ──────────────────────────────────────────────
// Model tiers / strategy
// ──────────────────────────────────────────────

type ModelTier string

const (
	TierSmall  ModelTier = "small"
	TierMedium ModelTier = "medium"
	TierLarge  ModelTier = "large"
)

type ModelStrategy struct {
	Tier      ModelTier
	Provider  ProviderType
	ModelName string
	MaxTokens int
	CostRank  int // 1=cheapest, 3=most expensive
}

// DefaultModelStrategies returns the default small/medium/large strategies.
func DefaultModelStrategies() []ModelStrategy {
	return []ModelStrategy{
		{Tier: TierSmall, Provider: ProviderAnthropic, ModelName: "claude-sonnet-4-20250514", MaxTokens: 8192, CostRank: 1},
		{Tier: TierMedium, Provider: ProviderAnthropic, ModelName: "claude-sonnet-4-20250514", MaxTokens: 32768, CostRank: 2},
		{Tier: TierLarge, Provider: ProviderAnthropic, ModelName: "claude-sonnet-4-20250514", MaxTokens: 65536, CostRank: 3},
	}
}

// ──────────────────────────────────────────────
// Model profile (persisted)
// ──────────────────────────────────────────────

type ModelProfile struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	Config    string    `json:"config"` // JSON string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ──────────────────────────────────────────────
// Task-to-tier mapping rules
// ──────────────────────────────────────────────

// ClassificationTier maps each task classification to the recommended tier.
func ClassificationTier(tc TaskClassification) ModelTier {
	switch tc {
	case TaskExtraction, TaskClassify, TaskSummarization:
		return TierSmall
	case TaskResearch, TaskPlanning, TaskValidation:
		return TierMedium
	case TaskImpactAnalysis:
		return TierLarge
	default:
		return TierMedium
	}
}
