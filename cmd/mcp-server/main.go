package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/plan-ai/plan-ai/internal/mcp"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return serveStdio()
	}

	ctx := mcp.ToolContext{}
	s := mcp.NewServer(ctx)

	deps := mcp.ToolDependencies{
		InitProject:          mcp.HandleInitProject,
		ProjectStatus:        mcp.HandleProjectStatus,
		CreateMasterPlan:     mcp.HandleCreateMasterPlan,
		CreateSpecificPlan:   mcp.HandleCreateSpecificPlan,
		ResearchTopic:        mcp.HandleResearchTopic,
		ApprovePlan:          mcp.HandleApprovePlan,
		RejectPlan:           mcp.HandleRejectPlan,
		AnalyzeImpact:        mcp.HandleAnalyzeImpact,
		GetNextTask:          mcp.HandleGetNextTask,
		MarkTaskDone:         mcp.HandleMarkTaskDone,
		CreateSnapshot:       mcp.HandleCreateSnapshot,
		ListPlans:            mcp.HandleListPlans,
		ListTasks:            mcp.HandleListTasks,
		AgentProcess:         mcp.HandleAgentProcess,
		AgentRuns:            mcp.HandleAgentRuns,
		ContinuousStatus:     mcp.HandleContinuousStatus,
		ContinuousEvents:     mcp.HandleContinuousEvents,
		ContinuousProposals:  mcp.HandleContinuousProposals,
		ContinuousContext:    mcp.HandleContinuousContext,
		GetContext:           mcp.HandleGetContext,
		DetectChanges:        mcp.HandleDetectChanges,
		UpdatePlan:           mcp.HandleUpdatePlan,
		RollbackSnapshot:     mcp.HandleRollbackSnapshot,
		ExportDocs:           mcp.HandleExportDocs,
		CreateProductIntent:  mcp.HandleCreateProductIntent,
		ListProductIntents:   mcp.HandleListProductIntents,
		GetProductIntent:     mcp.HandleGetProductIntent,
		SubmitProductIntent:  mcp.HandleSubmitProductIntent,
		ApproveProductIntent: mcp.HandleApproveProductIntent,
		RejectProductIntent:  mcp.HandleRejectProductIntent,
		DiscoverIntent:       mcp.HandleDiscoverIntent,
		ListDiscoveryResults: mcp.HandleListDiscoveryResults,
	}

	if err := mcp.RegisterDefaultTools(s, &deps); err != nil {
		return fmt.Errorf("register tools: %w", err)
	}

	switch os.Args[1] {
	case "stdio", "serve":
		return serveStdio()

	case "list-tools":
		tools := s.ListTools()
		data, _ := json.MarshalIndent(tools, "", "  ")
		fmt.Println(string(data))
		return nil

	case "call-tool":
		if len(os.Args) < 3 {
			return fmt.Errorf("usage: mcp-server call-tool <tool-name> [json-args]")
		}
		toolName := os.Args[2]
		var args map[string]any
		if len(os.Args) > 3 {
			if err := json.Unmarshal([]byte(os.Args[3]), &args); err != nil {
				return fmt.Errorf("invalid json args: %w", err)
			}
		}
		result := s.ExecuteTool(toolName, args)
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil

	default:
		return fmt.Errorf("unknown command: %s (use list-tools or call-tool)", os.Args[1])
	}
}

func serveStdio() error {
	ctx := mcp.ToolContext{}
	s := mcp.NewServer(ctx)
	deps := defaultToolDependencies()
	if err := mcp.RegisterDefaultTools(s, &deps); err != nil {
		return fmt.Errorf("register tools: %w", err)
	}
	return mcp.ServeStdio(os.Stdin, os.Stdout, s)
}

func defaultToolDependencies() mcp.ToolDependencies {
	return mcp.ToolDependencies{
		InitProject:          mcp.HandleInitProject,
		ProjectStatus:        mcp.HandleProjectStatus,
		CreateMasterPlan:     mcp.HandleCreateMasterPlan,
		CreateSpecificPlan:   mcp.HandleCreateSpecificPlan,
		ResearchTopic:        mcp.HandleResearchTopic,
		ApprovePlan:          mcp.HandleApprovePlan,
		RejectPlan:           mcp.HandleRejectPlan,
		AnalyzeImpact:        mcp.HandleAnalyzeImpact,
		GetNextTask:          mcp.HandleGetNextTask,
		MarkTaskDone:         mcp.HandleMarkTaskDone,
		CreateSnapshot:       mcp.HandleCreateSnapshot,
		ListPlans:            mcp.HandleListPlans,
		ListTasks:            mcp.HandleListTasks,
		AgentProcess:         mcp.HandleAgentProcess,
		AgentRuns:            mcp.HandleAgentRuns,
		ContinuousStatus:     mcp.HandleContinuousStatus,
		ContinuousEvents:     mcp.HandleContinuousEvents,
		ContinuousProposals:  mcp.HandleContinuousProposals,
		ContinuousContext:    mcp.HandleContinuousContext,
		GetContext:           mcp.HandleGetContext,
		DetectChanges:        mcp.HandleDetectChanges,
		UpdatePlan:           mcp.HandleUpdatePlan,
		RollbackSnapshot:     mcp.HandleRollbackSnapshot,
		ExportDocs:           mcp.HandleExportDocs,
		CreateProductIntent:  mcp.HandleCreateProductIntent,
		ListProductIntents:   mcp.HandleListProductIntents,
		GetProductIntent:     mcp.HandleGetProductIntent,
		SubmitProductIntent:  mcp.HandleSubmitProductIntent,
		ApproveProductIntent: mcp.HandleApproveProductIntent,
		RejectProductIntent:  mcp.HandleRejectProductIntent,
		DiscoverIntent:       mcp.HandleDiscoverIntent,
		ListDiscoveryResults: mcp.HandleListDiscoveryResults,
	}
}
