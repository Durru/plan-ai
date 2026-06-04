package modelstrategy

// Service provides the top-level model strategy operations.
type Service struct {
	Providers *ProviderRegistry
	Selector  *StrategySelector
	Retry     *RetryEngine
	Budgeter  *ContextBudgeter
}

// NewService creates a new model strategy service with defaults.
func NewService() *Service {
	return &Service{
		Providers: NewProviderRegistry(),
		Selector:  NewStrategySelector(nil),
		Retry:     NewRetryEngine(DefaultRetryConfig()),
		Budgeter:  NewContextBudgeter(nil),
	}
}

// NewServiceWith creates a service with explicit dependencies.
func NewServiceWith(
	providers *ProviderRegistry,
	selector *StrategySelector,
	retry *RetryEngine,
	budgeter *ContextBudgeter,
) *Service {
	if providers == nil {
		providers = NewProviderRegistry()
	}
	if selector == nil {
		selector = NewStrategySelector(nil)
	}
	if retry == nil {
		retry = NewRetryEngine(DefaultRetryConfig())
	}
	if budgeter == nil {
		budgeter = NewContextBudgeter(nil)
	}
	return &Service{
		Providers: providers,
		Selector:  selector,
		Retry:     retry,
		Budgeter:  budgeter,
	}
}
