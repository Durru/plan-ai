package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/plan-ai/plan-ai/internal/atomicfile"
	"github.com/plan-ai/plan-ai/internal/config"
)

// SetupResult holds paths to all generated OpenCode integration artifacts.
type SetupResult struct {
	OpenCodeConfigPath string `json:"opencode_config_path,omitempty"`
	MCPRegistryPath    string `json:"mcp_registry_path,omitempty"`
	AgentPath          string `json:"agent_path,omitempty"`
	ProfilesPath       string `json:"profiles_path,omitempty"`
	PromptsPath        string `json:"prompts_path,omitempty"`
	WorkflowsPath      string `json:"workflows_path,omitempty"`
	SyncMarkerPath     string `json:"sync_marker_path,omitempty"`
}

// setupOpenCodeConfig is the minimal OpenCode config written when none exists.
type setupOpenCodeConfig struct {
	AgentName string         `json:"agent_name"`
	AgentRole string         `json:"agent_role"`
	Mode      string         `json:"mode"`
	Skills    []string       `json:"skills"`
	Agent     map[string]any `json:"agent,omitempty"`
	MCP       map[string]any `json:"mcp,omitempty"`
}

// setupMCPRegistry describes the real local MCP server registration.
type setupMCPRegistry struct {
	Name    string              `json:"name"`
	Type    string              `json:"type"`
	Enabled bool                `json:"enabled"`
	Command []string            `json:"command"`
	Env     map[string]string   `json:"env,omitempty"`
	Tools   []setupMCPToolEntry `json:"tools"`
}

type setupMCPToolEntry struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

// setupAgent describes the Plan-AI agent for OpenCode.
type setupAgent struct {
	Name        string   `json:"name"`
	Role        string   `json:"role"`
	Description string   `json:"description"`
	Skills      []string `json:"skills"`
}

// setupProfile holds per-integration profile settings.
type setupProfile struct {
	Mode       string `json:"mode"`
	AutoDetect bool   `json:"auto_detect"`
	ReadOnly   bool   `json:"read_only"`
}

// setupPrompt is a keyed prompt template.
type setupPrompt struct {
	Key     string `json:"key"`
	Content string `json:"content"`
}

// syncMarker records what was generated and when.
type syncMarker struct {
	SyncedAt          string            `json:"synced_at"`
	OpenCodeConfigDir string            `json:"opencode_config_dir"`
	ProjectRoot       string            `json:"project_root"`
	Artifacts         map[string]string `json:"artifacts"`
	Status            string            `json:"status"`
}

// SetupService generates sandbox OpenCode integration artifacts.
// All paths are scoped to the sandbox env (OPENCODE_CONFIG_DIR, PLAN_AI_PROJECT_ROOT).
// It never touches real ~/.config/opencode or real ~/.plan-ai.
type SetupService struct{}

// NewSetupService creates a new SetupService.
func NewSetupService() *SetupService {
	return &SetupService{}
}

// Run generates all integration artifacts under the given sandbox paths.
//   - opencodeDir: the sandbox OPENCODE_CONFIG_DIR
//   - projectRoot: the sandbox PLAN_AI_PROJECT_ROOT
func (s *SetupService) Run(opencodeDir, projectRoot string) (*SetupResult, error) {
	if opencodeDir == "" {
		return nil, fmt.Errorf("opencode config dir is empty")
	}
	if projectRoot == "" {
		return nil, fmt.Errorf("project root is empty")
	}

	result := &SetupResult{}

	// 1. Detect or create minimal OpenCode config
	configPath, err := s.ensureOpenCodeConfig(opencodeDir, projectRoot)
	if err != nil {
		return result, fmt.Errorf("ensure opencode config: %w", err)
	}
	result.OpenCodeConfigPath = configPath

	// 2. MCP registry artifact
	mcpPath, err := s.writeMCPRegistry(opencodeDir)
	if err != nil {
		return result, fmt.Errorf("write mcp registry: %w", err)
	}
	result.MCPRegistryPath = mcpPath

	// 3. Agent registration artifact
	agentPath, err := s.writeAgentRegistration(opencodeDir)
	if err != nil {
		return result, fmt.Errorf("write agent registration: %w", err)
	}
	result.AgentPath = agentPath

	// 4. Profiles registration artifact
	profilesPath, err := s.writeProfiles(opencodeDir)
	if err != nil {
		return result, fmt.Errorf("write profiles: %w", err)
	}
	result.ProfilesPath = profilesPath

	// 5. Prompts registration artifact
	promptsPath, err := s.writePrompts(opencodeDir)
	if err != nil {
		return result, fmt.Errorf("write prompts: %w", err)
	}
	result.PromptsPath = promptsPath

	// 6. Workflow command surface for OpenCode.
	workflowsPath, err := s.writeWorkflows(opencodeDir)
	if err != nil {
		return result, fmt.Errorf("write workflows: %w", err)
	}
	result.WorkflowsPath = workflowsPath

	// 7. Sync/sync marker inside .plan-ai/
	syncPath, err := s.writeSyncMarker(projectRoot, opencodeDir, result)
	if err != nil {
		return result, fmt.Errorf("write sync marker: %w", err)
	}
	result.SyncMarkerPath = syncPath

	return result, nil
}

func (s *SetupService) writeWorkflows(opencodeDir string) (string, error) {
	path := filepath.Join(opencodeDir, "plan-ai-workflows.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("mkdir workflows dir: %w", err)
	}
	data, err := json.MarshalIndent(DefaultWorkflowCommands(), "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal workflows: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write workflows: %w", err)
	}
	return path, nil
}

// ensureOpenCodeConfig checks for an existing opencode.json[c] under opencodeDir.
// If found, returns its path. If not found, writes a minimal opencode.json
// and returns the new path.
func (s *SetupService) ensureOpenCodeConfig(opencodeDir, projectRoot string) (string, error) {
	// Check existing configs.
	candidates := []string{
		filepath.Join(opencodeDir, "opencode.json"),
		filepath.Join(opencodeDir, "opencode.jsonc"),
		filepath.Join(opencodeDir, ".opencode", "opencode.json"),
		filepath.Join(opencodeDir, ".opencode", "opencode.jsonc"),
	}
	for _, path := range candidates {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			if err := s.mergePlanAIIntoOpenCodeConfig(path, projectRoot); err != nil {
				return "", err
			}
			return path, nil
		}
	}

	// None found — write a minimal one.
	cfg := setupOpenCodeConfig{
		AgentName: "plan-ai",
		AgentRole: "Planning assistant for AI-assisted projects",
		Mode:      "tool",
		Skills:    []string{"planning", "research", "vision", "knowledge", "scanning"},
		Agent: map[string]any{
			"mode":             "tool",
			"read_only":        true,
			"auto_detect":      true,
			"warn_on_conflict": true,
		},
		MCP: map[string]any{
			"plan-ai": map[string]any{
				"type":    "local",
				"enabled": true,
				"command": planAIMCPCommand(),
				"env": map[string]string{
					"PLAN_AI_PROJECT_ROOT": projectRoot,
				},
			},
		},
	}

	configPath := filepath.Join(opencodeDir, "opencode.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return "", fmt.Errorf("mkdir opencode config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal opencode config: %w", err)
	}
	if _, err := atomicfile.WriteFileWithBackup(configPath, data, 0644, "pre-merge"); err != nil {
		return "", fmt.Errorf("write opencode config: %w", err)
	}

	return configPath, nil
}

func (s *SetupService) mergePlanAIIntoOpenCodeConfig(path, projectRoot string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read existing opencode config: %w", err)
	}
	var cfg map[string]any
	if err := json.Unmarshal(stripJSONCComments(data), &cfg); err != nil {
		return fmt.Errorf("parse existing opencode config: %w", err)
	}
	mcpRaw, ok := cfg["mcp"].(map[string]any)
	if !ok || mcpRaw == nil {
		mcpRaw = map[string]any{}
	}
	mcpRaw["plan-ai"] = map[string]any{
		"type":    "local",
		"enabled": true,
		"command": planAIMCPCommand(),
		"env": map[string]string{
			"PLAN_AI_PROJECT_ROOT": projectRoot,
		},
	}
	cfg["mcp"] = mcpRaw
	if _, ok := cfg["agent_name"]; !ok {
		cfg["agent_name"] = "plan-ai"
	}
	if _, ok := cfg["agent_role"]; !ok {
		cfg["agent_role"] = "Planning assistant for AI-assisted projects"
	}
	if _, ok := cfg["skills"]; !ok {
		cfg["skills"] = []string{"planning", "research", "vision", "knowledge", "scanning"}
	}
	merged, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal merged opencode config: %w", err)
	}
	if _, err := atomicfile.WriteFileWithBackup(path, merged, 0644, "pre-merge"); err != nil {
		return fmt.Errorf("write merged opencode config: %w", err)
	}
	return nil
}

func stripJSONCComments(data []byte) []byte {
	out := make([]byte, 0, len(data))
	inString := false
	escaped := false
	for i := 0; i < len(data); i++ {
		ch := data[i]
		if inString {
			out = append(out, ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out = append(out, ch)
			continue
		}
		if ch == '/' && i+1 < len(data) {
			next := data[i+1]
			if next == '/' {
				for i < len(data) && data[i] != '\n' {
					i++
				}
				if i < len(data) {
					out = append(out, data[i])
				}
				continue
			}
			if next == '*' {
				i += 2
				for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
					if data[i] == '\n' {
						out = append(out, '\n')
					}
					i++
				}
				i++
				continue
			}
		}
		out = append(out, ch)
	}
	return out
}

// writeMCPRegistry writes a Plan-AI MCP tool registry artifact.
func (s *SetupService) writeMCPRegistry(opencodeDir string) (string, error) {
	registry := setupMCPRegistry{
		Name:    "plan-ai",
		Type:    "local",
		Enabled: true,
		Command: planAIMCPCommand(),
		Tools:   defaultMCPToolEntries(),
	}

	path := filepath.Join(opencodeDir, "mcp-registry.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("mkdir mcp registry dir: %w", err)
	}

	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal mcp registry: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write mcp registry: %w", err)
	}

	return path, nil
}

func planAIMCPCommand() []string {
	return config.MCPCommand("")
}

func defaultMCPToolEntries() []setupMCPToolEntry {
	input := map[string]any{"type": "object"}
	return []setupMCPToolEntry{
		{Name: "plan_ai.init_project", Description: "Initialize Plan-AI for a project at the given root path.", InputSchema: input},
		{Name: "plan_ai.project_status", Description: "Get the current status of a Plan-AI project.", InputSchema: input},
		{Name: "plan_ai.create_master_plan", Description: "Create a new master plan from approved context.", InputSchema: input},
		{Name: "plan_ai.create_specific_plan", Description: "Create a specific plan under a master plan.", InputSchema: input},
		{Name: "plan_ai.research_topic", Description: "Create a new research entry for a topic.", InputSchema: input},
		{Name: "plan_ai.approve_plan", Description: "Approve a plan by ID.", InputSchema: input},
		{Name: "plan_ai.reject_plan", Description: "Reject a plan by ID.", InputSchema: input},
		{Name: "plan_ai.analyze_impact", Description: "Analyze the impact of a change event.", InputSchema: input},
		{Name: "plan_ai.get_next_task", Description: "Get the next pending task from the current plan.", InputSchema: input},
		{Name: "plan_ai.mark_task_done", Description: "Mark a task as completed.", InputSchema: input},
		{Name: "plan_ai.create_snapshot", Description: "Create a project state snapshot.", InputSchema: input},
		{Name: "plan_ai.list_plans", Description: "List all plans in the project.", InputSchema: input},
		{Name: "plan_ai.list_tasks", Description: "List all tasks, optionally filtered by plan, phase, or status.", InputSchema: input},
		{Name: "plan_ai.agent_process", Description: "Process a user message through the agent system for intent detection, routing, and delegation.", InputSchema: input},
		{Name: "plan_ai.agent_message", Description: "Process a user message through the agent system.", InputSchema: input},
		{Name: "plan_ai.agent_runs", Description: "List recent agent runs for a project.", InputSchema: input},
		{Name: "plan_ai.agent_status", Description: "Get recent agent activity and status.", InputSchema: input},
		{Name: "plan_ai.continuous_status", Description: "Get continuous planning status for a project.", InputSchema: input},
		{Name: "plan_ai.continuous_events", Description: "List recent continuous planning events.", InputSchema: input},
		{Name: "plan_ai.continuous_proposals", Description: "List plan update proposals.", InputSchema: input},
		{Name: "plan_ai.propose_plan_update", Description: "Create or list plan update proposals for continuous planning.", InputSchema: input},
		{Name: "plan_ai.approve_plan_update", Description: "Approve a pending plan update proposal.", InputSchema: input},
		{Name: "plan_ai.reject_plan_update", Description: "Reject a pending plan update proposal.", InputSchema: input},
		{Name: "plan_ai.continuous_context", Description: "Generate context at a specified level for a project.", InputSchema: input},
		{Name: "plan_ai.get_context_level", Description: "Generate context at a specified L0-L4 level for a project.", InputSchema: input},
		{Name: "plan_ai.get_context", Description: "Get context at a specified level (L0-L4) for a project.", InputSchema: input},
		{Name: "plan_ai.detect_changes", Description: "Detect and register changes in the project, returning impact analysis.", InputSchema: input},
		{Name: "plan_ai.update_plan", Description: "Update an existing plan's details (title, summary, status).", InputSchema: input},
		{Name: "plan_ai.rollback_snapshot", Description: "Rollback to a previous project state snapshot (not yet implemented).", InputSchema: input},
		{Name: "plan_ai.export_docs", Description: "Export project documentation (plans, decisions, research).", InputSchema: input},
		{Name: "plan_ai.create_product_intent", Description: "Create a Product Intent (Phase 51).", InputSchema: input},
		{Name: "plan_ai.list_product_intents", Description: "List all product intents for the project.", InputSchema: input},
		{Name: "plan_ai.get_product_intent", Description: "Get a single product intent by ID.", InputSchema: input},
		{Name: "plan_ai.submit_product_intent", Description: "Submit a product intent for approval.", InputSchema: input},
		{Name: "plan_ai.approve_product_intent", Description: "Approve a product intent.", InputSchema: input},
		{Name: "plan_ai.reject_product_intent", Description: "Reject a product intent.", InputSchema: input},
		{Name: "plan_ai.discover_intent", Description: "Analyze raw user input to extract structured intent, objectives, restrictions, and questions.", InputSchema: input},
		{Name: "plan_ai.list_discovery_results", Description: "List discovery results for the project.", InputSchema: input},
	}
}

// writeAgentRegistration writes a Plan-AI agent descriptor.
func (s *SetupService) writeAgentRegistration(opencodeDir string) (string, error) {
	agent := setupAgent{
		Name:        "plan-ai",
		Role:        "planning-assistant",
		Description: "Prepares implementation plans for AI-assisted projects using local-first SQLite persistence.",
		Skills:      []string{"planning", "research", "vision", "knowledge", "scanning", "context"},
	}

	agentsDir := filepath.Join(opencodeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir agents dir: %w", err)
	}

	path := filepath.Join(agentsDir, "plan-ai.json")
	data, err := json.MarshalIndent(agent, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal agent: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write agent: %w", err)
	}

	return path, nil
}

// writeProfiles writes Plan-AI integration profiles.
func (s *SetupService) writeProfiles(opencodeDir string) (string, error) {
	profiles := map[string]setupProfile{
		"plan-ai": {
			Mode:       "tool",
			AutoDetect: true,
			ReadOnly:   true,
		},
	}

	path := filepath.Join(opencodeDir, "profiles.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("mkdir profiles dir: %w", err)
	}

	data, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal profiles: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write profiles: %w", err)
	}

	return path, nil
}

// writePrompts writes prompt templates for OpenCode integration.
func (s *SetupService) writePrompts(opencodeDir string) (string, error) {
	prompts := []setupPrompt{
		{
			Key:     "plan-ai-setup",
			Content: "Plan-AI integration is enabled. Use `plan-ai` CLI commands to manage implementation plans, research, and knowledge.",
		},
		{
			Key:     "plan-ai-plan",
			Content: "Leverage Plan-AI planning artifacts, approved context, and reusable knowledge for implementation.",
		},
		{
			Key:     "plan-ai-doctor",
			Content: "Run `plan-ai doctor` to check store paths, migration status, and OpenCode integration health.",
		},
	}

	path := filepath.Join(opencodeDir, "prompts.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("mkdir prompts dir: %w", err)
	}

	data, err := json.MarshalIndent(prompts, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal prompts: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write prompts: %w", err)
	}

	return path, nil
}

// writeSyncMarker records what was generated and when inside .plan-ai/.
func (s *SetupService) writeSyncMarker(projectRoot, opencodeDir string, result *SetupResult) (string, error) {
	planAIDir := filepath.Join(projectRoot, ".plan-ai")
	if err := os.MkdirAll(planAIDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir .plan-ai: %w", err)
	}

	marker := syncMarker{
		SyncedAt:          time.Now().UTC().Format(time.RFC3339),
		OpenCodeConfigDir: opencodeDir,
		ProjectRoot:       projectRoot,
		Artifacts: map[string]string{
			"opencode_config": result.OpenCodeConfigPath,
			"mcp_registry":    result.MCPRegistryPath,
			"agent":           result.AgentPath,
			"profiles":        result.ProfilesPath,
			"prompts":         result.PromptsPath,
			"workflows":       result.WorkflowsPath,
		},
		Status: "synced",
	}

	path := filepath.Join(planAIDir, "opencode-sync.json")
	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal sync marker: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write sync marker: %w", err)
	}

	return path, nil
}

// SetupMCPConfig writes the Plan-AI MCP server entry to OpenCode's
// opencode.json (the file OpenCode reads at startup). It merges into
// the existing "mcp" section, preserving all other keys.
//
// The target directory is resolved as follows:
//  1. If $OPENCODE_CONFIG_DIR is set, it is used verbatim.
//  2. Otherwise, <homeRoot>/.config/opencode is used.
//
// A backup is created before writing, and the write is atomic.
// This implements ADR 0021 (Safe OpenCode Auto-Configuration).
//
// The allowReal parameter must be true to write to the user's real
// ~/.config/opencode/ directory. When false, the function only writes
// when $OPENCODE_CONFIG_DIR is set (sandbox mode). This is defense-in-depth:
// even if callers forget their own AllowReal guard, SetupMCPConfig itself
// refuses to touch the real OpenCode config.
func SetupMCPConfig(homeRoot string, binDir string, allowReal bool) (backupPath string, err error) {
	if !allowReal && os.Getenv("OPENCODE_CONFIG_DIR") == "" {
		if u, uErr := user.Current(); uErr == nil {
			realOCDir := filepath.Join(u.HomeDir, ".config", "opencode")
			ocDir := opencodeConfigDir(homeRoot)
			if ocDir == realOCDir || strings.HasPrefix(ocDir, realOCDir) {
				return "", fmt.Errorf("refusing to write to real OpenCode config at %s without OPENCODE_CONFIG_DIR; set OPENCODE_CONFIG_DIR for sandbox use or pass allowReal=true", ocDir)
			}
		}
	}

	configDir := opencodeConfigDir(homeRoot)

	// Write to opencode.json (the file OpenCode reads at startup).
	// config.json is a legacy/deprecated path that OpenCode 1.16+
	// rejects as "Unrecognized key: mcpServers".
	configPath := filepath.Join(configDir, "opencode.json")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("mkdir opencode config dir: %w", err)
	}

	cfg := readExistingConfigJSON(configPath)

	mcp, ok := cfg["mcp"].(map[string]any)
	if !ok || mcp == nil {
		mcp = map[string]any{}
	}
	mcp["plan-ai"] = map[string]any{
		"type":    "local",
		"enabled": true,
		"command": config.MCPCommand(binDir),
		"env": map[string]string{
			"PLAN_AI_PROJECT_ROOT": os.Getenv("PLAN_AI_PROJECT_ROOT"),
		},
	}
	cfg["mcp"] = mcp

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal opencode config: %w", err)
	}

	backupPath, err = atomicfile.WriteFileWithBackup(configPath, data, 0644, "pre-mcp-write")
	if err != nil {
		return backupPath, fmt.Errorf("write opencode config: %w", err)
	}

	return backupPath, nil
}

// readExistingConfigJSON reads and parses an existing config.json, or
// returns an empty map if the file doesn't exist or is unparseable.
func readExistingConfigJSON(path string) map[string]any {
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{}
	}
	var cfg map[string]any
	if err := json.Unmarshal(stripJSONCComments(data), &cfg); err != nil {
		return map[string]any{}
	}
	return cfg
}

// opencodeConfigDir resolves the OpenCode config directory from
// $OPENCODE_CONFIG_DIR (if set) or falls back to <homeRoot>/.config/opencode.
func opencodeConfigDir(homeRoot string) string {
	if d := os.Getenv("OPENCODE_CONFIG_DIR"); d != "" {
		return d
	}
	return filepath.Join(homeRoot, ".config", "opencode")
}

// DetectOpenCodeConfig is a convenience wrapper that uses the existing Detector
// to check for OpenCode config at the given opencode config dir.
func DetectOpenCodeConfig(opencodeDir string) *DetectionResult {
	// We search opencodeDir (which is the sandbox path) for config files.
	loc := filepath.Join(opencodeDir, "opencode.json")
	if info, err := os.Stat(loc); err == nil && !info.IsDir() {
		return &DetectionResult{Found: true, ConfigPath: loc, IsInitialized: true, AgentName: "plan-ai"}
	}
	loc = filepath.Join(opencodeDir, "opencode.jsonc")
	if info, err := os.Stat(loc); err == nil && !info.IsDir() {
		return &DetectionResult{Found: true, ConfigPath: loc, IsInitialized: true, AgentName: "plan-ai"}
	}
	loc = filepath.Join(opencodeDir, ".opencode", "opencode.json")
	if info, err := os.Stat(loc); err == nil && !info.IsDir() {
		return &DetectionResult{Found: true, ConfigPath: loc, IsInitialized: true, AgentName: "plan-ai"}
	}
	loc = filepath.Join(opencodeDir, ".opencode", "opencode.jsonc")
	if info, err := os.Stat(loc); err == nil && !info.IsDir() {
		return &DetectionResult{Found: true, ConfigPath: loc, IsInitialized: true, AgentName: "plan-ai"}
	}
	return &DetectionResult{Found: false}
}
