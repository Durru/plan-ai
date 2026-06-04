package modelstrategy

import "fmt"

// ──────────────────────────────────────────────
// Model compatibility — Phase 48
// ──────────────────────────────────────────────

// CompatibilityReport describes whether a model+provider combination is valid.
type CompatibilityReport struct {
	Model        string       `json:"model"`
	Provider     ProviderType `json:"provider"`
	Supported    bool         `json:"supported"`
	MaxTokens    int          `json:"max_tokens,omitempty"`
	Tier         ModelTier    `json:"tier,omitempty"`
	Alternatives []string     `json:"alternatives,omitempty"`
	Note         string       `json:"note,omitempty"`
}

// ModelEntry defines a known model with its default provider, max tokens, and tier.
type ModelEntry struct {
	Model     string
	Provider  ProviderType
	MaxTokens int
	Tier      ModelTier
}

// CompatibilityCatalog holds the known model universe and provides lookups.
type CompatibilityCatalog struct {
	entries []ModelEntry
	index   map[string]ModelEntry // key: "model|provider"
}

// NewCompatibilityCatalog creates a catalog with default known models.
func NewCompatibilityCatalog() *CompatibilityCatalog {
	c := &CompatibilityCatalog{
		index: make(map[string]ModelEntry),
	}
	for _, e := range defaultModels() {
		c.add(e)
	}
	return c
}

// NewCompatibilityCatalogWith creates a catalog from explicit entries.
func NewCompatibilityCatalogWith(entries []ModelEntry) *CompatibilityCatalog {
	c := &CompatibilityCatalog{
		entries: make([]ModelEntry, 0, len(entries)),
		index:   make(map[string]ModelEntry),
	}
	for _, e := range entries {
		c.add(e)
	}
	return c
}

func (c *CompatibilityCatalog) add(e ModelEntry) {
	c.entries = append(c.entries, e)
	c.index[e.Model+"|"+string(e.Provider)] = e
}

// Check returns a compatibility report for a model+provider pair.
func (c *CompatibilityCatalog) Check(model string, provider ProviderType) CompatibilityReport {
	key := model + "|" + string(provider)
	if e, ok := c.index[key]; ok {
		return CompatibilityReport{
			Model:     model,
			Provider:  provider,
			Supported: true,
			MaxTokens: e.MaxTokens,
			Tier:      e.Tier,
		}
	}
	// Model might be known — check if it exists under a different provider
	var alts []string
	for _, e := range c.entries {
		if e.Model == model {
			alts = append(alts, string(e.Provider))
		}
	}
	r := CompatibilityReport{
		Model:     model,
		Provider:  provider,
		Supported: false,
	}
	if len(alts) > 0 {
		r.Alternatives = alts
		r.Note = fmt.Sprintf("model %q is known but not supported with provider %q; try: %v", model, provider, alts)
	} else {
		r.Note = fmt.Sprintf("model %q is not in the compatibility catalog", model)
	}
	return r
}

// ListModels returns all known models, optionally filtered by provider.
func (c *CompatibilityCatalog) ListModels(provider ...ProviderType) []ModelEntry {
	if len(provider) == 0 {
		return c.entries
	}
	filter := make(map[ProviderType]bool, len(provider))
	for _, p := range provider {
		filter[p] = true
	}
	var out []ModelEntry
	for _, e := range c.entries {
		if filter[e.Provider] {
			out = append(out, e)
		}
	}
	return out
}

// ListProviders returns all distinct provider types in the catalog.
func (c *CompatibilityCatalog) ListProviders() []ProviderType {
	seen := make(map[ProviderType]bool)
	var out []ProviderType
	for _, e := range c.entries {
		if !seen[e.Provider] {
			seen[e.Provider] = true
			out = append(out, e.Provider)
		}
	}
	return out
}

// ──────────────────────────────────────────────
// Default model catalog
// ──────────────────────────────────────────────

func defaultModels() []ModelEntry {
	return []ModelEntry{
		// Anthropic
		{Model: "claude-sonnet-4-20250514", Provider: ProviderAnthropic, MaxTokens: 65536, Tier: TierLarge},
		{Model: "claude-sonnet-4-20250514", Provider: ProviderAnthropic, MaxTokens: 32768, Tier: TierMedium},
		{Model: "claude-sonnet-4-20250514", Provider: ProviderAnthropic, MaxTokens: 8192, Tier: TierSmall},
		{Model: "claude-haiku-3-5-20241022", Provider: ProviderAnthropic, MaxTokens: 8192, Tier: TierSmall},
		{Model: "claude-opus-4-20250514", Provider: ProviderAnthropic, MaxTokens: 65536, Tier: TierLarge},

		// OpenAI
		{Model: "gpt-4o", Provider: ProviderOpenAI, MaxTokens: 16384, Tier: TierLarge},
		{Model: "gpt-4o-mini", Provider: ProviderOpenAI, MaxTokens: 16384, Tier: TierSmall},
		{Model: "o3-mini", Provider: ProviderOpenAI, MaxTokens: 32768, Tier: TierMedium},

		// Google Gemini
		{Model: "gemini-2.0-flash", Provider: ProviderGemini, MaxTokens: 32768, Tier: TierMedium},
		{Model: "gemini-2.5-pro", Provider: ProviderGemini, MaxTokens: 65536, Tier: TierLarge},

		// DeepSeek
		{Model: "deepseek-chat", Provider: ProviderDeepSeek, MaxTokens: 32768, Tier: TierMedium},
		{Model: "deepseek-reasoner", Provider: ProviderDeepSeek, MaxTokens: 8192, Tier: TierLarge},

		// Qwen
		{Model: "qwen-max", Provider: ProviderQwen, MaxTokens: 32768, Tier: TierLarge},
		{Model: "qwen-plus", Provider: ProviderQwen, MaxTokens: 32768, Tier: TierMedium},
		{Model: "qwen-turbo", Provider: ProviderQwen, MaxTokens: 8192, Tier: TierSmall},

		// OpenAI-compatible (generic)
		{Model: "*", Provider: ProviderOpenAICompatible, MaxTokens: 16384, Tier: TierMedium},

		// OpenRouter
		{Model: "*", Provider: ProviderOpenRouter, MaxTokens: 65536, Tier: TierMedium},

		// Local
		{Model: "*", Provider: ProviderLocal, MaxTokens: 4096, Tier: TierSmall},
	}
}

// ──────────────────────────────────────────────
// Compatibility helper methods
// ──────────────────────────────────────────────

// IsModelKnown returns true if the model name appears explicitly in the catalog.
// Wildcard entries ("*") do NOT count — this checks for specific model registrations.
func (c *CompatibilityCatalog) IsModelKnown(model string) bool {
	for _, e := range c.entries {
		if e.Model == model {
			return true
		}
	}
	return false
}

// SupportsProvider returns true if the provider has any entries.
func (c *CompatibilityCatalog) SupportsProvider(pt ProviderType) bool {
	for _, e := range c.entries {
		if e.Provider == pt {
			return true
		}
	}
	return false
}
