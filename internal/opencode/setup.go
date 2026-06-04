package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
}

// setupMCPItem mirrors RegistryItem for the setup registry artifact.
type setupMCPItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	PlansBuilt  int    `json:"plans_built"`
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
	if err := os.WriteFile(path, data, 0644); err != nil {
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
	}

	configPath := filepath.Join(opencodeDir, "opencode.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return "", fmt.Errorf("mkdir opencode config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal opencode config: %w", err)
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return "", fmt.Errorf("write opencode config: %w", err)
	}

	return configPath, nil
}

// writeMCPRegistry writes a Plan-AI MCP tool registry artifact.
func (s *SetupService) writeMCPRegistry(opencodeDir string) (string, error) {
	items := []setupMCPItem{
		{Name: "plan_ai.status", Description: "Get Plan-AI status and domain counts", PlansBuilt: 0},
		{Name: "plan_ai.scan", Description: "Scan the project for structure and dependencies", PlansBuilt: 0},
		{Name: "plan_ai.plan", Description: "Create planning artifacts from vision and approved context", PlansBuilt: 0},
		{Name: "plan_ai.research", Description: "Manage research entries and findings", PlansBuilt: 0},
		{Name: "plan_ai.knowledge", Description: "Query reusable knowledge base", PlansBuilt: 0},
	}

	path := filepath.Join(opencodeDir, "mcp-registry.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("mkdir mcp registry dir: %w", err)
	}

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal mcp registry: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("write mcp registry: %w", err)
	}

	return path, nil
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
	if err := os.WriteFile(path, data, 0644); err != nil {
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
	if err := os.WriteFile(path, data, 0644); err != nil {
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
	if err := os.WriteFile(path, data, 0644); err != nil {
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
