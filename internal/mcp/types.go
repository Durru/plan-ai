package mcp

// ToolDefinition describes an MCP tool that can be called.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Schema      JSONSchema      `json:"schema"`
	Handler     ToolHandlerFunc `json:"-"`
}

// ToolHandlerFunc is the signature for MCP tool handlers.
type ToolHandlerFunc func(args map[string]any) (*ToolResult, error)

// ToolResult is the standard response from a tool call.
type ToolResult struct {
	Success bool           `json:"success"`
	Content string         `json:"content,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// JSONSchema is a simplified schema for tool argument validation.
type JSONSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property describes a single field in a JSON Schema.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"-"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolContext carries dependencies into tool handlers.
type ToolContext struct {
	ProjectRoot string
	HomeRoot    string
	DBPath      string
}

// RunRecord tracks a tool execution for audit.
type RunRecord struct {
	ID        string `json:"id"`
	ToolName  string `json:"tool_name"`
	Arguments string `json:"arguments"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}
