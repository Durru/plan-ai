package modelstrategy

import "fmt"

// StrategySelector picks the appropriate model strategy for a task.
type StrategySelector struct {
	strategies []ModelStrategy
}

// NewStrategySelector creates a selector with the given strategies.
func NewStrategySelector(strategies []ModelStrategy) *StrategySelector {
	if len(strategies) == 0 {
		strategies = DefaultModelStrategies()
	}
	return &StrategySelector{strategies: strategies}
}

// Select returns the strategy for the given classification.
func (s *StrategySelector) Select(classification TaskClassification) (ModelStrategy, error) {
	tier := ClassificationTier(classification)
	for _, strat := range s.strategies {
		if strat.Tier == tier {
			return strat, nil
		}
	}
	return ModelStrategy{}, fmt.Errorf("no strategy found for tier %q (classification: %s)", tier, classification)
}

// SelectByTier returns the strategy for the given tier directly.
func (s *StrategySelector) SelectByTier(tier ModelTier) (ModelStrategy, error) {
	for _, strat := range s.strategies {
		if strat.Tier == tier {
			return strat, nil
		}
	}
	return ModelStrategy{}, fmt.Errorf("no strategy found for tier %q", tier)
}

// All returns all configured strategies.
func (s *StrategySelector) All() []ModelStrategy {
	return s.strategies
}
