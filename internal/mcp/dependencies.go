package mcp

// DefaultToolDependencies wires the production handlers for all built-in MCP tools.
func DefaultToolDependencies() ToolDependencies {
	return ToolDependencies{
		InitProject:          HandleInitProject,
		ProjectStatus:        HandleProjectStatus,
		CreateMasterPlan:     HandleCreateMasterPlan,
		CreateSpecificPlan:   HandleCreateSpecificPlan,
		ResearchTopic:        HandleResearchTopic,
		ApprovePlan:          HandleApprovePlan,
		RejectPlan:           HandleRejectPlan,
		AnalyzeImpact:        HandleAnalyzeImpact,
		GetNextTask:          HandleGetNextTask,
		MarkTaskDone:         HandleMarkTaskDone,
		CreateSnapshot:       HandleCreateSnapshot,
		ListPlans:            HandleListPlans,
		ListTasks:            HandleListTasks,
		AgentProcess:         HandleAgentProcess,
		AgentRuns:            HandleAgentRuns,
		ContinuousStatus:     HandleContinuousStatus,
		ContinuousEvents:     HandleContinuousEvents,
		ContinuousProposals:  HandleContinuousProposals,
		ContinuousContext:    HandleContinuousContext,
		GetContext:           HandleGetContext,
		DetectChanges:        HandleDetectChanges,
		UpdatePlan:           HandleUpdatePlan,
		RollbackSnapshot:     HandleRollbackSnapshot,
		ExportDocs:           HandleExportDocs,
		CreateProductIntent:  HandleCreateProductIntent,
		ListProductIntents:   HandleListProductIntents,
		GetProductIntent:     HandleGetProductIntent,
		SubmitProductIntent:  HandleSubmitProductIntent,
		ApproveProductIntent: HandleApproveProductIntent,
		RejectProductIntent:  HandleRejectProductIntent,
		DiscoverIntent:       HandleDiscoverIntent,
		ListDiscoveryResults: HandleListDiscoveryResults,
	}
}
