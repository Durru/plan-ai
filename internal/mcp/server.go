package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"
)

// Server is the MCP tool registry and execution engine.
type Server struct {
	mu      sync.RWMutex
	tools   map[string]ToolDefinition
	records []RunRecord
	ctx     ToolContext
}

// NewServer creates an MCP server with the given tool context.
func NewServer(ctx ToolContext) *Server {
	return &Server{
		tools:   make(map[string]ToolDefinition),
		records: []RunRecord{},
		ctx:     ctx,
	}
}

// RegisterTool adds a tool definition to the server.
func (s *Server) RegisterTool(td ToolDefinition) error {
	if td.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	if td.Handler == nil {
		return fmt.Errorf("tool %q has no handler", td.Name)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[td.Name] = td
	return nil
}

// GetTool returns a registered tool definition.
func (s *Server) GetTool(name string) (ToolDefinition, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	td, ok := s.tools[name]
	return td, ok
}

// ListTools returns all registered tool definitions sorted by name.
func (s *Server) ListTools() []ToolDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]ToolDefinition, 0, len(s.tools))
	for _, td := range s.tools {
		out = append(out, td)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ExecuteTool runs a registered tool with validated arguments.
func (s *Server) ExecuteTool(name string, args map[string]any) *ToolResult {
	td, ok := s.GetTool(name)
	if !ok {
		return &ToolResult{Success: false, Error: fmt.Sprintf("unknown tool: %s", name)}
	}

	// Validate
	if err := ValidateArgs(td.Schema, args); err != nil {
		result := &ToolResult{Success: false, Error: err.Error()}
		s.recordRun(name, args, result)
		return result
	}

	// Execute
	result, err := td.Handler(args)
	if err != nil {
		result = &ToolResult{Success: false, Error: err.Error()}
	}

	if result == nil {
		result = &ToolResult{Success: true}
	}

	s.recordRun(name, args, result)
	return result
}

// GetRuns returns the recorded tool execution history.
func (s *Server) GetRuns(limit int) []RunRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 || limit > len(s.records) {
		limit = len(s.records)
	}
	out := make([]RunRecord, limit)
	copy(out, s.records[len(s.records)-limit:])
	return out
}

func (s *Server) recordRun(name string, args map[string]any, result *ToolResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	argBytes, _ := json.Marshal(args)
	errMsg := ""
	if !result.Success {
		errMsg = result.Error
	}

	record := RunRecord{
		ID:        fmt.Sprintf("run_%d", time.Now().UTC().UnixNano()),
		ToolName:  name,
		Arguments: string(argBytes),
		Success:   result.Success,
		Error:     errMsg,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.records = append(s.records, record)
}
