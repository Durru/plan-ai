package mcp

import (
	"encoding/json"
	"fmt"

	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

// allToolDefs returns all standard Plan-AI MCP tool definitions.
func allToolDefs(deps *ToolDependencies) []ToolDefinition {
	tools := []ToolDefinition{
		{
			Name:        "plan_ai.init_project",
			Description: "Initialize Plan-AI for a project at the given root path.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Absolute path to the project root", Required: true},
				},
				Required: []string{"project_root"},
			},
			Handler: makeHandler(deps, deps.InitProject),
		},
		{
			Name:        "plan_ai.project_status",
			Description: "Get the current status of a Plan-AI project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Absolute path to the project root (default: current directory)"},
				},
			},
			Handler: makeHandler(deps, deps.ProjectStatus),
		},
		{
			Name:        "plan_ai.create_master_plan",
			Description: "Create a new master plan from approved context.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"title":        {Type: "string", Description: "Master plan title", Required: true},
					"summary":      {Type: "string", Description: "Master plan summary", Required: true},
				},
				Required: []string{"title", "summary"},
			},
			Handler: makeHandler(deps, deps.CreateMasterPlan),
		},
		{
			Name:        "plan_ai.create_specific_plan",
			Description: "Create a specific plan under a master plan.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"master_plan_id": {Type: "string", Description: "Parent master plan ID", Required: true},
					"title":          {Type: "string", Description: "Specific plan title", Required: true},
					"goal":           {Type: "string", Description: "Plan goal"},
				},
				Required: []string{"master_plan_id", "title"},
			},
			Handler: makeHandler(deps, deps.CreateSpecificPlan),
		},
		{
			Name:        "plan_ai.research_topic",
			Description: "Create a new research entry for a topic.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"topic":      {Type: "string", Description: "Research topic", Required: true},
					"summary":    {Type: "string", Description: "Brief summary"},
					"confidence": {Type: "integer", Description: "Confidence 0-100"},
				},
				Required: []string{"topic"},
			},
			Handler: makeHandler(deps, deps.ResearchTopic),
		},
		{
			Name:        "plan_ai.approve_plan",
			Description: "Approve a plan by ID.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"plan_id": {Type: "string", Description: "The plan ID to approve", Required: true},
				},
				Required: []string{"plan_id"},
			},
			Handler: makeHandler(deps, deps.ApprovePlan),
		},
		{
			Name:        "plan_ai.reject_plan",
			Description: "Reject a plan by ID.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"plan_id": {Type: "string", Description: "The plan ID to reject", Required: true},
				},
				Required: []string{"plan_id"},
			},
			Handler: makeHandler(deps, deps.RejectPlan),
		},
		{
			Name:        "plan_ai.analyze_impact",
			Description: "Analyze the impact of a change event.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"change_type": {
						Type:        "string",
						Description: "Type of change",
						Required:    true,
						Enum: []string{
							"vision_changed", "requirement_added", "requirement_removed",
							"constraint_changed", "decision_changed", "research_updated",
							"knowledge_updated", "plan_changed", "technology_changed",
							"implementation_feedback",
						},
					},
					"summary": {Type: "string", Description: "Summary of what changed", Required: true},
				},
				Required: []string{"change_type", "summary"},
			},
			Handler: makeHandler(deps, deps.AnalyzeImpact),
		},
		{
			Name:        "plan_ai.get_next_task",
			Description: "Get the next pending task from the current plan.",
			Schema: JSONSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
			Handler: makeHandler(deps, deps.GetNextTask),
		},
		{
			Name:        "plan_ai.mark_task_done",
			Description: "Mark a task as completed.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"task_id": {Type: "string", Description: "Task ID to mark done", Required: true},
				},
				Required: []string{"task_id"},
			},
			Handler: makeHandler(deps, deps.MarkTaskDone),
		},
		{
			Name:        "plan_ai.create_snapshot",
			Description: "Create a project state snapshot.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"reason": {Type: "string", Description: "Reason for the snapshot", Required: true},
				},
				Required: []string{"reason"},
			},
			Handler: makeHandler(deps, deps.CreateSnapshot),
		},
		{
			Name:        "plan_ai.list_plans",
			Description: "List all plans in the project.",
			Schema: JSONSchema{
				Type:       "object",
				Properties: map[string]Property{},
			},
			Handler: makeHandler(deps, deps.ListPlans),
		},
		{
			Name:        "plan_ai.list_tasks",
			Description: "List all tasks, optionally filtered by plan, phase, or status.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"plan_id": {Type: "string", Description: "Filter by plan ID"},
					"status":  {Type: "string", Description: "Filter by status (pending|in_progress|completed|blocked)"},
				},
			},
			Handler: makeHandler(deps, deps.ListTasks),
		},
	}

	agentTools := []ToolDefinition{
		{
			Name:        "plan_ai.agent_process",
			Description: "Process a user message through the agent system for intent detection, routing, and delegation.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"message":      {Type: "string", Description: "User message to process", Required: true},
				},
				Required: []string{"message"},
			},
			Handler: makeHandler(deps, deps.AgentProcess),
		},
		{
			Name:        "plan_ai.agent_message",
			Description: "Process a user message through the agent system.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"message":      {Type: "string", Description: "User message to process", Required: true},
				},
				Required: []string{"message"},
			},
			Handler: makeHandler(deps, deps.AgentProcess),
		},
		{
			Name:        "plan_ai.agent_runs",
			Description: "List recent agent runs for a project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"limit":        {Type: "integer", Description: "Maximum number of runs to return"},
				},
			},
			Handler: makeHandler(deps, deps.AgentRuns),
		},
		{
			Name:        "plan_ai.agent_status",
			Description: "Get recent agent activity and status.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"limit":        {Type: "integer", Description: "Maximum number of runs to return"},
				},
			},
			Handler: makeHandler(deps, deps.AgentRuns),
		},
		{
			Name:        "plan_ai.continuous_status",
			Description: "Get continuous planning status for a project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
				},
			},
			Handler: makeHandler(deps, deps.ContinuousStatus),
		},
		{
			Name:        "plan_ai.continuous_events",
			Description: "List recent continuous planning events.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"limit":        {Type: "integer", Description: "Maximum number of events"},
				},
			},
			Handler: makeHandler(deps, deps.ContinuousEvents),
		},
		{
			Name:        "plan_ai.continuous_proposals",
			Description: "List plan update proposals.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
				},
			},
			Handler: makeHandler(deps, deps.ContinuousProposals),
		},
		{
			Name:        "plan_ai.propose_plan_update",
			Description: "Create or list plan update proposals for continuous planning.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"reason":       {Type: "string", Description: "Reason for the proposed update"},
				},
			},
			Handler: makeHandler(deps, deps.ContinuousProposals),
		},
		{
			Name:        "plan_ai.approve_plan_update",
			Description: "Approve a pending plan update proposal.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"proposal_id":  {Type: "string", Description: "Plan update proposal ID", Required: true},
				},
				Required: []string{"proposal_id"},
			},
			Handler: makeHandler(deps, deps.ContinuousProposals),
		},
		{
			Name:        "plan_ai.reject_plan_update",
			Description: "Reject a pending plan update proposal.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"proposal_id":  {Type: "string", Description: "Plan update proposal ID", Required: true},
				},
				Required: []string{"proposal_id"},
			},
			Handler: makeHandler(deps, deps.ContinuousProposals),
		},
		{
			Name:        "plan_ai.continuous_context",
			Description: "Generate context at a specified level for a project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"level":        {Type: "string", Description: "Context level (L0_Executive, L1_Planning, L2_Specific_Plan, L3_Task, L4_Implementation)", Required: true},
				},
				Required: []string{"level"},
			},
			Handler: makeHandler(deps, deps.ContinuousContext),
		},
		{
			Name:        "plan_ai.get_context_level",
			Description: "Generate context at a specified L0-L4 level for a project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"level":        {Type: "string", Description: "Context level (L0_Executive, L1_Planning, L2_Specific_Plan, L3_Task, L4_Implementation)", Required: true},
				},
				Required: []string{"level"},
			},
			Handler: makeHandler(deps, deps.ContinuousContext),
		},
	}

	phase29Tools := []ToolDefinition{
		{
			Name:        "plan_ai.get_context",
			Description: "Get context at a specified level (L0-L4) for a project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"level":        {Type: "string", Description: "Context level (L0_executive|L1_planning|L2_implementation|L3_research)", Required: true},
					"task_id":      {Type: "string", Description: "Task ID (required for L2)"},
					"topic":        {Type: "string", Description: "Topic (used for L3)"},
				},
				Required: []string{"level"},
			},
			Handler: makeHandler(deps, deps.GetContext),
		},
		{
			Name:        "plan_ai.detect_changes",
			Description: "Detect and register changes in the project, returning impact analysis.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"change_type": {Type: "string", Description: "Type of change", Required: true,
						Enum: []string{
							"vision_changed", "requirement_added", "requirement_removed",
							"constraint_changed", "decision_changed", "research_updated",
							"knowledge_updated", "plan_changed", "technology_changed",
							"implementation_feedback",
						},
					},
					"summary":     {Type: "string", Description: "Summary of the change", Required: true},
					"description": {Type: "string", Description: "Detailed description"},
					"entity_type": {Type: "string", Description: "Affected entity type"},
					"entity_id":   {Type: "string", Description: "Affected entity ID"},
				},
				Required: []string{"change_type", "summary"},
			},
			Handler: makeHandler(deps, deps.DetectChanges),
		},
		{
			Name:        "plan_ai.update_plan",
			Description: "Update an existing plan's details (title, summary, status).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"plan_id":      {Type: "string", Description: "Plan ID to update", Required: true},
					"title":        {Type: "string", Description: "New title"},
					"summary":      {Type: "string", Description: "New summary"},
					"status":       {Type: "string", Description: "New status (draft|approved|archived)"},
				},
				Required: []string{"plan_id"},
			},
			Handler: makeHandler(deps, deps.UpdatePlan),
		},
		{
			Name:        "plan_ai.rollback_snapshot",
			Description: "Rollback to a previous project state snapshot (not yet implemented).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"snapshot_id":  {Type: "string", Description: "Snapshot ID to rollback to", Required: true},
				},
				Required: []string{"snapshot_id"},
			},
			Handler: makeHandler(deps, deps.RollbackSnapshot),
		},
		{
			Name:        "plan_ai.export_docs",
			Description: "Export project documentation (plans, decisions, research).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"format":       {Type: "string", Description: "Export format (markdown|json)", Required: true},
					"scope":        {Type: "string", Description: "What to export (plans|decisions|research|all)", Required: true},
				},
				Required: []string{"format", "scope"},
			},
			Handler: makeHandler(deps, deps.ExportDocs),
		},
	}

	phase51Tools := []ToolDefinition{
		{
			Name:        "plan_ai.create_product_intent",
			Description: "Create a Product Intent (Phase 51). Captures user expectations, success criteria, and what the product should achieve.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root":        {Type: "string", Description: "Project root path"},
					"description":         {Type: "string", Description: "Product intent description", Required: true},
					"expected_outcome":    {Type: "string", Description: "What outcome is expected"},
					"desired_experience":  {Type: "string", Description: "How the user expects the experience to feel"},
					"desired_result":      {Type: "string", Description: "Tangible desired result"},
					"user_expectations":   {Type: "string", Description: "Newline-separated list of what the user expects"},
					"non_expectations":    {Type: "string", Description: "Newline-separated list of what is NOT expected"},
					"success_definition":  {Type: "string", Description: "How success is measured"},
					"failure_definition":  {Type: "string", Description: "What would constitute failure"},
					"discovery_result_id": {Type: "string", Description: "Link to a Phase 52 discovery result"},
				},
				Required: []string{"description"},
			},
			Handler: makeHandler(deps, deps.CreateProductIntent),
		},
		{
			Name:        "plan_ai.list_product_intents",
			Description: "List all product intents for the project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
				},
			},
			Handler: makeHandler(deps, deps.ListProductIntents),
		},
		{
			Name:        "plan_ai.get_product_intent",
			Description: "Get a single product intent by ID.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"intent_id":    {Type: "string", Description: "Product intent ID", Required: true},
				},
				Required: []string{"intent_id"},
			},
			Handler: makeHandler(deps, deps.GetProductIntent),
		},
		{
			Name:        "plan_ai.submit_product_intent",
			Description: "Submit a product intent for approval (draft → pending_approval).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"intent_id":    {Type: "string", Description: "Product intent ID", Required: true},
				},
				Required: []string{"intent_id"},
			},
			Handler: makeHandler(deps, deps.SubmitProductIntent),
		},
		{
			Name:        "plan_ai.approve_product_intent",
			Description: "Approve a product intent (pending_approval → approved).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"intent_id":    {Type: "string", Description: "Product intent ID", Required: true},
				},
				Required: []string{"intent_id"},
			},
			Handler: makeHandler(deps, deps.ApproveProductIntent),
		},
		{
			Name:        "plan_ai.reject_product_intent",
			Description: "Reject a product intent (pending_approval → draft).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"intent_id":    {Type: "string", Description: "Product intent ID", Required: true},
				},
				Required: []string{"intent_id"},
			},
			Handler: makeHandler(deps, deps.RejectProductIntent),
		},
		{
			Name:        "plan_ai.discover_intent",
			Description: "Analyze raw user input to extract structured intent, objectives, restrictions, and questions (Phase 52 Discovery Engine).",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
					"content":      {Type: "string", Description: "Raw user input to analyze", Required: true},
				},
				Required: []string{"content"},
			},
			Handler: makeHandler(deps, deps.DiscoverIntent),
		},
		{
			Name:        "plan_ai.list_discovery_results",
			Description: "List discovery results for the project.",
			Schema: JSONSchema{
				Type: "object",
				Properties: map[string]Property{
					"project_root": {Type: "string", Description: "Project root path"},
				},
			},
			Handler: makeHandler(deps, deps.ListDiscoveryResults),
		},
	}

	tools = append(tools, agentTools...)
	tools = append(tools, phase29Tools...)
	tools = append(tools, phase51Tools...)

	return tools
}

// RegisterDefaultTools registers all standard Plan-AI MCP tools into a legacy Server.
func RegisterDefaultTools(s *Server, deps *ToolDependencies) error {
	for _, tool := range allToolDefs(deps) {
		if err := s.RegisterTool(tool); err != nil {
			return fmt.Errorf("register tool %s: %w", tool.Name, err)
		}
	}
	return nil
}

// RegisterSDKDefaultTools registers all tools directly into an mcpserver.MCPServer
// using SDK-native types. This is the production registration path.
func RegisterSDKDefaultTools(srv *mcpserver.MCPServer, deps *ToolDependencies, minimal bool) error {
	adapt := sdkHandlerAdapter()
	for _, td := range allToolDefs(deps) {
		if minimal && !MinimalToolNames[td.Name] {
			continue
		}
		schema, err := json.Marshal(td.Schema)
		if err != nil {
			return fmt.Errorf("marshal schema for %s: %w", td.Name, err)
		}
		td := td
		srv.AddTool(
			mcpsdk.NewToolWithRawSchema(td.Name, td.Description, schema),
			adapt(td.Handler),
		)
	}
	return nil
}

// ToolDependencies carries service implementations used by MCP tool handlers.
type ToolDependencies struct {
	InitProject        func(args map[string]any) (map[string]any, error)
	ProjectStatus      func(args map[string]any) (map[string]any, error)
	CreateMasterPlan   func(args map[string]any) (map[string]any, error)
	CreateSpecificPlan func(args map[string]any) (map[string]any, error)
	ResearchTopic      func(args map[string]any) (map[string]any, error)
	ApprovePlan        func(args map[string]any) (map[string]any, error)
	RejectPlan         func(args map[string]any) (map[string]any, error)
	AnalyzeImpact      func(args map[string]any) (map[string]any, error)
	GetNextTask        func(args map[string]any) (map[string]any, error)
	MarkTaskDone       func(args map[string]any) (map[string]any, error)
	CreateSnapshot     func(args map[string]any) (map[string]any, error)
	ListPlans          func(args map[string]any) (map[string]any, error)
	ListTasks          func(args map[string]any) (map[string]any, error)
	// Phase 21: Agent System
	AgentProcess func(args map[string]any) (map[string]any, error)
	AgentRuns    func(args map[string]any) (map[string]any, error)
	// Phase 22: Continuous Planning
	ContinuousStatus    func(args map[string]any) (map[string]any, error)
	ContinuousEvents    func(args map[string]any) (map[string]any, error)
	ContinuousProposals func(args map[string]any) (map[string]any, error)
	ContinuousContext   func(args map[string]any) (map[string]any, error)
	// Phase 29: Context & Change Management
	GetContext       func(args map[string]any) (map[string]any, error)
	DetectChanges    func(args map[string]any) (map[string]any, error)
	UpdatePlan       func(args map[string]any) (map[string]any, error)
	RollbackSnapshot func(args map[string]any) (map[string]any, error)
	ExportDocs       func(args map[string]any) (map[string]any, error)
	// Phase 51: Product Intent Engine
	CreateProductIntent  func(args map[string]any) (map[string]any, error)
	ListProductIntents   func(args map[string]any) (map[string]any, error)
	GetProductIntent     func(args map[string]any) (map[string]any, error)
	SubmitProductIntent  func(args map[string]any) (map[string]any, error)
	ApproveProductIntent func(args map[string]any) (map[string]any, error)
	RejectProductIntent  func(args map[string]any) (map[string]any, error)
	// Phase 52: Discovery Engine
	DiscoverIntent       func(args map[string]any) (map[string]any, error)
	ListDiscoveryResults func(args map[string]any) (map[string]any, error)
}

func makeHandler(deps *ToolDependencies, fn func(args map[string]any) (map[string]any, error)) ToolHandlerFunc {
	return func(args map[string]any) (*ToolResult, error) {
		data, err := fn(args)
		if err != nil {
			return &ToolResult{Success: false, Error: err.Error()}, nil
		}
		return &ToolResult{Success: true, Data: data}, nil
	}
}
