package capabilities

import (
	"database/sql"
	"fmt"
	"os"
	"sort"
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
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM capabilities_v2").Scan(&count); err == nil && count == 0 {
			if err := SeedDefaults(db); err != nil {
				fmt.Fprintf(os.Stderr, "warning: seed defaults: %v\n", err)
			}
		}
	} else {
		r.mem = make(map[CapabilityType]Capability)
	}
	return r
}

// NewDefaultRegistry creates a registry with all standard capabilities seeded.
func NewDefaultRegistry(db *sql.DB) *Registry {
	r := NewRegistry(db)
	if db != nil {
		if err := SeedDefaults(db); err != nil {
			fmt.Fprintf(os.Stderr, "warning: seed defaults: %v\n", err)
		}
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
