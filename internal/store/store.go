package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/config"
	_ "modernc.org/sqlite"
)

type GlobalLayout struct {
	Dir        string
	DBPath     string
	ConfigPath string
	CacheDir   string
	SkillsDir  string
	LogsDir    string
	DataDir    string
	BackupsDir string
}

type ProjectLayout struct {
	Dir          string
	DBPath       string
	ConfigPath   string
	CacheDir     string
	SnapshotsDir string
	ExportsDir   string
	DocsDir      string
	LocksDir     string
	BackupsDir   string
}

type Migration struct {
	ID   string
	Name string
	SQL  string
}

func EnsureGlobalLayout(homeDir string) (GlobalLayout, error) {
	layout := GlobalLayout{
		Dir:        config.GlobalDir(homeDir),
		DBPath:     config.GlobalDBPath(homeDir),
		ConfigPath: config.GlobalConfigPath(homeDir),
		CacheDir:   config.GlobalCacheDir(homeDir),
		SkillsDir:  config.GlobalSkillsDir(homeDir),
		LogsDir:    config.GlobalLogsDir(homeDir),
		DataDir:    config.GlobalDataDir(homeDir),
		BackupsDir: config.GlobalBackupsDir(homeDir),
	}
	for _, dir := range []string{layout.Dir, layout.CacheDir, layout.SkillsDir, layout.LogsDir, layout.DataDir, layout.BackupsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return layout, err
		}
	}
	return layout, nil
}

func EnsureProjectLayout(projectRoot string) (ProjectLayout, error) {
	layout := ProjectLayout{
		Dir:          config.ProjectDir(projectRoot),
		DBPath:       config.ProjectDBPath(projectRoot),
		ConfigPath:   config.ProjectConfigPath(projectRoot),
		CacheDir:     config.ProjectCacheDir(projectRoot),
		SnapshotsDir: config.ProjectSnapshotsDir(projectRoot),
		ExportsDir:   config.ProjectExportsDir(projectRoot),
		DocsDir:      config.ProjectDocsDir(projectRoot),
		LocksDir:     config.ProjectLocksDir(projectRoot),
		BackupsDir:   config.ProjectBackupsDir(projectRoot),
	}
	for _, dir := range []string{layout.Dir, layout.CacheDir, layout.SnapshotsDir, layout.ExportsDir, layout.DocsDir, layout.LocksDir, layout.BackupsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return layout, err
		}
	}
	return layout, nil
}

// ExternalProjectLayout describes the on-disk layout of an external project
// (one that lives under the global Plan-AI home instead of inside the
// project's working tree).
type ExternalProjectLayout struct {
	Dir          string
	DBPath       string
	ConfigPath   string
	CacheDir     string
	SnapshotsDir string
	ExportsDir   string
	DocsDir      string
	LocksDir     string
	BackupsDir   string
}

// EnsureExternalProjectLayout creates the on-disk directories required for
// an external project under <home>/.plan-ai/projects/<slug>/. It is safe to
// call repeatedly: existing directories are left untouched.
func EnsureExternalProjectLayout(homeDir, projectSlug string) (ExternalProjectLayout, error) {
	layout := ExternalProjectLayout{
		Dir:          config.ExternalProjectDir(homeDir, projectSlug),
		DBPath:       config.ExternalProjectDBPath(homeDir, projectSlug),
		ConfigPath:   config.ExternalProjectConfigPath(homeDir, projectSlug),
		CacheDir:     config.ExternalProjectCacheDir(homeDir, projectSlug),
		SnapshotsDir: config.ExternalProjectSnapshotsDir(homeDir, projectSlug),
		ExportsDir:   config.ExternalProjectExportsDir(homeDir, projectSlug),
		DocsDir:      config.ExternalProjectDocsDir(homeDir, projectSlug),
		LocksDir:     config.ExternalProjectLocksDir(homeDir, projectSlug),
		BackupsDir:   config.ExternalProjectBackupsDir(homeDir, projectSlug),
	}
	for _, dir := range []string{layout.Dir, layout.CacheDir, layout.SnapshotsDir, layout.ExportsDir, layout.DocsDir, layout.LocksDir, layout.BackupsDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return layout, err
		}
	}
	return layout, nil
}

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000`); err != nil {
		db.Close()
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func RunGlobalMigrations(db *sql.DB) error {
	return runMigrations(db, []Migration{{
		ID:   "0001_global_base",
		Name: "Create global persistence tables",
		SQL: `
CREATE TABLE IF NOT EXISTS global_metadata (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS global_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS known_projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  path TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);`,
	}, {
		ID:   "0007_store_layer_v2_global",
		Name: "Create definitive global store tables",
		SQL: `
CREATE TABLE IF NOT EXISTS global_config (key TEXT PRIMARY KEY, value TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_tools (id TEXT PRIMARY KEY, name TEXT NOT NULL, version TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_integrations (id TEXT PRIMARY KEY, name TEXT NOT NULL, kind TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'inactive', config TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_skills (id TEXT PRIMARY KEY, name TEXT NOT NULL, source TEXT NOT NULL DEFAULT '', checksum TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_skill_cache (id TEXT PRIMARY KEY, skill_id TEXT NOT NULL, cache_key TEXT NOT NULL, value TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(skill_id, cache_key));
CREATE TABLE IF NOT EXISTS global_knowledge (id TEXT PRIMARY KEY, topic TEXT NOT NULL, category TEXT NOT NULL DEFAULT 'general', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_research (id TEXT PRIMARY KEY, topic TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_templates (id TEXT PRIMARY KEY, name TEXT NOT NULL, template_type TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_model_profiles (id TEXT PRIMARY KEY, name TEXT NOT NULL, provider TEXT NOT NULL DEFAULT '', model TEXT NOT NULL DEFAULT '', config TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS global_logs (id TEXT PRIMARY KEY, level TEXT NOT NULL, message TEXT NOT NULL, metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_global_tools_name ON global_tools(name);
CREATE INDEX IF NOT EXISTS idx_global_skills_name ON global_skills(name);
CREATE INDEX IF NOT EXISTS idx_global_knowledge_topic ON global_knowledge(topic);
CREATE INDEX IF NOT EXISTS idx_global_research_topic ON global_research(topic);
`,
	}, {
		ID:   "0008_project_registry_v2",
		Name: "Extend known_projects with mode, slug, last_seen_at",
		SQL: `
ALTER TABLE known_projects ADD COLUMN mode TEXT NOT NULL DEFAULT 'external';
ALTER TABLE known_projects ADD COLUMN slug TEXT NOT NULL DEFAULT '';
ALTER TABLE known_projects ADD COLUMN last_seen_at TEXT NOT NULL DEFAULT '';
UPDATE known_projects SET slug = name WHERE slug = '';
CREATE INDEX IF NOT EXISTS idx_known_projects_slug ON known_projects(slug);
CREATE INDEX IF NOT EXISTS idx_known_projects_mode ON known_projects(mode);
`,
	}})
}

func RunProjectMigrations(db *sql.DB) error {
	return runMigrations(db, []Migration{
		{
			ID:   "0001_project_base",
			Name: "Create project persistence tables",
			SQL: `
CREATE TABLE IF NOT EXISTS project_metadata (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_settings (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_state (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  root_path TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);`,
		},
		{
			ID:   "0002_project_domain",
			Name: "Create project domain tables",
			SQL: `
CREATE TABLE IF NOT EXISTS plans (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  status TEXT NOT NULL,
  version INTEGER NOT NULL,
  parent_plan_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS phases (
  id TEXT PRIMARY KEY,
  plan_id TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  status TEXT NOT NULL,
  position INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  phase_id TEXT NOT NULL,
  plan_id TEXT NOT NULL,
  title TEXT NOT NULL,
  summary TEXT NOT NULL,
  status TEXT NOT NULL,
  position INTEGER NOT NULL,
  context_size TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS decisions (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  context TEXT NOT NULL,
  decision TEXT NOT NULL,
  status TEXT NOT NULL,
  impact TEXT NOT NULL,
  supersedes_id TEXT NOT NULL DEFAULT '',
  superseded_by_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS research_entries (
  id TEXT PRIMARY KEY,
  topic TEXT NOT NULL,
  source TEXT NOT NULL,
  summary TEXT NOT NULL,
  conclusion TEXT NOT NULL,
  confidence REAL NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS knowledge_objects (
  id TEXT PRIMARY KEY,
  topic TEXT NOT NULL,
  summary TEXT NOT NULL,
  content TEXT NOT NULL,
  confidence REAL NOT NULL,
  reuse_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS validations (
  id TEXT PRIMARY KEY,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  status TEXT NOT NULL,
  summary TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS snapshots (
  id TEXT PRIMARY KEY,
  reason TEXT NOT NULL,
  summary TEXT NOT NULL,
  created_at TEXT NOT NULL
);`,
		},
		{
			ID:   "0003_project_scan",
			Name: "Create project scan tables",
			SQL: `
CREATE TABLE IF NOT EXISTS project_scans (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL,
  git_detected INTEGER NOT NULL,
  git_branch TEXT,
  fingerprint TEXT NOT NULL,
  summary TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_scan_languages (
  id TEXT PRIMARY KEY,
  scan_id TEXT NOT NULL,
  language TEXT NOT NULL,
  files_count INTEGER NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_scan_frameworks (
  id TEXT PRIMARY KEY,
  scan_id TEXT NOT NULL,
  framework TEXT NOT NULL,
  evidence TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_scan_package_managers (
  id TEXT PRIMARY KEY,
  scan_id TEXT NOT NULL,
  manager TEXT NOT NULL,
  evidence TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_scan_dependencies (
  id TEXT PRIMARY KEY,
  scan_id TEXT NOT NULL,
  name TEXT NOT NULL,
  version TEXT,
  source TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS project_scan_files (
  id TEXT PRIMARY KEY,
  scan_id TEXT NOT NULL,
  path TEXT NOT NULL,
  kind TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  created_at TEXT NOT NULL
);`,
		},
		{
			ID:   "0004_knowledge_base",
			Name: "Create knowledge base tables and extend knowledge_objects",
			SQL: `
ALTER TABLE knowledge_objects ADD COLUMN category TEXT NOT NULL DEFAULT 'general';
ALTER TABLE knowledge_objects ADD COLUMN status TEXT NOT NULL DEFAULT 'draft';
ALTER TABLE knowledge_objects ADD COLUMN source_type TEXT NOT NULL DEFAULT 'manual';
CREATE INDEX IF NOT EXISTS idx_knowledge_objects_category ON knowledge_objects(category);
CREATE INDEX IF NOT EXISTS idx_knowledge_objects_status ON knowledge_objects(status);
CREATE TABLE IF NOT EXISTS knowledge_tags (
  id TEXT PRIMARY KEY,
  knowledge_id TEXT NOT NULL,
  tag TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(knowledge_id, tag)
);
CREATE INDEX IF NOT EXISTS idx_knowledge_tags_knowledge_id ON knowledge_tags(knowledge_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_tags_tag ON knowledge_tags(tag);
CREATE TABLE IF NOT EXISTS knowledge_relations (
  id TEXT PRIMARY KEY,
  source_id TEXT NOT NULL,
  target_id TEXT NOT NULL,
  relation_type TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(source_id, target_id, relation_type)
);
CREATE INDEX IF NOT EXISTS idx_knowledge_relations_source ON knowledge_relations(source_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_relations_target ON knowledge_relations(target_id);
CREATE TABLE IF NOT EXISTS knowledge_references (
  id TEXT PRIMARY KEY,
  knowledge_id TEXT NOT NULL,
  reference_type TEXT NOT NULL,
  reference_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(knowledge_id, reference_type, reference_id)
);
CREATE INDEX IF NOT EXISTS idx_knowledge_references_knowledge_id ON knowledge_references(knowledge_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_references_type ON knowledge_references(reference_type);`,
		},
		{
			ID:   "0005_research_engine",
			Name: "Extend research_entries and create research sub-tables",
			SQL: `
ALTER TABLE research_entries ADD COLUMN category TEXT NOT NULL DEFAULT 'general';
ALTER TABLE research_entries ADD COLUMN status TEXT NOT NULL DEFAULT 'draft';
CREATE INDEX IF NOT EXISTS idx_research_entries_category ON research_entries(category);
CREATE INDEX IF NOT EXISTS idx_research_entries_status ON research_entries(status);
CREATE TABLE IF NOT EXISTS research_findings (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL,
  importance INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  FOREIGN KEY (research_id) REFERENCES research_entries(id)
);
CREATE INDEX IF NOT EXISTS idx_research_findings_research_id ON research_findings(research_id);
CREATE TABLE IF NOT EXISTS research_sources (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  title TEXT NOT NULL,
  url TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL DEFAULT 'manual',
  created_at TEXT NOT NULL,
  FOREIGN KEY (research_id) REFERENCES research_entries(id)
);
CREATE INDEX IF NOT EXISTS idx_research_sources_research_id ON research_sources(research_id);
CREATE TABLE IF NOT EXISTS research_conclusions (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  content TEXT NOT NULL,
  confidence INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  FOREIGN KEY (research_id) REFERENCES research_entries(id)
);
CREATE INDEX IF NOT EXISTS idx_research_conclusions_research_id ON research_conclusions(research_id);
CREATE TABLE IF NOT EXISTS research_tags (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  tag TEXT NOT NULL,
  UNIQUE(research_id, tag)
);
CREATE INDEX IF NOT EXISTS idx_research_tags_research_id ON research_tags(research_id);
CREATE INDEX IF NOT EXISTS idx_research_tags_tag ON research_tags(tag);
CREATE TABLE IF NOT EXISTS research_knowledge_links (
  id TEXT PRIMARY KEY,
  research_id TEXT NOT NULL,
  knowledge_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(research_id, knowledge_id)
);
CREATE INDEX IF NOT EXISTS idx_research_knowledge_links_research_id ON research_knowledge_links(research_id);
CREATE INDEX IF NOT EXISTS idx_research_knowledge_links_knowledge_id ON research_knowledge_links(knowledge_id);`,
		},
		{
			ID:   "0007_store_layer_v2",
			Name: "Create definitive project store tables",
			SQL:  projectStoreLayerV2SQL,
		},
		{
			ID:   "0008_ingestion_vision_context",
			Name: "Extend ingestion, vision, and approved context tables",
			SQL:  projectIngestionVisionContextSQL,
		},
		{
			ID:   "0009_research_knowledge",
			Name: "Create research and knowledge engine tables",
			SQL:  projectResearchKnowledgeSQL,
		},
		{
			ID:   "0010_planning_framework",
			Name: "Create planning framework tables",
			SQL:  projectPlanningFrameworkSQL,
		},
		{
			ID:   "0011_workflows",
			Name: "Create workflow run tables",
			SQL:  projectWorkflowsSQL,
		},
		{
			ID:   "0012_model_strategy",
			Name: "Create model strategy tables",
			SQL:  projectModelStrategySQL,
		},
		{
			ID:   "0013_orchestrator",
			Name: "Create orchestrator tables",
			SQL:  projectOrchestratorSQL,
		},
		{
			ID:   "0014_context_engine",
			Name: "Create context engine tables",
			SQL:  projectContextEngineSQL,
		},
		{
			ID:   "0015_change_engine",
			Name: "Create change engine tables",
			SQL:  projectChangeEngineSQL,
		},
		{
			ID:   "0016_mcp_server",
			Name: "Create MCP server tables",
			SQL:  projectMCPServerSQL,
		},
		{
			ID:   "0017_opencode_integration",
			Name: "Create opencode integration tables",
			SQL:  projectOpenCodeIntegrationSQL,
		},
		{
			ID:   "0018_compatibility",
			Name: "Create compatibility tables and views",
			SQL:  projectCompatibilitySQL,
		},
		{
			ID:   "0019_agent_system",
			Name: "Create agent system tables",
			SQL:  projectAgentSystemSQL,
		},
		{
			ID:   "0020_continuous_planning",
			Name: "Create continuous planning tables",
			SQL:  projectContinuousPlanningSQL,
		},
		{
			ID:   "0021_agent_continuous_compatibility",
			Name: "Create agent and continuous planning compatibility views",
			SQL:  projectAgentContinuousCompatibilitySQL,
		},
		{
			ID:   "0022_vision_discovery",
			Name: "Create vision discovery engine tables",
			SQL:  projectVisionDiscoverySQL,
		},
		{
			ID:   "0023_master_plan_v2",
			Name: "Create master plan v2 tables",
			SQL:  projectMasterPlanV2SQL,
		},
		{
			ID:   "0024_specific_plan_v2",
			Name: "Create specific plan v2 tables",
			SQL:  projectSpecificPlanV2SQL,
		},
		{
			ID:   "0025_context_delivery_engine",
			Name: "Create context delivery engine tables",
			SQL:  projectContextDeliveryEngineSQL,
		},
		{
			ID:   "0026_v2_user_intent_engine",
			Name: "Create V2 user intent engine tables",
			SQL:  projectUserIntentEngineSQL,
		},
		{
			ID:   "0027_v2_vision_documents",
			Name: "Create V2 vision document tables",
			SQL:  projectVisionDocumentsSQL,
		},
		{
			ID:   "0028_v2_approval_workflow",
			Name: "Create V2 approval workflow tables",
			SQL:  projectApprovalWorkflowSQL,
		},
		{
			ID:   "0029_v2_requirement_discovery",
			Name: "Create V2 requirement discovery tables",
			SQL:  projectRequirementDiscoverySQL,
		},
		{
			ID:   "0030_v2_context_packages",
			Name: "Create V2 smart context package tables",
			SQL:  projectContextPackagesV2SQL,
		},
		{
			ID:   "0031_v2_research_orchestration",
			Name: "Create V2 research orchestration tables",
			SQL:  projectResearchOrchestrationV2SQL,
		},
		{
			ID:   "0032_v2_reference_engine",
			Name: "Create V2 reference engine tables",
			SQL:  projectReferenceEngineV2SQL,
		},
		{
			ID:   "0033_v2_plan_evolution_v3",
			Name: "Create V2 plan evolution v3 tables",
			SQL:  projectPlanEvolutionV3SQL,
		},
		{
			ID:   "0034_v2_implementation_packages",
			Name: "Create V2 implementation package tables",
			SQL:  projectImplementationPackagesV2SQL,
		},
		{
			ID:   "0035_v2_change_impact",
			Name: "Create V2 change impact report tables",
			SQL:  projectChangeImpactV2SQL,
		},
		{
			ID:   "0036_v2_continuous_regeneration",
			Name: "Create V2 continuous regeneration tables",
			SQL:  projectContinuousRegenerationV2SQL,
		},
		{
			ID:   "0037_v2_subagent_orchestrator",
			Name: "Create V2 subagent orchestrator tables",
			SQL:  projectSubagentOrchestratorV2SQL,
		},
		{
			ID:   "0038_v2_opencode_workflows",
			Name: "Create V2 OpenCode workflow tables",
			SQL:  projectOpenCodeWorkflowsV2SQL,
		},
		{
			ID:   "0039_v2_project_memory",
			Name: "Create V2 project memory tables",
			SQL:  projectMemoryV2SQL,
		},
		{
			ID:   "0040_v3_product_intent",
			Name: "Create V3 product intent and discovery tables (Phase 51+52)",
			SQL:  projectV3ProductIntentSQL,
		},
		{
			ID:   "0041_v3_discovery_progressive",
			Name: "Create V3 progressive discovery tables (Phase 53)",
			SQL:  projectV3DiscoveryProgressiveSQL,
		},
		{
			ID:   "0042_fts5_search",
			Name: "Create FTS5 indexes for full-text search",
			SQL:  projectFTS5SQL,
		},
		{
			ID:   "0043_impact_graph",
			Name: "Create impact graph edges table",
			SQL:  projectImpactGraphSQL,
		},
		{
			ID:   "0044_capabilities_v2",
			Name: "Create capabilities_v2 registry table",
			SQL:  projectCapabilitiesV2SQL,
		},
		{
			ID:   "0045_workflow_runs_steps",
			Name: "Add steps column to workflow_runs for step-by-step execution",
			SQL:  projectWorkflowRunsStepsSQL,
		},
	})
}

const projectStoreLayerV2SQL = `
CREATE TABLE IF NOT EXISTS projects (id TEXT PRIMARY KEY, name TEXT NOT NULL, root_path TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS raw_inputs (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_type TEXT NOT NULL DEFAULT 'manual', content TEXT NOT NULL, metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS ingested_sources (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, raw_input_id TEXT, uri TEXT NOT NULL DEFAULT '', source_type TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS visions (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, expected_outcome TEXT NOT NULL DEFAULT '', approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS requirements (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, type TEXT NOT NULL, statement TEXT NOT NULL, approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS constraints (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, type TEXT NOT NULL, description TEXT NOT NULL, approved INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS decision_history (id TEXT PRIMARY KEY, decision_id TEXT NOT NULL, from_status TEXT NOT NULL DEFAULT '', to_status TEXT NOT NULL, note TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS implementation_documents (id TEXT PRIMARY KEY, project_id TEXT NOT NULL DEFAULT '', specific_plan_id TEXT NOT NULL, title TEXT NOT NULL, content TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS specific_plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, master_plan_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS master_plans (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL DEFAULT 1, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS task_steps (id TEXT PRIMARY KEY, task_id TEXT NOT NULL, title TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'pending', position INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS change_requests (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, reason TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', status TEXT NOT NULL, requester TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS impact_reports (id TEXT PRIMARY KEY, change_request_id TEXT NOT NULL, affected_plans TEXT NOT NULL DEFAULT '[]', affected_phases TEXT NOT NULL DEFAULT '[]', affected_tasks TEXT NOT NULL DEFAULT '[]', affected_decisions TEXT NOT NULL DEFAULT '[]', affected_entities TEXT NOT NULL DEFAULT '[]', summary TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS context_views (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, name TEXT NOT NULL, content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS context_chunks (id TEXT PRIMARY KEY, context_view_id TEXT NOT NULL, chunk_index INTEGER NOT NULL DEFAULT 0, content TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS agent_runs (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, agent_name TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'draft', metadata TEXT NOT NULL DEFAULT '{}', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS subagent_outputs (id TEXT PRIMARY KEY, agent_run_id TEXT NOT NULL, subagent_name TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS file_snapshots (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, path TEXT NOT NULL, hash TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS file_change_events (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, path TEXT NOT NULL, event_type TEXT NOT NULL, created_at TEXT NOT NULL);
-- project_id, supersedes_id, superseded_by_id are now in CREATE TABLE decisions; keep only legacy ALTER columns
ALTER TABLE decisions ADD COLUMN rationale TEXT NOT NULL DEFAULT '';
ALTER TABLE decisions ADD COLUMN alternatives TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN objective TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN date TEXT NOT NULL DEFAULT '';
ALTER TABLE research_entries ADD COLUMN reuse_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE research_entries ADD COLUMN reused_at TEXT NOT NULL DEFAULT '';
ALTER TABLE knowledge_objects ADD COLUMN type TEXT NOT NULL DEFAULT 'reference';
ALTER TABLE phases ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE validations ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE validations ADD COLUMN type TEXT NOT NULL DEFAULT 'manual';
ALTER TABLE validations ADD COLUMN details TEXT NOT NULL DEFAULT '';
ALTER TABLE snapshots ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE snapshots ADD COLUMN updated_at TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_raw_inputs_project ON raw_inputs(project_id);
CREATE INDEX IF NOT EXISTS idx_visions_project ON visions(project_id);
CREATE INDEX IF NOT EXISTS idx_requirements_project ON requirements(project_id);
CREATE INDEX IF NOT EXISTS idx_constraints_project ON constraints(project_id);
CREATE INDEX IF NOT EXISTS idx_decisions_project ON decisions(project_id);
CREATE INDEX IF NOT EXISTS idx_research_entries_project ON research_entries(project_id);
CREATE INDEX IF NOT EXISTS idx_master_plans_project ON master_plans(project_id);
CREATE INDEX IF NOT EXISTS idx_specific_plans_master ON specific_plans(master_plan_id);
CREATE INDEX IF NOT EXISTS idx_phases_plan ON phases(plan_id);
CREATE INDEX IF NOT EXISTS idx_tasks_phase ON tasks(phase_id);
CREATE INDEX IF NOT EXISTS idx_validations_project ON validations(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_project ON snapshots(project_id);
CREATE INDEX IF NOT EXISTS idx_change_requests_project ON change_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_impact_reports_change ON impact_reports(change_request_id);
CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_objects_fts USING fts5(id UNINDEXED, topic, summary, content);
CREATE VIRTUAL TABLE IF NOT EXISTS research_entries_fts USING fts5(id UNINDEXED, topic, objective, summary);
CREATE VIRTUAL TABLE IF NOT EXISTS implementation_documents_fts USING fts5(id UNINDEXED, title, content);
CREATE VIRTUAL TABLE IF NOT EXISTS raw_inputs_fts USING fts5(id UNINDEXED, content);
`

const projectIngestionVisionContextSQL = `
ALTER TABLE raw_inputs ADD COLUMN raw_content TEXT NOT NULL DEFAULT '';
ALTER TABLE ingested_sources ADD COLUMN normalized_content TEXT NOT NULL DEFAULT '';
ALTER TABLE ingested_sources ADD COLUMN classification TEXT NOT NULL DEFAULT 'unknown';
ALTER TABLE visions ADD COLUMN target_users TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN functional_goals TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN ux_goals TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN business_goals TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN constraints TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN assumptions TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN missing_information TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN visual_references TEXT NOT NULL DEFAULT '[]';
ALTER TABLE visions ADD COLUMN success_criteria TEXT NOT NULL DEFAULT '[]';
-- approved_requirements: requirements approved for planning
CREATE TABLE IF NOT EXISTS approved_requirements (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
-- approved_constraints: constraints approved for planning  
CREATE TABLE IF NOT EXISTS approved_constraints (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
-- approved_decisions: decisions approved for planning
CREATE TABLE IF NOT EXISTS approved_decisions (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
-- approved_preferences: preferences approved for planning
CREATE TABLE IF NOT EXISTS approved_preferences (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
-- approved_references: references approved for planning
CREATE TABLE IF NOT EXISTS approved_references (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
-- approved_goals: goals approved for planning
CREATE TABLE IF NOT EXISTS approved_goals (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, source_id TEXT NOT NULL DEFAULT '', content TEXT NOT NULL, state TEXT NOT NULL DEFAULT 'approved', supersedes_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(project_id, content));
CREATE INDEX IF NOT EXISTS idx_ingested_sources_project ON ingested_sources(project_id);
CREATE INDEX IF NOT EXISTS idx_ingested_sources_classification ON ingested_sources(classification);
CREATE INDEX IF NOT EXISTS idx_approved_requirements_project ON approved_requirements(project_id);
CREATE INDEX IF NOT EXISTS idx_approved_constraints_project ON approved_constraints(project_id);
CREATE INDEX IF NOT EXISTS idx_approved_decisions_project ON approved_decisions(project_id);
CREATE INDEX IF NOT EXISTS idx_approved_preferences_project ON approved_preferences(project_id);
CREATE INDEX IF NOT EXISTS idx_approved_references_project ON approved_references(project_id);
CREATE INDEX IF NOT EXISTS idx_approved_goals_project ON approved_goals(project_id);
`

const projectResearchKnowledgeSQL = `
CREATE TABLE IF NOT EXISTS research_jobs (id TEXT PRIMARY KEY, project_id TEXT NOT NULL, topic TEXT NOT NULL, summary TEXT NOT NULL DEFAULT '', confidence REAL NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'draft', created_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_research_jobs_project ON research_jobs(project_id);
CREATE TABLE IF NOT EXISTS research_recommendations (id TEXT PRIMARY KEY, research_id TEXT NOT NULL, content TEXT NOT NULL, created_at TEXT NOT NULL);
CREATE INDEX IF NOT EXISTS idx_research_recommendations_research ON research_recommendations(research_id);
CREATE TABLE IF NOT EXISTS knowledge_links (id TEXT PRIMARY KEY, knowledge_id TEXT NOT NULL, link_type TEXT NOT NULL, target_id TEXT NOT NULL, created_at TEXT NOT NULL, UNIQUE(knowledge_id, link_type, target_id));
ALTER TABLE knowledge_objects ADD COLUMN project_id TEXT NOT NULL DEFAULT '';
ALTER TABLE knowledge_objects ADD COLUMN title TEXT NOT NULL DEFAULT '';
ALTER TABLE knowledge_objects ADD COLUMN research_ids TEXT NOT NULL DEFAULT '[]';
ALTER TABLE knowledge_objects ADD COLUMN related_decisions TEXT NOT NULL DEFAULT '[]';
ALTER TABLE knowledge_objects ADD COLUMN related_requirements TEXT NOT NULL DEFAULT '[]';
ALTER TABLE knowledge_objects ADD COLUMN related_constraints TEXT NOT NULL DEFAULT '[]';
CREATE INDEX IF NOT EXISTS idx_knowledge_objects_project ON knowledge_objects(project_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_objects_title ON knowledge_objects(title);
`

const projectPlanningFrameworkSQL = `
ALTER TABLE master_plans ADD COLUMN vision_reference TEXT NOT NULL DEFAULT '';
ALTER TABLE master_plans ADD COLUMN objectives TEXT NOT NULL DEFAULT '[]';
ALTER TABLE master_plans ADD COLUMN scope TEXT NOT NULL DEFAULT '[]';
ALTER TABLE master_plans ADD COLUMN out_of_scope TEXT NOT NULL DEFAULT '[]';
ALTER TABLE master_plans ADD COLUMN recommended_specific_plans TEXT NOT NULL DEFAULT '[]';
ALTER TABLE master_plans ADD COLUMN risks TEXT NOT NULL DEFAULT '[]';
ALTER TABLE master_plans ADD COLUMN assumptions TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN goal TEXT NOT NULL DEFAULT '';
ALTER TABLE specific_plans ADD COLUMN requirements TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN constraints TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN decisions TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN knowledge_used TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN research_used TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN implementation_strategy TEXT NOT NULL DEFAULT '';
ALTER TABLE specific_plans ADD COLUMN risks TEXT NOT NULL DEFAULT '[]';
ALTER TABLE specific_plans ADD COLUMN validation_criteria TEXT NOT NULL DEFAULT '[]';
ALTER TABLE implementation_documents ADD COLUMN objective TEXT NOT NULL DEFAULT '';
ALTER TABLE implementation_documents ADD COLUMN architecture TEXT NOT NULL DEFAULT '';
ALTER TABLE implementation_documents ADD COLUMN expected_files TEXT NOT NULL DEFAULT '[]';
ALTER TABLE implementation_documents ADD COLUMN expected_directories TEXT NOT NULL DEFAULT '[]';
ALTER TABLE implementation_documents ADD COLUMN validations TEXT NOT NULL DEFAULT '[]';
ALTER TABLE implementation_documents ADD COLUMN known_risks TEXT NOT NULL DEFAULT '[]';
ALTER TABLE implementation_documents ADD COLUMN testing_strategy TEXT NOT NULL DEFAULT '';
ALTER TABLE implementation_documents ADD COLUMN rollback_strategy TEXT NOT NULL DEFAULT '';
`

const projectWorkflowsSQL = `
CREATE TABLE IF NOT EXISTS workflow_runs (id TEXT PRIMARY KEY, workflow_type TEXT NOT NULL, status TEXT NOT NULL, started_at TEXT NOT NULL, finished_at TEXT NOT NULL DEFAULT '');
CREATE INDEX IF NOT EXISTS idx_workflow_runs_type ON workflow_runs(workflow_type);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs(status);
`

const projectWorkflowRunsStepsSQL = `
ALTER TABLE workflow_runs ADD COLUMN steps TEXT NOT NULL DEFAULT '[]';
`

const projectModelStrategySQL = `
CREATE TABLE IF NOT EXISTS model_profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  provider TEXT NOT NULL,
  model TEXT NOT NULL,
  config TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS prompt_contracts (
  id TEXT PRIMARY KEY,
  contract_type TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS output_schemas (
  id TEXT PRIMARY KEY,
  schema_type TEXT NOT NULL,
  fields TEXT NOT NULL DEFAULT '{}',
  required TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_prompt_contracts_type ON prompt_contracts(contract_type);
CREATE INDEX IF NOT EXISTS idx_output_schemas_type ON output_schemas(schema_type);
`

const projectOrchestratorSQL = `
CREATE TABLE IF NOT EXISTS jobs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  workflow_type TEXT NOT NULL,
  capability TEXT NOT NULL DEFAULT '',
  strategy TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  error TEXT NOT NULL DEFAULT '',
  started_at TEXT NOT NULL,
  finished_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS job_runs (
  id TEXT PRIMARY KEY,
  job_id TEXT NOT NULL,
  step TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'running',
  output TEXT NOT NULL DEFAULT '',
  error TEXT NOT NULL DEFAULT '',
  started_at TEXT NOT NULL,
  finished_at TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS capabilities (
  id TEXT PRIMARY KEY,
  type TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_jobs_project ON jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_job_runs_job ON job_runs(job_id);
CREATE INDEX IF NOT EXISTS idx_capabilities_type ON capabilities(type);
`

const projectContextEngineSQL = `
CREATE TABLE IF NOT EXISTS context_views_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  name TEXT NOT NULL,
  view_type TEXT NOT NULL DEFAULT 'general',
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
ALTER TABLE context_views ADD COLUMN view_type TEXT NOT NULL DEFAULT 'general';
CREATE INDEX IF NOT EXISTS idx_context_views_v2_project ON context_views_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_context_views_v2_type ON context_views_v2(view_type);
CREATE INDEX IF NOT EXISTS idx_context_chunks_view ON context_chunks(context_view_id);
`

func runMigrations(db *sql.DB, migrations []Migration) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  applied_at TEXT NOT NULL
)`); err != nil {
		return err
	}

	for _, migration := range migrations {
		var exists int
		if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE id = ?`, migration.ID).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return err
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (id, name, applied_at) VALUES (?, ?, ?)`, migration.ID, migration.Name, time.Now().UTC().Format(time.RFC3339)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

func UpsertKnownProject(db *sql.DB, id, name, path string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`INSERT INTO known_projects (id, name, path, mode, slug, created_at, updated_at, last_seen_at)
VALUES (?, ?, ?, 'external', ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET name = excluded.name, updated_at = excluded.updated_at, last_seen_at = excluded.last_seen_at, slug = excluded.slug`, id, name, path, name, now, now, now)
	return err
}

// UpsertKnownProjectWithMode registers or refreshes a project in the global
// registry, recording whether it uses external or local storage.
func UpsertKnownProjectWithMode(db *sql.DB, id, name, path, slug, mode string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`INSERT INTO known_projects (id, name, path, mode, slug, created_at, updated_at, last_seen_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET
  name = excluded.name,
  mode = excluded.mode,
  slug = excluded.slug,
  updated_at = excluded.updated_at,
  last_seen_at = excluded.last_seen_at`, id, name, path, mode, slug, now, now, now)
	return err
}

func UpsertProjectState(db *sql.DB, id, name, rootPath, status string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := db.Exec(`INSERT INTO project_state (id, name, root_path, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET name = excluded.name, root_path = excluded.root_path, status = excluded.status, updated_at = excluded.updated_at`, id, name, rootPath, status, now, now)
	return err
}

func ProjectID(rootPath string) string {
	return fmt.Sprintf("project:%s", filepath.Clean(rootPath))
}

const projectChangeEngineSQL = `
CREATE TABLE IF NOT EXISTS change_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  change_type TEXT NOT NULL,
  summary TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  severity TEXT NOT NULL DEFAULT 'medium',
  status TEXT NOT NULL DEFAULT 'pending',
  source TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS change_reports (
  id TEXT PRIMARY KEY,
  change_event_id TEXT NOT NULL,
  project_id TEXT NOT NULL,
  analysis TEXT NOT NULL DEFAULT '{}',
  affected_entities TEXT NOT NULL DEFAULT '[]',
  review_required INTEGER NOT NULL DEFAULT 0,
  summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY (change_event_id) REFERENCES change_events(id)
);
CREATE TABLE IF NOT EXISTS snapshots_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  entity_snapshot TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS entity_states (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'valid',
  last_change_id TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  UNIQUE(project_id, entity_type, entity_id)
);
CREATE INDEX IF NOT EXISTS idx_change_events_project ON change_events(project_id);
CREATE INDEX IF NOT EXISTS idx_change_events_type ON change_events(change_type);
CREATE INDEX IF NOT EXISTS idx_change_events_status ON change_events(status);
CREATE INDEX IF NOT EXISTS idx_change_reports_event ON change_reports(change_event_id);
CREATE INDEX IF NOT EXISTS idx_change_reports_project ON change_reports(project_id);
CREATE INDEX IF NOT EXISTS idx_snapshots_v2_project ON snapshots_v2(project_id);

-- entity_links: cross-entity traceability for Impact Graph traversal
CREATE TABLE IF NOT EXISTS entity_links (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  link_type TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(source_type, source_id, target_type, target_id, link_type)
);
CREATE INDEX IF NOT EXISTS idx_entity_links_project ON entity_links(project_id);
CREATE INDEX IF NOT EXISTS idx_entity_links_source ON entity_links(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_entity_links_target ON entity_links(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_project ON entity_states(project_id);
CREATE INDEX IF NOT EXISTS idx_entity_states_type ON entity_states(entity_type);
CREATE INDEX IF NOT EXISTS idx_entity_states_status ON entity_states(status);
`

const projectMCPServerSQL = `
CREATE TABLE IF NOT EXISTS mcp_tools (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  schema_def TEXT NOT NULL DEFAULT '{}',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS mcp_runs (
  id TEXT PRIMARY KEY,
  tool_name TEXT NOT NULL,
  arguments TEXT NOT NULL DEFAULT '{}',
  success INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_mcp_tools_name ON mcp_tools(name);
CREATE INDEX IF NOT EXISTS idx_mcp_tools_enabled ON mcp_tools(enabled);
CREATE INDEX IF NOT EXISTS idx_mcp_runs_tool ON mcp_runs(tool_name);
CREATE INDEX IF NOT EXISTS idx_mcp_runs_created ON mcp_runs(created_at);
`

const projectOpenCodeIntegrationSQL = `
CREATE TABLE IF NOT EXISTS opencode_detections (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL,
  found INTEGER NOT NULL DEFAULT 0,
  config_path TEXT NOT NULL DEFAULT '',
  is_initialized INTEGER NOT NULL DEFAULT 0,
  agent_name TEXT NOT NULL DEFAULT '',
  skill_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS opencode_integration_state (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL UNIQUE,
  mode TEXT NOT NULL DEFAULT 'standalone',
  enabled INTEGER NOT NULL DEFAULT 1,
  read_only INTEGER NOT NULL DEFAULT 1,
  last_detected_at TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS opencode_doctor_checks (
  id TEXT PRIMARY KEY,
  project_root TEXT NOT NULL,
  check_name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pass',
  message TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_opencode_detections_project ON opencode_detections(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_integration_project ON opencode_integration_state(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_doctor_project ON opencode_doctor_checks(project_root);
CREATE INDEX IF NOT EXISTS idx_opencode_doctor_name ON opencode_doctor_checks(check_name);
`

const projectAgentSystemSQL = `
-- agent_runs_v2: enhanced agent run records (Phase 21)
CREATE TABLE IF NOT EXISTS agent_runs_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'processed',
  response TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- agent_messages: conversation messages within agent runs
CREATE TABLE IF NOT EXISTS agent_messages (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL,
  role TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  FOREIGN KEY (run_id) REFERENCES agent_runs_v2(id)
);

-- agent_delegated_jobs: delegated sub-agent jobs
CREATE TABLE IF NOT EXISTS agent_delegated_jobs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent TEXT NOT NULL DEFAULT '',
  capability TEXT NOT NULL DEFAULT '',
  workflow_type TEXT NOT NULL DEFAULT '',
  job_type TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  result_summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  completed_at TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_agent_runs_v2_project ON agent_runs_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_runs_v2_status ON agent_runs_v2(status);
CREATE INDEX IF NOT EXISTS idx_agent_messages_run ON agent_messages(run_id);
CREATE INDEX IF NOT EXISTS idx_agent_delegated_jobs_project ON agent_delegated_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_agent_delegated_jobs_status ON agent_delegated_jobs(status);

-- Compatibility views: map physical table names to definitive schema names
CREATE VIEW IF NOT EXISTS agent_runs AS SELECT * FROM agent_runs_v2;
CREATE VIEW IF NOT EXISTS subagent_outputs AS
  SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at
  FROM agent_delegated_jobs;
`

const projectContinuousPlanningSQL = `
-- continuous_events: detected events for continuous planning
CREATE TABLE IF NOT EXISTS continuous_events (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  event_type TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  details TEXT NOT NULL DEFAULT '{}',
  source TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

-- plan_update_proposals: proposed plan updates
CREATE TABLE IF NOT EXISTS plan_update_proposals (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  affected_plans TEXT NOT NULL DEFAULT '[]',
  affected_tasks TEXT NOT NULL DEFAULT '[]',
  affected_decisions TEXT NOT NULL DEFAULT '[]',
  suggested_updates TEXT NOT NULL DEFAULT '',
  requires_research INTEGER NOT NULL DEFAULT 0,
  requires_approval INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- context_deliveries: delivered context at various levels
CREATE TABLE IF NOT EXISTS context_deliveries (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  level TEXT NOT NULL DEFAULT 'L0_Executive',
  content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_continuous_events_project ON continuous_events(project_id);
CREATE INDEX IF NOT EXISTS idx_continuous_events_type ON continuous_events(event_type);
CREATE INDEX IF NOT EXISTS idx_plan_update_proposals_project ON plan_update_proposals(project_id);
CREATE INDEX IF NOT EXISTS idx_plan_update_proposals_status ON plan_update_proposals(status);
CREATE INDEX IF NOT EXISTS idx_context_deliveries_project ON context_deliveries(project_id);
CREATE INDEX IF NOT EXISTS idx_context_deliveries_level ON context_deliveries(level);
`

const projectCompatibilitySQL = `
-- tool_runs: compatibility view over mcp_runs (Phase 19 MCP Server)
CREATE VIEW IF NOT EXISTS tool_runs AS
SELECT id, tool_name, arguments, success, error_message, created_at FROM mcp_runs;

-- tool_audit: compatibility view joining mcp_tools and mcp_runs
CREATE VIEW IF NOT EXISTS tool_audit AS
SELECT
  r.id AS run_id,
  t.id AS tool_id,
  t.name AS tool_name,
  t.description,
  r.arguments,
  r.success,
  r.error_message,
  r.created_at AS executed_at
FROM mcp_runs r
LEFT JOIN mcp_tools t ON r.tool_name = t.name;

-- provider_registry: model provider registry (Phase 19 / MCP)
CREATE TABLE IF NOT EXISTS provider_registry (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  provider_type TEXT NOT NULL DEFAULT '',
  endpoint TEXT NOT NULL DEFAULT '',
  config TEXT NOT NULL DEFAULT '{}',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- skill_registry: skill registry (Phase 20 / OpenCode Integration)
CREATE TABLE IF NOT EXISTS skill_registry (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  version TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  checksum TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_provider_registry_name ON provider_registry(name);
CREATE INDEX IF NOT EXISTS idx_provider_registry_type ON provider_registry(provider_type);
CREATE INDEX IF NOT EXISTS idx_skill_registry_name ON skill_registry(name);
CREATE INDEX IF NOT EXISTS idx_skill_registry_source ON skill_registry(source);
`

const projectAgentContinuousCompatibilitySQL = `
-- delegated_jobs: contract-compatible view over Phase 21 delegated jobs.
CREATE VIEW IF NOT EXISTS delegated_jobs AS
SELECT id, project_id, intent, capability, workflow_type, job_type, status, result_summary, created_at, completed_at
FROM agent_delegated_jobs;

-- agent_responses: contract-compatible response view from agent run records.
CREATE VIEW IF NOT EXISTS agent_responses AS
SELECT
  id,
  id AS run_id,
  response AS content,
  status,
  created_at,
  updated_at
FROM agent_runs_v2;

-- continuous_status: persisted snapshots of continuous planning status.
CREATE TABLE IF NOT EXISTS continuous_status (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  active_plan TEXT NOT NULL DEFAULT '',
  active_phase TEXT NOT NULL DEFAULT '',
  next_task TEXT NOT NULL DEFAULT '',
  blocked_items TEXT NOT NULL DEFAULT '[]',
  approvals_needed TEXT NOT NULL DEFAULT '[]',
  outdated_plans TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

-- context_delivery_logs: contract-compatible view over context deliveries.
CREATE VIEW IF NOT EXISTS context_delivery_logs AS
SELECT id, project_id, level, content, created_at
FROM context_deliveries;

CREATE INDEX IF NOT EXISTS idx_continuous_status_project ON continuous_status(project_id);
`

const projectVisionDiscoverySQL = `
CREATE TABLE IF NOT EXISTS vision_discovery_sessions (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'draft',
    summary TEXT NOT NULL DEFAULT '', raw_context TEXT NOT NULL DEFAULT '',
    findings TEXT NOT NULL DEFAULT '[]', created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS vision_assumptions (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, session_id TEXT NOT NULL,
    description TEXT NOT NULL, category TEXT NOT NULL DEFAULT '',
    confidence REAL NOT NULL DEFAULT 0.5, status TEXT NOT NULL DEFAULT 'unvalidated',
    validated_by TEXT, validated_at TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS vision_ambiguities (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, session_id TEXT NOT NULL,
    description TEXT NOT NULL, category TEXT NOT NULL DEFAULT '',
    resolution TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'open',
    resolved_at TEXT, created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS vision_approvals (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, session_id TEXT NOT NULL,
    vision_id TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
    approved_by TEXT, approved_at TEXT, feedback TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_vision_discovery_project ON vision_discovery_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_vision_assumptions_session ON vision_assumptions(session_id);
CREATE INDEX IF NOT EXISTS idx_vision_ambiguities_session ON vision_ambiguities(session_id);
CREATE INDEX IF NOT EXISTS idx_vision_approvals_vision ON vision_approvals(vision_id);
`

const projectMasterPlanV2SQL = `
CREATE TABLE IF NOT EXISTS master_plan_versions (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, version INT NOT NULL,
    plan_id TEXT NOT NULL, title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    phases TEXT NOT NULL DEFAULT '[]', timeline TEXT NOT NULL DEFAULT '{}',
    risks TEXT NOT NULL DEFAULT '[]', dependencies TEXT NOT NULL DEFAULT '[]',
    status TEXT NOT NULL DEFAULT 'draft', changelog TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);
CREATE TABLE IF NOT EXISTS master_plan_changes (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, plan_id TEXT NOT NULL,
    version_from INT NOT NULL, version_to INT NOT NULL, change_type TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '', author TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);
CREATE TABLE IF NOT EXISTS master_plan_approvals (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, plan_id TEXT NOT NULL,
    version INT NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
    approved_by TEXT, approved_at TEXT, feedback TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')), FOREIGN KEY (plan_id) REFERENCES master_plans(id)
);
CREATE TABLE IF NOT EXISTS plan_evolution_events (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL,
    entity_type TEXT NOT NULL, entity_id TEXT NOT NULL, event_type TEXT NOT NULL,
    description TEXT NOT NULL, details TEXT NOT NULL DEFAULT '{}',
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_mp_versions_plan ON master_plan_versions(plan_id);
CREATE INDEX IF NOT EXISTS idx_mp_changes_plan ON master_plan_changes(plan_id);
CREATE INDEX IF NOT EXISTS idx_mp_approvals_plan ON master_plan_approvals(plan_id);
CREATE INDEX IF NOT EXISTS idx_plan_evolution_project ON plan_evolution_events(project_id);
`

const projectSpecificPlanV2SQL = `
CREATE TABLE IF NOT EXISTS specific_plan_versions (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, version INT NOT NULL,
    plan_id TEXT NOT NULL, domain TEXT NOT NULL DEFAULT 'general',
    title TEXT NOT NULL, description TEXT NOT NULL DEFAULT '',
    tasks TEXT NOT NULL DEFAULT '[]', dependencies TEXT NOT NULL DEFAULT '[]',
    risks TEXT NOT NULL DEFAULT '[]', status TEXT NOT NULL DEFAULT 'draft',
    changelog TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id)
);
CREATE TABLE IF NOT EXISTS specific_plan_research_links (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, plan_id TEXT NOT NULL,
    research_id TEXT NOT NULL, section TEXT NOT NULL DEFAULT '',
    relevance REAL NOT NULL DEFAULT 0.0, created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id), FOREIGN KEY (research_id) REFERENCES research(id)
);
CREATE TABLE IF NOT EXISTS specific_plan_regenerations (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, plan_id TEXT NOT NULL,
    version_from INT NOT NULL, version_to INT NOT NULL,
    reason TEXT NOT NULL DEFAULT '', scope TEXT NOT NULL DEFAULT 'full',
    status TEXT NOT NULL DEFAULT 'completed', created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (plan_id) REFERENCES specific_plans(id)
);
CREATE INDEX IF NOT EXISTS idx_sp_versions_plan ON specific_plan_versions(plan_id);
CREATE INDEX IF NOT EXISTS idx_sp_research_plan ON specific_plan_research_links(plan_id);
CREATE INDEX IF NOT EXISTS idx_sp_regenerations_plan ON specific_plan_regenerations(plan_id);
`

const projectContextDeliveryEngineSQL = `
CREATE TABLE IF NOT EXISTS context_delivery_sessions (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, level TEXT NOT NULL,
    budget_tokens INT NOT NULL DEFAULT 0, tokens_used INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL DEFAULT '', metadata TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'delivered', created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE TABLE IF NOT EXISTS context_delivery_usage (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, session_id TEXT,
    level TEXT NOT NULL, tokens INT NOT NULL DEFAULT 0, source TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (session_id) REFERENCES context_delivery_sessions(id)
);
CREATE TABLE IF NOT EXISTS context_delivery_budgets (
    id TEXT PRIMARY KEY, project_id TEXT NOT NULL, level TEXT NOT NULL,
    max_tokens INT NOT NULL DEFAULT 4096, current_usage INT NOT NULL DEFAULT 0,
    strategy TEXT NOT NULL DEFAULT 'fixed', created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_ctx_session_project ON context_delivery_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_ctx_usage_project ON context_delivery_usage(project_id);
CREATE INDEX IF NOT EXISTS idx_ctx_budgets_level ON context_delivery_budgets(level);
`

const projectUserIntentEngineSQL = `
CREATE TABLE IF NOT EXISTS intent_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  primary_intent TEXT NOT NULL DEFAULT '{}',
  secondary_goals TEXT NOT NULL DEFAULT '[]',
  constraints_json TEXT NOT NULL DEFAULT '[]',
  expectations TEXT NOT NULL DEFAULT '[]',
  success_criteria TEXT NOT NULL DEFAULT '[]',
  priorities TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  approved INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_intent_profiles_project ON intent_profiles(project_id);
CREATE INDEX IF NOT EXISTS idx_intent_profiles_status ON intent_profiles(status);
`

const projectVisionDocumentsSQL = `
CREATE TABLE IF NOT EXISTS vision_documents (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  intent_profile_id TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  functional_vision TEXT NOT NULL DEFAULT '',
  visual_vision TEXT NOT NULL DEFAULT '',
  technical_vision TEXT NOT NULL DEFAULT '',
  operational_vision TEXT NOT NULL DEFAULT '',
  business_vision TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'draft',
  approved INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_vision_documents_project ON vision_documents(project_id);
CREATE INDEX IF NOT EXISTS idx_vision_documents_status ON vision_documents(status);
`

const projectApprovalWorkflowSQL = `
CREATE TABLE IF NOT EXISTS approval_records (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  state TEXT NOT NULL DEFAULT 'draft',
  reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_approval_records_project ON approval_records(project_id);
CREATE INDEX IF NOT EXISTS idx_approval_records_target ON approval_records(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_approval_records_state ON approval_records(state);
`

const projectRequirementDiscoverySQL = `
CREATE TABLE IF NOT EXISTS requirement_candidates (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source TEXT NOT NULL DEFAULT '',
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  reason TEXT NOT NULL DEFAULT '',
  dependencies TEXT NOT NULL DEFAULT '[]',
  ambiguities TEXT NOT NULL DEFAULT '[]',
  state TEXT NOT NULL DEFAULT 'candidate',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_requirement_candidates_project ON requirement_candidates(project_id);
CREATE INDEX IF NOT EXISTS idx_requirement_candidates_state ON requirement_candidates(state);
`

const projectContextPackagesV2SQL = `
CREATE TABLE IF NOT EXISTS context_packages_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  package_type TEXT NOT NULL,
  model_target TEXT NOT NULL DEFAULT 'generic',
  summary TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  priority INTEGER NOT NULL DEFAULT 5,
  token_budget INTEGER NOT NULL DEFAULT 4096,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_context_packages_v2_project ON context_packages_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_context_packages_v2_type ON context_packages_v2(package_type);
`

const projectResearchOrchestrationV2SQL = `
CREATE TABLE IF NOT EXISTS research_orchestration_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  agent_type TEXT NOT NULL,
  topic TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  evidence TEXT NOT NULL DEFAULT '[]',
  confidence INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_research_orchestration_project ON research_orchestration_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_research_orchestration_agent ON research_orchestration_runs(agent_type);
`

const projectReferenceEngineV2SQL = `
CREATE TABLE IF NOT EXISTS project_references_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_type TEXT NOT NULL,
  uri TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT 'functional',
  notes TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'needs_review',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_project_references_v2_project ON project_references_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_project_references_v2_status ON project_references_v2(status);
`

const projectPlanEvolutionV3SQL = `
CREATE TABLE IF NOT EXISTS plan_evolution_blueprints_v3 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  objective TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT '[]',
  exclusions TEXT NOT NULL DEFAULT '[]',
  dependencies TEXT NOT NULL DEFAULT '[]',
  stack TEXT NOT NULL DEFAULT '[]',
  versions TEXT NOT NULL DEFAULT '[]',
  libraries TEXT NOT NULL DEFAULT '[]',
  folders TEXT NOT NULL DEFAULT '[]',
  files TEXT NOT NULL DEFAULT '[]',
  validations TEXT NOT NULL DEFAULT '[]',
  tests TEXT NOT NULL DEFAULT '[]',
  risks TEXT NOT NULL DEFAULT '[]',
  rollback TEXT NOT NULL DEFAULT '[]',
  approved_inputs TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_plan_evolution_blueprints_v3_project ON plan_evolution_blueprints_v3(project_id);
`

const projectImplementationPackagesV2SQL = `
CREATE TABLE IF NOT EXISTS implementation_packages_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_id TEXT NOT NULL DEFAULT '',
  model_target TEXT NOT NULL DEFAULT 'opencode',
  what_to_do TEXT NOT NULL DEFAULT '',
  how_to_do_it TEXT NOT NULL DEFAULT '',
  files_to_touch TEXT NOT NULL DEFAULT '[]',
  files_not_to_touch TEXT NOT NULL DEFAULT '[]',
  examples TEXT NOT NULL DEFAULT '[]',
  commands TEXT NOT NULL DEFAULT '[]',
  validations TEXT NOT NULL DEFAULT '[]',
  rollback_notes TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_implementation_packages_v2_project ON implementation_packages_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_implementation_packages_v2_plan ON implementation_packages_v2(plan_id);
`

const projectChangeImpactV2SQL = `
CREATE TABLE IF NOT EXISTS change_impact_reports_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  change_type TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  architecture_concerns TEXT NOT NULL DEFAULT '[]',
  backend_concerns TEXT NOT NULL DEFAULT '[]',
  migration_concerns TEXT NOT NULL DEFAULT '[]',
  docs_concerns TEXT NOT NULL DEFAULT '[]',
  api_concerns TEXT NOT NULL DEFAULT '[]',
  plan_concerns TEXT NOT NULL DEFAULT '[]',
  validation_commands TEXT NOT NULL DEFAULT '[]',
  rollback_strategy TEXT NOT NULL DEFAULT '[]',
  affected_plans TEXT NOT NULL DEFAULT '[]',
  affected_tasks TEXT NOT NULL DEFAULT '[]',
  severity TEXT NOT NULL DEFAULT 'medium',
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_change_impact_reports_v2_project ON change_impact_reports_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_change_impact_reports_v2_type ON change_impact_reports_v2(change_type);
`

const projectContinuousRegenerationV2SQL = `
CREATE TABLE IF NOT EXISTS continuous_regenerations_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  scope TEXT NOT NULL DEFAULT 'affected-sections',
  affected_sections TEXT NOT NULL DEFAULT '[]',
  preserved_sections TEXT NOT NULL DEFAULT '[]',
  snapshot_required INTEGER NOT NULL DEFAULT 1,
  approval_required INTEGER NOT NULL DEFAULT 1,
  status TEXT NOT NULL DEFAULT 'draft',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_continuous_regenerations_v2_project ON continuous_regenerations_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_continuous_regenerations_v2_status ON continuous_regenerations_v2(status);
`

const projectSubagentOrchestratorV2SQL = `
CREATE TABLE IF NOT EXISTS subagent_tasks_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  agent_type TEXT NOT NULL,
  objective TEXT NOT NULL DEFAULT '',
  capability TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  provenance TEXT NOT NULL DEFAULT '',
  validation_status TEXT NOT NULL DEFAULT 'pending',
  isolated INTEGER NOT NULL DEFAULT 1,
  temporary INTEGER NOT NULL DEFAULT 1,
  memory_policy TEXT NOT NULL DEFAULT 'no-independent-persistent-memory',
  result_summary TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_subagent_tasks_v2_project ON subagent_tasks_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_subagent_tasks_v2_type ON subagent_tasks_v2(agent_type);
`

const projectOpenCodeWorkflowsV2SQL = `
CREATE TABLE IF NOT EXISTS opencode_workflows_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  commands TEXT NOT NULL DEFAULT '[]',
  status TEXT NOT NULL DEFAULT 'synced',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_opencode_workflows_v2_project ON opencode_workflows_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_opencode_workflows_v2_status ON opencode_workflows_v2(status);
`

const projectMemoryV2SQL = `
CREATE TABLE IF NOT EXISTS project_memory_v2 (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  entry_type TEXT NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  question TEXT NOT NULL DEFAULT '',
  answer TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  citation TEXT NOT NULL DEFAULT '',
  source TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'active',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_project ON project_memory_v2(project_id);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_type ON project_memory_v2(entry_type);
CREATE INDEX IF NOT EXISTS idx_project_memory_v2_status ON project_memory_v2(status);
`

const projectFTS5SQL = `
-- Rebuild existing knowledge_objects_fts index from current data
INSERT INTO knowledge_objects_fts(knowledge_objects_fts) VALUES('rebuild');

-- Keep knowledge_objects_fts in sync
-- NOTE: modernc.org/sqlite does not support the fts5 'delete' command on
-- plain (non-content) fts5 tables, so we use DELETE FROM ... WHERE rowid = ?
CREATE TRIGGER IF NOT EXISTS knowledge_objects_fts_insert AFTER INSERT ON knowledge_objects BEGIN
  INSERT INTO knowledge_objects_fts(rowid, id, topic, summary, content)
  VALUES (new.rowid, new.id, new.topic, new.summary, new.content);
END;

CREATE TRIGGER IF NOT EXISTS knowledge_objects_fts_delete AFTER DELETE ON knowledge_objects BEGIN
  DELETE FROM knowledge_objects_fts WHERE rowid = old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS knowledge_objects_fts_update AFTER UPDATE ON knowledge_objects BEGIN
  DELETE FROM knowledge_objects_fts WHERE rowid = old.rowid;
  INSERT INTO knowledge_objects_fts(rowid, id, topic, summary, content)
  VALUES (new.rowid, new.id, new.topic, new.summary, new.content);
END;

-- FTS5 external content table for project_memory_v2
CREATE VIRTUAL TABLE IF NOT EXISTS project_memory_v2_fts USING fts5(
  id UNINDEXED,
  title,
  question,
  answer,
  content,
  citation,
  source,
  tokenize='porter unicode61',
  content=project_memory_v2,
  content_rowid=rowid
);

-- Populate project_memory_v2_fts from existing data
INSERT INTO project_memory_v2_fts(rowid, id, title, question, answer, content, citation, source)
SELECT rowid, id, title, question, answer, content, citation, source FROM project_memory_v2;

-- Keep project_memory_v2_fts in sync
CREATE TRIGGER IF NOT EXISTS project_memory_v2_fts_insert AFTER INSERT ON project_memory_v2 BEGIN
  INSERT INTO project_memory_v2_fts(rowid, id, title, question, answer, content, citation, source)
  VALUES (new.rowid, new.id, new.title, new.question, new.answer, new.content, new.citation, new.source);
END;

CREATE TRIGGER IF NOT EXISTS project_memory_v2_fts_delete AFTER DELETE ON project_memory_v2 BEGIN
  INSERT INTO project_memory_v2_fts(project_memory_v2_fts, rowid, id, title, question, answer, content, citation, source)
  VALUES ('delete', old.rowid, old.id, old.title, old.question, old.answer, old.content, old.citation, old.source);
END;

CREATE TRIGGER IF NOT EXISTS project_memory_v2_fts_update AFTER UPDATE ON project_memory_v2 BEGIN
  INSERT INTO project_memory_v2_fts(project_memory_v2_fts, rowid, id, title, question, answer, content, citation, source)
  VALUES ('delete', old.rowid, old.id, old.title, old.question, old.answer, old.content, old.citation, old.source);
  INSERT INTO project_memory_v2_fts(rowid, id, title, question, answer, content, citation, source)
  VALUES (new.rowid, new.id, new.title, new.question, new.answer, new.content, new.citation, new.source);
END;
`

const projectImpactGraphSQL = `
CREATE TABLE IF NOT EXISTS impact_edges (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_type TEXT NOT NULL,
  source_id TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  edge_type TEXT NOT NULL,
  weight INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL,
  UNIQUE(source_type, source_id, target_type, target_id, edge_type)
);
CREATE INDEX IF NOT EXISTS idx_impact_edges_project ON impact_edges(project_id);
CREATE INDEX IF NOT EXISTS idx_impact_edges_source ON impact_edges(source_type, source_id);
`

const projectCapabilitiesV2SQL = `
CREATE TABLE IF NOT EXISTS capabilities_v2 (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  schema_info TEXT NOT NULL DEFAULT '{}',
  version TEXT NOT NULL DEFAULT '1.0',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);
`

// sanitizeFTS5 sanitizes a user-supplied query string for safe use in
// FTS5 MATCH. Each word is quoted to prevent FTS5 syntax errors from
// special characters (operators, parens, etc.). Returns "" for empty
// or whitespace-only input so callers can fall back to listing.
func sanitizeFTS5(query string) string {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return ""
	}
	escaped := make([]string, len(terms))
	for i, t := range terms {
		t = strings.ReplaceAll(t, `"`, `""`)
		escaped[i] = `"` + t + `"`
	}
	return strings.Join(escaped, " ")
}
