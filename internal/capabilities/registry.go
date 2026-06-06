package capabilities

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// Registry stores and manages available capabilities.
type Registry struct {
	repo    *Repository
	db      *sql.DB
	mem     map[CapabilityType]Capability
}

// NewRegistry creates an empty capability registry backed by the given database.
// If db is nil, the registry operates in-memory.
func NewRegistry(db *sql.DB) *Registry {
	r := &Registry{db: db}
	if db != nil {
		r.repo = NewRepository(db)
	} else {
		r.mem = make(map[CapabilityType]Capability)
	}
	return r
}

// NewDefaultRegistry creates a registry with all standard capabilities seeded.
func NewDefaultRegistry(db *sql.DB) *Registry {
	r := NewRegistry(db)
	now := time.Now().UTC()
	defaults := []Capability{
		{Type: CapVision, Name: string(CapVision), Description: "Create vision drafts from ingested inputs", Enabled: true, CreatedAt: now},
		{Type: CapResearch, Name: string(CapResearch), Description: "Perform research on topics with sources and findings", Enabled: true, CreatedAt: now},
		{Type: CapPlanning, Name: string(CapPlanning), Description: "Create master plans, specific plans, and implementation documents", Enabled: true, CreatedAt: now},
		{Type: CapArchitecture, Name: string(CapArchitecture), Description: "Design system architecture and component relationships", Enabled: true, CreatedAt: now},
		{Type: CapDatabase, Name: string(CapDatabase), Description: "Design database schemas, migrations, and queries", Enabled: true, CreatedAt: now},
		{Type: CapBackend, Name: string(CapBackend), Description: "Implement server-side logic and APIs", Enabled: true, CreatedAt: now},
		{Type: CapFrontend, Name: string(CapFrontend), Description: "Implement client-side interfaces and interactions", Enabled: true, CreatedAt: now},
		{Type: CapSecurity, Name: string(CapSecurity), Description: "Review code and architecture for security issues", Enabled: true, CreatedAt: now},
		{Type: CapTesting, Name: string(CapTesting), Description: "Write and execute tests for verification", Enabled: true, CreatedAt: now},
		{Type: CapImpactAnalysis, Name: string(CapImpactAnalysis), Description: "Analyze impact of changes on plans and decisions", Enabled: true, CreatedAt: now},
		{Type: CapValidation, Name: string(CapValidation), Description: "Validate plans, tasks, and implementations against criteria", Enabled: true, CreatedAt: now},
	}
	for _, cap := range defaults {
		_ = r.RegisterCapability(cap)
	}
	return r
}

// RegisterCapability adds a capability to the registry.
func (r *Registry) RegisterCapability(c Capability) error {
	if c.Type == "" {
		return fmt.Errorf("capability type is required")
	}
	if c.Name == "" {
		c.Name = string(c.Type)
	}
	if r.repo != nil {
		return r.repo.Save(c)
	}
	r.mem[c.Type] = c
	return nil
}

// GetCapability returns a capability by type.
func (r *Registry) GetCapability(ct CapabilityType) (Capability, error) {
	if r.repo != nil {
		return r.repo.GetByName(string(ct))
	}
	c, ok := r.mem[ct]
	if !ok {
		return Capability{}, fmt.Errorf("capability %q not found", ct)
	}
	return c, nil
}

// ListCapabilities returns all registered capabilities sorted by type.
func (r *Registry) ListCapabilities() []Capability {
	if r.repo != nil {
		result, err := r.repo.List()
		if err != nil {
			return nil
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i].Type < result[j].Type
		})
		return result
	}
	result := make([]Capability, 0, len(r.mem))
	for _, c := range r.mem {
		result = append(result, c)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Type < result[j].Type
	})
	return result
}
