package modelstrategy_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Durru/plan-ai/internal/modelstrategy"
)

func TestClassificationTier(t *testing.T) {
	tests := []struct {
		class modelstrategy.TaskClassification
		want  modelstrategy.ModelTier
	}{
		{modelstrategy.TaskExtraction, modelstrategy.TierSmall},
		{modelstrategy.TaskClassify, modelstrategy.TierSmall},
		{modelstrategy.TaskSummarization, modelstrategy.TierSmall},
		{modelstrategy.TaskResearch, modelstrategy.TierMedium},
		{modelstrategy.TaskPlanning, modelstrategy.TierMedium},
		{modelstrategy.TaskValidation, modelstrategy.TierMedium},
		{modelstrategy.TaskImpactAnalysis, modelstrategy.TierLarge},
		{"unknown", modelstrategy.TierMedium},
	}
	for _, tt := range tests {
		got := modelstrategy.ClassificationTier(tt.class)
		if got != tt.want {
			t.Errorf("ClassificationTier(%q) = %q, want %q", tt.class, got, tt.want)
		}
	}
}

func TestDefaultModelStrategies(t *testing.T) {
	strats := modelstrategy.DefaultModelStrategies()
	if len(strats) != 3 {
		t.Fatalf("got %d strategies, want 3", len(strats))
	}
	tiers := map[modelstrategy.ModelTier]bool{}
	for _, s := range strats {
		tiers[s.Tier] = true
		if s.MaxTokens <= 0 {
			t.Errorf("strategy %q has MaxTokens=%d", s.Tier, s.MaxTokens)
		}
		if s.CostRank < 1 || s.CostRank > 3 {
			t.Errorf("strategy %q has CostRank=%d", s.Tier, s.CostRank)
		}
	}
	if !tiers[modelstrategy.TierSmall] || !tiers[modelstrategy.TierMedium] || !tiers[modelstrategy.TierLarge] {
		t.Errorf("missing tiers, got %v", tiers)
	}
}

func TestProviderRegistry(t *testing.T) {
	r := modelstrategy.NewProviderRegistry()

	// Register a provider
	p := &mockProvider{pt: modelstrategy.ProviderAnthropic}
	if err := r.Register(p); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Duplicate registration should fail
	if err := r.Register(p); err == nil {
		t.Fatal("expected error on duplicate registration")
	}

	// Get registered provider
	got, err := r.Get(modelstrategy.ProviderAnthropic)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.ProviderType() != modelstrategy.ProviderAnthropic {
		t.Errorf("got provider type %q", got.ProviderType())
	}

	// Get unregistered provider
	if _, err := r.Get("nonexistent"); err == nil {
		t.Fatal("expected error for unregistered provider")
	}

	// List
	list := r.List()
	if len(list) != 1 || list[0] != modelstrategy.ProviderAnthropic {
		t.Errorf("list = %v, want [anthropic]", list)
	}
}

func TestProviderRegistryNilProvider(t *testing.T) {
	r := modelstrategy.NewProviderRegistry()
	if err := r.Register(nil); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

func TestProviderRegistryEmptyType(t *testing.T) {
	r := modelstrategy.NewProviderRegistry()
	if err := r.Register(&mockProvider{pt: ""}); err == nil {
		t.Fatal("expected error for empty provider type")
	}
}

func TestStrategySelector(t *testing.T) {
	s := modelstrategy.NewStrategySelector(nil) // uses defaults
	// Select by classification
	strat, err := s.Select(modelstrategy.TaskExtraction)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if strat.Tier != modelstrategy.TierSmall {
		t.Errorf("tier = %q, want small", strat.Tier)
	}

	// SelectByTier
	strat, err = s.SelectByTier(modelstrategy.TierLarge)
	if err != nil {
		t.Fatalf("selectByTier: %v", err)
	}
	if strat.Tier != modelstrategy.TierLarge {
		t.Errorf("tier = %q, want large", strat.Tier)
	}

	// All
	all := s.All()
	if len(all) != 3 {
		t.Errorf("All() returned %d strategies", len(all))
	}
}

func TestStrategySelectorCustomStrategies(t *testing.T) {
	custom := []modelstrategy.ModelStrategy{
		{Tier: modelstrategy.TierSmall, Provider: modelstrategy.ProviderLocal, ModelName: "custom", MaxTokens: 4096, CostRank: 1},
	}
	s := modelstrategy.NewStrategySelector(custom)
	strat, err := s.SelectByTier(modelstrategy.TierSmall)
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if strat.ModelName != "custom" {
		t.Errorf("model = %q, want custom", strat.ModelName)
	}
	// Medium not in custom
	if _, err := s.SelectByTier(modelstrategy.TierMedium); err == nil {
		t.Fatal("expected error for missing tier")
	}
}

func TestRetryEngineConfig(t *testing.T) {
	cfg := modelstrategy.DefaultRetryConfig()
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.BaseDelay <= 0 {
		t.Errorf("BaseDelay = %v", cfg.BaseDelay)
	}
}

func TestRetryEngineSuccess(t *testing.T) {
	e := modelstrategy.NewRetryEngine(modelstrategy.DefaultRetryConfig())
	result, err := e.Execute(func() (string, error) {
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result != "ok" {
		t.Errorf("result = %q", result)
	}
}

func TestRetryEngineAllFailures(t *testing.T) {
	e := modelstrategy.NewRetryEngine(modelstrategy.DefaultRetryConfig())
	attempts := 0
	_, err := e.Execute(func() (string, error) {
		attempts++
		return "", errors.New("always fails")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 4 { // initial + 3 retries
		t.Errorf("attempts = %d, want 4", attempts)
	}
}

func TestRetryEngineSuccessAfterRetry(t *testing.T) {
	e := modelstrategy.NewRetryEngine(modelstrategy.RetryConfig{MaxRetries: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond})
	attempts := 0
	result, err := e.Execute(func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("transient error")
		}
		return "recovered", nil
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result != "recovered" {
		t.Errorf("result = %q", result)
	}
	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestValidateJSON(t *testing.T) {
	if err := modelstrategy.ValidateJSON(`{"a": 1}`); err != nil {
		t.Errorf("valid json: %v", err)
	}
	if err := modelstrategy.ValidateJSON(""); err == nil {
		t.Error("expected error for empty")
	}
	if err := modelstrategy.ValidateJSON("not json"); err == nil {
		t.Error("expected error for invalid json")
	}
}

func TestValidateFields(t *testing.T) {
	if err := modelstrategy.ValidateFields(`{"a": 1, "b": 2}`, []string{"a", "b"}); err != nil {
		t.Errorf("valid fields: %v", err)
	}
	if err := modelstrategy.ValidateFields(`{"a": 1}`, []string{"b"}); err == nil {
		t.Error("expected error for missing field")
	}
}

func TestIsRetryableError(t *testing.T) {
	if !modelstrategy.IsRetryableError(errors.New("invalid JSON")) {
		t.Error("invalid JSON should be retryable")
	}
	if !modelstrategy.IsRetryableError(errors.New("missing required fields")) {
		t.Error("missing fields should be retryable")
	}
	if !modelstrategy.IsRetryableError(errors.New("timeout")) {
		t.Error("timeout should be retryable")
	}
	if modelstrategy.IsRetryableError(errors.New("permanent error")) {
		t.Error("permanent error should not be retryable")
	}
	if modelstrategy.IsRetryableError(nil) {
		t.Error("nil should not be retryable")
	}
}

func TestContextBudgeterDefaults(t *testing.T) {
	b := modelstrategy.NewContextBudgeter(nil)
	budgets := b.All()
	if len(budgets) != 5 {
		t.Fatalf("got %d budgets, want 5", len(budgets))
	}
}

func TestContextBudgeterEstimate(t *testing.T) {
	b := modelstrategy.NewContextBudgeter(nil)
	bc, err := b.Estimate(modelstrategy.BudgetStandard)
	if err != nil {
		t.Fatalf("estimate: %v", err)
	}
	if bc.MaxTokens != 32768 {
		t.Errorf("MaxTokens = %d, want 32768", bc.MaxTokens)
	}
}

func TestContextBudgeterRecommendLevel(t *testing.T) {
	b := modelstrategy.NewContextBudgeter(nil)
	if l := b.RecommendLevel(1000); l != modelstrategy.BudgetMini {
		t.Errorf("level = %q, want mini", l)
	}
	if l := b.RecommendLevel(50000); l != modelstrategy.BudgetFull {
		t.Errorf("level = %q, want full", l)
	}
	if l := b.RecommendLevel(200000); l != modelstrategy.BudgetImplementation {
		t.Errorf("level = %q, want implementation", l)
	}
}

func TestContextBudgeterUnknownLevel(t *testing.T) {
	b := modelstrategy.NewContextBudgeter([]modelstrategy.BudgetConfig{
		{Level: modelstrategy.BudgetMini, MaxTokens: 1000, Description: "custom"},
	})
	_, err := b.Estimate("nonexistent")
	if err == nil {
		t.Fatal("expected error for unknown level")
	}
}

func TestNewService(t *testing.T) {
	s := modelstrategy.NewService()
	if s.Providers == nil {
		t.Error("Providers is nil")
	}
	if s.Selector == nil {
		t.Error("Selector is nil")
	}
	if s.Retry == nil {
		t.Error("Retry is nil")
	}
	if s.Budgeter == nil {
		t.Error("Budgeter is nil")
	}
}

func TestNewServiceWith(t *testing.T) {
	s := modelstrategy.NewServiceWith(nil, nil, nil, nil)
	if s.Providers == nil || s.Selector == nil || s.Retry == nil || s.Budgeter == nil {
		t.Error("NewServiceWith should initialize nil fields")
	}
}

func TestCompatibilityCatalogCheckSupported(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	r := c.Check("claude-sonnet-4-20250514", modelstrategy.ProviderAnthropic)
	if !r.Supported {
		t.Fatalf("expected supported, got note: %s", r.Note)
	}
	if r.MaxTokens <= 0 {
		t.Errorf("MaxTokens = %d", r.MaxTokens)
	}
	if r.Tier == "" {
		t.Error("Tier is empty")
	}
}

func TestCompatibilityCatalogCheckUnsupportedProvider(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	r := c.Check("claude-sonnet-4-20250514", modelstrategy.ProviderOpenAI)
	if r.Supported {
		t.Fatal("expected not supported")
	}
	if len(r.Alternatives) == 0 {
		t.Errorf("expected alternatives, got note: %s", r.Note)
	}
}

func TestCompatibilityCatalogCheckUnknownModel(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	r := c.Check("nonexistent-model-v99", modelstrategy.ProviderAnthropic)
	if r.Supported {
		t.Fatal("expected not supported")
	}
	if r.Alternatives != nil {
		t.Errorf("expected no alternatives for unknown model, got %v", r.Alternatives)
	}
}

func TestCompatibilityCatalogListModels(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	all := c.ListModels()
	if len(all) == 0 {
		t.Fatal("expected non-empty model list")
	}
	anthropic := c.ListModels(modelstrategy.ProviderAnthropic)
	if len(anthropic) == 0 {
		t.Fatal("expected anthropic models")
	}
	for _, e := range anthropic {
		if e.Provider != modelstrategy.ProviderAnthropic {
			t.Errorf("model %s has provider %q", e.Model, e.Provider)
		}
	}
}

func TestCompatibilityCatalogListProviders(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	providers := c.ListProviders()
	if len(providers) == 0 {
		t.Fatal("expected non-empty provider list")
	}
	seen := map[modelstrategy.ProviderType]bool{}
	for _, p := range providers {
		if seen[p] {
			t.Errorf("duplicate provider %q", p)
		}
		seen[p] = true
	}
	if !seen[modelstrategy.ProviderDeepSeek] {
		t.Error("expected DeepSeek provider")
	}
	if !seen[modelstrategy.ProviderQwen] {
		t.Error("expected Qwen provider")
	}
	if !seen[modelstrategy.ProviderOpenAICompatible] {
		t.Error("expected OpenAI-compatible provider")
	}
}

func TestCompatibilityCatalogIsModelKnown(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	if !c.IsModelKnown("gpt-4o") {
		t.Error("gpt-4o should be known")
	}
	if !c.IsModelKnown("deepseek-chat") {
		t.Error("deepseek-chat should be known")
	}
	if c.IsModelKnown("made-up-model-xyz") {
		t.Error("made-up model should not be known")
	}
}

func TestCompatibilityCatalogSupportsProvider(t *testing.T) {
	c := modelstrategy.NewCompatibilityCatalog()
	if !c.SupportsProvider(modelstrategy.ProviderDeepSeek) {
		t.Error("should support DeepSeek")
	}
	if !c.SupportsProvider(modelstrategy.ProviderQwen) {
		t.Error("should support Qwen")
	}
	if c.SupportsProvider("nonexistent") {
		t.Error("should not support nonexistent")
	}
}

func TestCompatibilityCatalogCustomEntries(t *testing.T) {
	entries := []modelstrategy.ModelEntry{
		{Model: "custom-model", Provider: modelstrategy.ProviderLocal, MaxTokens: 2048, Tier: modelstrategy.TierSmall},
	}
	c := modelstrategy.NewCompatibilityCatalogWith(entries)
	r := c.Check("custom-model", modelstrategy.ProviderLocal)
	if !r.Supported {
		t.Fatalf("expected supported, got: %s", r.Note)
	}
	if r.MaxTokens != 2048 {
		t.Errorf("MaxTokens = %d, want 2048", r.MaxTokens)
	}
	if r.Tier != modelstrategy.TierSmall {
		t.Errorf("Tier = %q, want small", r.Tier)
	}
}

type mockProvider struct {
	pt modelstrategy.ProviderType
}

func (m *mockProvider) Generate(_ modelstrategy.GenerateRequest) (modelstrategy.GenerateResponse, error) {
	return modelstrategy.GenerateResponse{}, nil
}

func (m *mockProvider) ProviderType() modelstrategy.ProviderType {
	return m.pt
}
