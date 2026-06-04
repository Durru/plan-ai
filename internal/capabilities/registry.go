package capabilities

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Registry stores and manages available capabilities.
type Registry struct {
	mu           sync.RWMutex
	capabilities map[CapabilityType]Capability
}

// NewRegistry creates an empty capability registry.
func NewRegistry() *Registry {
	return &Registry{capabilities: make(map[CapabilityType]Capability)}
}

// NewDefaultRegistry creates a registry with all standard capabilities registered.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	now := time.Now().UTC()
	defaults := []Capability{
		{Type: CapVision, Name: "Vision Drafting", Description: "Create vision drafts from ingested inputs", CreatedAt: now},
		{Type: CapResearch, Name: "Research Execution", Description: "Perform research on topics with sources and findings", CreatedAt: now},
		{Type: CapPlanning, Name: "Planning", Description: "Create master plans, specific plans, and implementation documents", CreatedAt: now},
		{Type: CapArchitecture, Name: "Architecture Design", Description: "Design system architecture and component relationships", CreatedAt: now},
		{Type: CapDatabase, Name: "Database Design", Description: "Design database schemas, migrations, and queries", CreatedAt: now},
		{Type: CapBackend, Name: "Backend Development", Description: "Implement server-side logic and APIs", CreatedAt: now},
		{Type: CapFrontend, Name: "Frontend Development", Description: "Implement client-side interfaces and interactions", CreatedAt: now},
		{Type: CapSecurity, Name: "Security Review", Description: "Review code and architecture for security issues", CreatedAt: now},
		{Type: CapTesting, Name: "Testing", Description: "Write and execute tests for verification", CreatedAt: now},
		{Type: CapImpactAnalysis, Name: "Impact Analysis", Description: "Analyze impact of changes on plans and decisions", CreatedAt: now},
		{Type: CapValidation, Name: "Validation", Description: "Validate plans, tasks, and implementations against criteria", CreatedAt: now},
	}
	for _, cap := range defaults {
		r.capabilities[cap.Type] = cap
	}
	return r
}

// RegisterCapability adds a capability to the registry.
func (r *Registry) RegisterCapability(c Capability) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c.Type == "" {
		return fmt.Errorf("capability type is required")
	}
	if c.Name == "" {
		return fmt.Errorf("capability name is required")
	}
	r.capabilities[c.Type] = c
	return nil
}

// GetCapability returns a capability by type.
func (r *Registry) GetCapability(ct CapabilityType) (Capability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.capabilities[ct]
	if !ok {
		return Capability{}, fmt.Errorf("capability %q not found", ct)
	}
	return c, nil
}

// ListCapabilities returns all registered capabilities sorted by type.
func (r *Registry) ListCapabilities() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Capability, 0, len(r.capabilities))
	for _, c := range r.capabilities {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})
	return result
}
