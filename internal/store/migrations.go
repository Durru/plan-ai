package store

const (
	GlobalStoreLayerV2MigrationID            = "0007_store_layer_v2_global"
	ProjectStoreLayerV2MigrationID           = "0007_store_layer_v2"
	ProjectIngestionVisionContextMigrationID = "0008_ingestion_vision_context"
	ProjectModelStrategyMigrationID          = "0012_model_strategy"
	ProjectOrchestratorMigrationID           = "0013_orchestrator"
	ProjectContextEngineMigrationID          = "0014_context_engine"
	ProjectChangeEngineMigrationID           = "0015_change_engine"
	ProjectMCPServerMigrationID              = "0016_mcp_server"
	ProjectOpenCodeIntegrationMigrationID    = "0017_opencode_integration"
	ProjectCompatibilityMigrationID          = "0018_compatibility"
	ProjectAgentSystemMigrationID            = "0019_agent_system"
	ProjectContinuousPlanningMigrationID     = "0020_continuous_planning"
)

// StoreLayerV2Tables documents the definitive Phase 8 project tables.
// Existing legacy tables are intentionally not listed here; they remain
// additive compatibility tables managed by earlier migrations.
var StoreLayerV2Tables = []string{
	"projects", "raw_inputs", "ingested_sources", "visions", "requirements",
	"constraints", "decisions", "decision_history", "research_entries",
	"research_sources", "research_findings", "knowledge_objects",
	"knowledge_relations", "knowledge_references", "master_plans",
	"specific_plans", "implementation_documents", "phases", "tasks",
	"task_steps", "validations", "snapshots", "change_requests",
	"impact_reports", "context_views", "context_chunks", "agent_runs",
	"subagent_outputs", "project_scans", "file_snapshots", "file_change_events",
	"approved_requirements", "approved_constraints", "approved_decisions",
	"approved_preferences", "approved_references", "approved_goals",
	"model_profiles", "prompt_contracts", "output_schemas",
	"jobs", "job_runs", "capabilities",
	"context_views_v2",
}

// Phase15Tables lists the tables added by migration 0012 (model strategy).
var Phase15Tables = []string{"model_profiles", "prompt_contracts", "output_schemas"}

// Phase16Tables lists the tables added by migration 0013 (orchestrator).
var Phase16Tables = []string{"jobs", "job_runs", "capabilities"}

// Phase17Tables lists the tables added by migration 0014 (context engine).
var Phase17Tables = []string{"context_views_v2"}

// Phase18Tables lists the tables added by migration 0015 (change engine).
var Phase18Tables = []string{"change_events", "change_reports", "snapshots_v2", "entity_states"}

// Phase19Tables lists the tables added by migration 0016 (MCP server).
var Phase19Tables = []string{"mcp_tools", "mcp_runs"}

// Phase20Tables lists the tables added by migration 0017 (opencode integration).
var Phase20Tables = []string{"opencode_detections", "opencode_integration_state", "opencode_doctor_checks"}

// Phase1820CompatibilityTables lists the compatibility tables/views added by
// migration 0018. These ensure all required table names exist for acceptance.
var Phase1820CompatibilityTables = []string{"tool_runs", "tool_audit", "provider_registry", "skill_registry"}

// Phase21Tables lists the tables added by migration 0019 (agent system).
var Phase21Tables = []string{"agent_runs_v2", "agent_messages", "agent_delegated_jobs"}

// Phase22Tables lists the tables added by migration 0020 (continuous planning).
var Phase22Tables = []string{"continuous_events", "plan_update_proposals", "context_deliveries"}
