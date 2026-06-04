package modelstrategy

// ──────────────────────────────────────────────
// Context budget levels
// ──────────────────────────────────────────────

type BudgetLevel string

const (
	BudgetMini           BudgetLevel = "mini"
	BudgetShort          BudgetLevel = "short"
	BudgetStandard       BudgetLevel = "standard"
	BudgetFull           BudgetLevel = "full"
	BudgetImplementation BudgetLevel = "implementation"
)

// BudgetConfig describes the token budget for a context level.
type BudgetConfig struct {
	Level       BudgetLevel
	MaxTokens   int
	Description string
}

// DefaultBudgets returns recommended token budgets per level.
func DefaultBudgets() []BudgetConfig {
	return []BudgetConfig{
		{Level: BudgetMini, MaxTokens: 2048, Description: "Quick lookup or status check"},
		{Level: BudgetShort, MaxTokens: 8192, Description: "Small extraction or classification"},
		{Level: BudgetStandard, MaxTokens: 32768, Description: "Research, planning, or validation"},
		{Level: BudgetFull, MaxTokens: 65536, Description: "Complex architecture or deep research"},
		{Level: BudgetImplementation, MaxTokens: 131072, Description: "Full implementation tasks"},
	}
}

// ──────────────────────────────────────────────
// Context budgeter
// ──────────────────────────────────────────────

// ContextBudgeter estimates and returns the appropriate budget for a context.
type ContextBudgeter struct {
	budgets []BudgetConfig
}

// NewContextBudgeter creates a budgeter with the given budgets.
func NewContextBudgeter(budgets []BudgetConfig) *ContextBudgeter {
	if len(budgets) == 0 {
		budgets = DefaultBudgets()
	}
	return &ContextBudgeter{budgets: budgets}
}

// Estimate returns the budget for a given level.
func (b *ContextBudgeter) Estimate(level BudgetLevel) (BudgetConfig, error) {
	for _, bc := range b.budgets {
		if bc.Level == level {
			return bc, nil
		}
	}
	return BudgetConfig{}, ErrBudgetLevelNotFound{Level: level}
}

// RecommendLevel recommends a budget level based on estimated content size.
func (b *ContextBudgeter) RecommendLevel(contentSize int) BudgetLevel {
	for _, bc := range b.budgets {
		if contentSize <= bc.MaxTokens {
			return bc.Level
		}
	}
	return BudgetImplementation
}

// All returns all configured budgets.
func (b *ContextBudgeter) All() []BudgetConfig {
	return b.budgets
}

// ErrBudgetLevelNotFound is returned when a budget level is unknown.
type ErrBudgetLevelNotFound struct {
	Level BudgetLevel
}

func (e ErrBudgetLevelNotFound) Error() string {
	return "budget level not found: " + string(e.Level)
}
