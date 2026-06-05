package mcp

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(ToolContext{ProjectRoot: "/tmp/test"})
	if s == nil {
		t.Fatal("expected non-nil server")
	}
	tools := s.ListTools()
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestRegisterAndListTools(t *testing.T) {
	s := NewServer(ToolContext{})
	err := s.RegisterTool(ToolDefinition{
		Name:        "test_tool",
		Description: "A test tool",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]Property{
				"name": {Type: "string", Description: "A name"},
			},
			Required: []string{"name"},
		},
		Handler: func(args map[string]any) (*ToolResult, error) {
			return &ToolResult{Success: true, Content: "hello"}, nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tools := s.ListTools()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Name != "test_tool" {
		t.Fatalf("expected 'test_tool', got %q", tools[0].Name)
	}
}

func TestRegisterTool_NoName(t *testing.T) {
	s := NewServer(ToolContext{})
	err := s.RegisterTool(ToolDefinition{
		Handler: func(args map[string]any) (*ToolResult, error) {
			return &ToolResult{Success: true}, nil
		},
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterTool_NoHandler(t *testing.T) {
	s := NewServer(ToolContext{})
	err := s.RegisterTool(ToolDefinition{Name: "noop"})
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestExecuteTool(t *testing.T) {
	s := NewServer(ToolContext{})
	s.RegisterTool(ToolDefinition{
		Name: "echo",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]Property{
				"msg": {Type: "string", Description: "Message to echo"},
			},
			Required: []string{"msg"},
		},
		Handler: func(args map[string]any) (*ToolResult, error) {
			msg, _ := args["msg"].(string)
			return &ToolResult{Success: true, Content: msg}, nil
		},
	})

	result := s.ExecuteTool("echo", map[string]any{"msg": "hello"})
	if !result.Success {
		t.Fatalf("expected success, got error: %s", result.Error)
	}
	if result.Content != "hello" {
		t.Fatalf("expected 'hello', got %q", result.Content)
	}
}

func TestExecuteTool_Unknown(t *testing.T) {
	s := NewServer(ToolContext{})
	result := s.ExecuteTool("nonexistent", nil)
	if result.Success {
		t.Fatal("expected failure for unknown tool")
	}
	if result.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestExecuteTool_MissingArg(t *testing.T) {
	s := NewServer(ToolContext{})
	s.RegisterTool(ToolDefinition{
		Name: "requires_name",
		Schema: JSONSchema{
			Type: "object",
			Properties: map[string]Property{
				"name": {Type: "string", Description: "Required name"},
			},
			Required: []string{"name"},
		},
		Handler: func(args map[string]any) (*ToolResult, error) {
			return &ToolResult{Success: true}, nil
		},
	})

	result := s.ExecuteTool("requires_name", map[string]any{})
	if result.Success {
		t.Fatal("expected failure for missing arg")
	}
}

func TestGetTool(t *testing.T) {
	s := NewServer(ToolContext{})
	s.RegisterTool(ToolDefinition{
		Name:    "my_tool",
		Handler: func(args map[string]any) (*ToolResult, error) { return &ToolResult{Success: true}, nil },
	})

	td, ok := s.GetTool("my_tool")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if td.Name != "my_tool" {
		t.Fatalf("expected 'my_tool', got %q", td.Name)
	}

	_, ok = s.GetTool("unknown")
	if ok {
		t.Fatal("expected tool to not be found")
	}
}

func TestGetRuns(t *testing.T) {
	s := NewServer(ToolContext{})
	s.RegisterTool(ToolDefinition{
		Name: "test",
		Handler: func(args map[string]any) (*ToolResult, error) {
			return &ToolResult{Success: true}, nil
		},
	})

	s.ExecuteTool("test", nil)
	s.ExecuteTool("test", nil)

	runs := s.GetRuns(10)
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	if runs[0].ToolName != "test" {
		t.Fatalf("expected 'test', got %q", runs[0].ToolName)
	}
}

func TestRegisterDefaultTools(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:         func(args map[string]any) (map[string]any, error) { return map[string]any{"ok": true}, nil },
		ProjectStatus:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentRuns:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousStatus:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousEvents:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousContext:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
	}

	err := RegisterDefaultTools(s, deps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tools := s.ListTools()
	if len(tools) != 38 {
		t.Fatalf("expected 38 default tools, got %d", len(tools))
	}

	for _, name := range []string{
		"plan_ai.agent_message",
		"plan_ai.agent_status",
		"plan_ai.continuous_status",
		"plan_ai.propose_plan_update",
		"plan_ai.approve_plan_update",
		"plan_ai.reject_plan_update",
		"plan_ai.get_context_level",
	} {
		if _, ok := s.GetTool(name); !ok {
			t.Fatalf("required tool %s not registered", name)
		}
	}
}

func TestPhase21And22RequiredToolsExecute(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ProjectStatus:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:     func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"message": args["message"]}, nil
		},
		AgentRuns:        func(args map[string]any) (map[string]any, error) { return map[string]any{"status": "ok"}, nil },
		ContinuousStatus: func(args map[string]any) (map[string]any, error) { return map[string]any{"status": "ok"}, nil },
		ContinuousEvents: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"proposal": args["proposal_id"]}, nil
		},
		ContinuousContext: func(args map[string]any) (map[string]any, error) { return map[string]any{"level": args["level"]}, nil },
		GetContext:        func(args map[string]any) (map[string]any, error) { return map[string]any{"level": args["level"]}, nil },
		DetectChanges:     func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		UpdatePlan:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RollbackSnapshot:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ExportDocs:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
	}
	if err := RegisterDefaultTools(s, deps); err != nil {
		t.Fatalf("RegisterDefaultTools: %v", err)
	}

	cases := []struct {
		name string
		args map[string]any
	}{
		{"plan_ai.agent_message", map[string]any{"message": "create a plan"}},
		{"plan_ai.agent_status", map[string]any{}},
		{"plan_ai.continuous_status", map[string]any{}},
		{"plan_ai.propose_plan_update", map[string]any{"reason": "new requirement"}},
		{"plan_ai.approve_plan_update", map[string]any{"proposal_id": "proposal-1"}},
		{"plan_ai.reject_plan_update", map[string]any{"proposal_id": "proposal-1"}},
		{"plan_ai.get_context_level", map[string]any{"level": "L0_Executive"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ExecuteTool(tt.name, tt.args)
			if !result.Success {
				t.Fatalf("expected %s success, got %s", tt.name, result.Error)
			}
		})
	}
}

func TestPhase29RequiredToolsExecute(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ProjectStatus:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:     func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentRuns:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousStatus:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousEvents:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"proposals": []string{}}, nil
		},
		ContinuousContext: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetContext:        func(args map[string]any) (map[string]any, error) { return map[string]any{"level": args["level"]}, nil },
		DetectChanges:     func(args map[string]any) (map[string]any, error) { return map[string]any{"change_id": "chg_1"}, nil },
		UpdatePlan: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"plan_id": args["plan_id"]}, nil
		},
		RollbackSnapshot: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"supported": false, "message": "rollback not yet implemented"}, nil
		},
		ExportDocs: func(args map[string]any) (map[string]any, error) {
			return map[string]any{"format": args["format"], "scope": args["scope"]}, nil
		},
	}
	if err := RegisterDefaultTools(s, deps); err != nil {
		t.Fatalf("RegisterDefaultTools: %v", err)
	}

	cases := []struct {
		name string
		args map[string]any
	}{
		{"plan_ai.get_context", map[string]any{"level": "L1_planning"}},
		{"plan_ai.detect_changes", map[string]any{"change_type": "plan_changed", "summary": "Updated scope"}},
		{"plan_ai.update_plan", map[string]any{"plan_id": "plan-1", "title": "New Title"}},
		{"plan_ai.rollback_snapshot", map[string]any{"snapshot_id": "snap-1"}},
		{"plan_ai.export_docs", map[string]any{"format": "markdown", "scope": "all"}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := s.ExecuteTool(tt.name, tt.args)
			if !result.Success {
				t.Fatalf("expected %s success, got %s", tt.name, result.Error)
			}
		})
	}
}

func TestPhase29GetContext_RequiresLevel(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ProjectStatus:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentRuns:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousStatus:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousEvents:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousContext:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetContext:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		DetectChanges:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		UpdatePlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RollbackSnapshot:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ExportDocs:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
	}
	RegisterDefaultTools(s, deps)

	// Missing required level
	result := s.ExecuteTool("plan_ai.get_context", map[string]any{})
	if result.Success {
		t.Fatal("expected validation failure for missing level")
	}
}

func TestPhase29ExportDocs_RequiresFormatAndScope(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ProjectStatus:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentRuns:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousStatus:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousEvents:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousContext:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetContext:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		DetectChanges:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		UpdatePlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RollbackSnapshot:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ExportDocs:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
	}
	RegisterDefaultTools(s, deps)

	// Missing required format
	result := s.ExecuteTool("plan_ai.export_docs", map[string]any{"scope": "all"})
	if result.Success {
		t.Fatal("expected validation failure for missing format")
	}

	// Missing required scope
	result = s.ExecuteTool("plan_ai.export_docs", map[string]any{"format": "markdown"})
	if result.Success {
		t.Fatal("expected validation failure for missing scope")
	}
}

func TestValidateArgs(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]Property{
			"name": {Type: "string", Description: "Name"},
		},
		Required: []string{"name"},
	}

	// Valid
	err := ValidateArgs(schema, map[string]any{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Missing required
	err = ValidateArgs(schema, map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing required field")
	}

	// Wrong type
	err = ValidateArgs(schema, map[string]any{"name": 42})
	if err == nil {
		t.Fatal("expected error for wrong type")
	}
}

func TestServer_MinimalMode(t *testing.T) {
	s := NewServer(ToolContext{})
	// Register a full tool and a minimal tool
	s.RegisterTool(ToolDefinition{
		Name:        "plan_ai.full_tool",
		Description: "A full tool not in minimal set",
		Schema:      JSONSchema{Type: "object"},
		Handler:     func(args map[string]any) (*ToolResult, error) { return &ToolResult{Success: true}, nil },
	})
	s.RegisterTool(ToolDefinition{
		Name:        "plan_ai.project_status",
		Description: "Get project status",
		Schema:      JSONSchema{Type: "object"},
		Handler:     func(args map[string]any) (*ToolResult, error) { return &ToolResult{Success: true}, nil },
	})
	s.RegisterTool(ToolDefinition{
		Name:        "plan_ai.get_context",
		Description: "Get context",
		Schema:      JSONSchema{Type: "object"},
		Handler:     func(args map[string]any) (*ToolResult, error) { return &ToolResult{Success: true}, nil },
	})

	// Without minimal mode — all 3 tools
	all := s.ListTools()
	if len(all) != 3 {
		t.Fatalf("expected 3 tools without minimal mode, got %d", len(all))
	}

	// With minimal mode — only minimal tools
	s.SetMinimalMode(true)
	minimal := s.ListTools()
	if len(minimal) != 2 {
		t.Fatalf("expected 2 minimal tools, got %d: %v", len(minimal), toolNames(minimal))
	}
	// Verify only minimal tools returned
	for _, td := range minimal {
		if !MinimalToolNames[td.Name] {
			t.Fatalf("non-minimal tool %q returned in minimal mode", td.Name)
		}
	}

	// ExecuteTool still works for all tools regardless of mode
	result := s.ExecuteTool("plan_ai.full_tool", nil)
	if !result.Success {
		t.Fatalf("ExecuteTool should work for non-minimal tool in minimal mode: %s", result.Error)
	}
}

func toolNames(tools []ToolDefinition) []string {
	names := make([]string, len(tools))
	for i, td := range tools {
		names[i] = td.Name
	}
	return names
}

func TestValidateTools(t *testing.T) {
	tools := []ToolDefinition{
		{
			Name:        "plan_ai.valid_tool",
			Description: "A valid tool",
			Schema:      JSONSchema{Type: "object", Properties: map[string]Property{}},
		},
		{
			Name:        "plan_ai.missing_description",
			Description: "",
			Schema:      JSONSchema{Type: "object"},
		},
		{
			Name:        "",
			Description: "Empty name",
			Schema:      JSONSchema{Type: "object"},
		},
		{
			Name:        "plan_ai.bad_schema_type",
			Description: "Schema without type",
			Schema:      JSONSchema{Type: ""},
		},
		{
			Name:        "plan_ai.duplicate",
			Description: "First duplicate",
			Schema:      JSONSchema{Type: "object"},
		},
		{
			Name:        "plan_ai.duplicate",
			Description: "Second duplicate",
			Schema:      JSONSchema{Type: "object"},
		},
	}

	results := ValidateTools(tools)
	if len(results) != len(tools) {
		t.Fatalf("expected %d results, got %d", len(tools), len(results))
	}

	// valid_tool should be valid
	for _, r := range results {
		if r.Name == "plan_ai.valid_tool" && !r.Valid {
			t.Fatalf("valid_tool should be valid: %v", r.Issues)
		}
	}

	// missing_description should be invalid
	found := false
	for _, r := range results {
		if r.Name == "plan_ai.missing_description" && !r.Valid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("missing_description should be invalid")
	}

	// empty name should be invalid
	found = false
	for _, r := range results {
		if r.Name == "" && !r.Valid {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("empty name should be invalid")
	}

	// duplicates should be invalid
	count := 0
	for _, r := range results {
		if r.Name == "plan_ai.duplicate" {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 duplicate results, got %d", count)
	}

	// ValidateAllTools should return false
	allValid, _ := ValidateAllTools(tools)
	if allValid {
		t.Fatal("ValidateAllTools should return false for invalid tools")
	}
}

func TestValidateTools_AllValid(t *testing.T) {
	s := NewServer(ToolContext{})
	deps := &ToolDependencies{
		InitProject:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ProjectStatus:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateMasterPlan:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSpecificPlan:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ResearchTopic:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApprovePlan:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RejectPlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AnalyzeImpact:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetNextTask:         func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		MarkTaskDone:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateSnapshot:      func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListPlans:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListTasks:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentProcess:        func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		AgentRuns:           func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousStatus:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousEvents:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousProposals: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ContinuousContext:   func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetContext:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		DetectChanges:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		UpdatePlan:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		RollbackSnapshot:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ExportDocs:          func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		CreateProductIntent: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListProductIntents:  func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		GetProductIntent:    func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		SubmitProductIntent: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ApproveProductIntent: func(args map[string]any) (map[string]any, error) {
			return map[string]any{}, nil
		},
		RejectProductIntent: func(args map[string]any) (map[string]any, error) {
			return map[string]any{}, nil
		},
		DiscoverIntent:       func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
		ListDiscoveryResults: func(args map[string]any) (map[string]any, error) { return map[string]any{}, nil },
	}
	RegisterDefaultTools(s, deps)

	allValid, results := ValidateAllTools(s.ListTools())
	if !allValid {
		t.Errorf("all registered tools should be valid, got failures:")
		for _, r := range results {
			if !r.Valid {
				t.Errorf("  %s: %v", r.Name, r.Issues)
			}
		}
	}
}

func TestValidateArgs_Enum(t *testing.T) {
	schema := JSONSchema{
		Type: "object",
		Properties: map[string]Property{
			"mode": {
				Type:        "string",
				Description: "Mode",
				Enum:        []string{"a", "b", "c"},
			},
		},
		Required: []string{"mode"},
	}

	err := ValidateArgs(schema, map[string]any{"mode": "a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = ValidateArgs(schema, map[string]any{"mode": "z"})
	if err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}
