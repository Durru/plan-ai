package mcp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

var validToolNameRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)*$`)

// ToolValidationResult holds validation details for a single tool.
type ToolValidationResult struct {
	Name     string   `json:"name"`
	Valid    bool     `json:"valid"`
	Issues   []string `json:"issues,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ValidateTools validates a set of tool definitions for OpenCode MCP compatibility.
func ValidateTools(tools []ToolDefinition) []ToolValidationResult {
	results := make([]ToolValidationResult, 0, len(tools))
	names := make(map[string]bool)

	for _, tool := range tools {
		r := ToolValidationResult{Name: tool.Name, Valid: true}

		// Check for duplicate names
		if names[tool.Name] {
			r.Issues = append(r.Issues, "duplicate tool name")
			r.Valid = false
		}
		names[tool.Name] = true

		// Tool name must be valid
		if tool.Name == "" {
			r.Issues = append(r.Issues, "tool name is empty")
			r.Valid = false
		} else if !validToolNameRegexp.MatchString(tool.Name) {
			r.Issues = append(r.Issues, fmt.Sprintf("tool name %q does not match pattern %s", tool.Name, validToolNameRegexp.String()))
			r.Valid = false
		}

		// Description must not be empty
		if strings.TrimSpace(tool.Description) == "" {
			r.Issues = append(r.Issues, "description is empty")
			r.Valid = false
		}

		// inputSchema must have type "object"
		if tool.Schema.Type != "object" {
			r.Issues = append(r.Issues, fmt.Sprintf("inputSchema.type is %q, want \"object\"", tool.Schema.Type))
			r.Valid = false
		}

		// inputSchema.properties must be a map (nil is OK, serializes as {})
		if tool.Schema.Properties == nil {
			r.Warnings = append(r.Warnings, "inputSchema.properties is nil (will serialize as {})")
		}

		// Check for malformed enums (empty enum arrays with no type)
		for key, prop := range tool.Schema.Properties {
			if prop.Type == "" && prop.Description == "" {
				r.Warnings = append(r.Warnings, fmt.Sprintf("property %q has empty type and description", key))
			}
			if len(prop.Enum) == 0 && prop.Type == "" {
				r.Warnings = append(r.Warnings, fmt.Sprintf("property %q has no type and no enum", key))
			}
		}

		// Check required fields reference existing properties
		for _, req := range tool.Schema.Required {
			if _, ok := tool.Schema.Properties[req]; !ok {
				r.Warnings = append(r.Warnings, fmt.Sprintf("required field %q is not defined in properties", req))
			}
		}

		// JSON serializable check
		if _, err := json.Marshal(tool); err != nil {
			r.Issues = append(r.Issues, fmt.Sprintf("not JSON serializable: %v", err))
			r.Valid = false
		}

		results = append(results, r)
	}

	return results
}

// ValidateAllTools checks all registered tools and returns a summary.
// This is used by the validate-tools CLI command.
func ValidateAllTools(tools []ToolDefinition) (allValid bool, results []ToolValidationResult) {
	results = ValidateTools(tools)
	allValid = true
	for _, r := range results {
		if !r.Valid {
			allValid = false
		}
	}
	return
}
