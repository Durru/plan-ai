package modelstrategy

import "fmt"

// ProviderRegistry holds registered model providers by type.
type ProviderRegistry struct {
	providers map[ProviderType]ModelProvider
}

// NewProviderRegistry creates an empty registry.
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{providers: make(map[ProviderType]ModelProvider)}
}

// Register adds a provider to the registry.
func (r *ProviderRegistry) Register(provider ModelProvider) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}
	pt := provider.ProviderType()
	if pt == "" {
		return fmt.Errorf("provider type is empty")
	}
	if _, exists := r.providers[pt]; exists {
		return fmt.Errorf("provider %q already registered", pt)
	}
	r.providers[pt] = provider
	return nil
}

// Get returns a registered provider by type.
func (r *ProviderRegistry) Get(pt ProviderType) (ModelProvider, error) {
	p, ok := r.providers[pt]
	if !ok {
		return nil, fmt.Errorf("provider %q not registered", pt)
	}
	return p, nil
}

// List returns all registered provider types.
func (r *ProviderRegistry) List() []ProviderType {
	types := make([]ProviderType, 0, len(r.providers))
	for pt := range r.providers {
		types = append(types, pt)
	}
	return types
}
