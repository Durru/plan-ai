package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// RegistryItem describes a Plan-AI tool that can be exposed via OpenCode MCP.
type RegistryItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Schema      any    `json:"schema"`
	PlansBuilt  int    `json:"plans_built"`
}

// ToolRegistry maintains the list of MCP tools Plan-AI exposes to OpenCode.
type ToolRegistry struct {
	items []RegistryItem
	path  string
}

// NewToolRegistry creates a tool registry stored at the given path.
func NewToolRegistry(homeRoot string) *ToolRegistry {
	return &ToolRegistry{
		items: []RegistryItem{},
		path:  filepath.Join(homeRoot, "mcp-registry.json"),
	}
}

// Add registers a new tool in the registry.
func (r *ToolRegistry) Add(item RegistryItem) {
	for i, existing := range r.items {
		if existing.Name == item.Name {
			r.items[i] = item
			return
		}
	}
	r.items = append(r.items, item)
}

// List returns all registered tools sorted by name.
func (r *ToolRegistry) List() []RegistryItem {
	sort.Slice(r.items, func(i, j int) bool {
		return r.items[i].Name < r.items[j].Name
	})
	return r.items
}

// Save persists the registry to disk.
func (r *ToolRegistry) Save() error {
	data, err := json.MarshalIndent(r.items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	return os.WriteFile(r.path, data, 0644)
}

// Load reads the registry from disk.
func (r *ToolRegistry) Load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read registry: %w", err)
	}
	return json.Unmarshal(data, &r.items)
}

// IncrementPlanCount increments the PlansBuilt counter for a named tool.
func (r *ToolRegistry) IncrementPlanCount(name string) {
	for i, item := range r.items {
		if item.Name == name {
			r.items[i].PlansBuilt++
			return
		}
	}
}
