package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Detector discovers OpenCode configuration in a project directory.
type Detector struct{}

// NewDetector creates a new OpenCode detector.
func NewDetector() *Detector {
	return &Detector{}
}

// Detect checks for OpenCode configuration in the given project root.
// It searches for opencode.json or opencode.jsonc in the project root,
// the parent directory, or the project's .opencode/ subdirectory.
func (d *Detector) Detect(projectRoot string) *DetectionResult {
	result := &DetectionResult{Found: false}

	// Sanitize
	projectRoot = filepath.Clean(projectRoot)

	// Search locations in order of priority
	locations := []string{
		filepath.Join(projectRoot, "opencode.json"),
		filepath.Join(projectRoot, "opencode.jsonc"),
		filepath.Join(projectRoot, ".opencode", "opencode.json"),
		filepath.Join(projectRoot, ".opencode", "opencode.jsonc"),
	}

	// Also check parent directory for workspace-level config
	parent := filepath.Dir(projectRoot)
	if parent != projectRoot {
		locations = append(locations,
			filepath.Join(parent, "opencode.json"),
			filepath.Join(parent, "opencode.jsonc"),
		)
	}

	for _, path := range locations {
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			result.Found = true
			result.ConfigPath = path
			d.parseConfig(path, result)
			return result
		}
	}

	return result
}

// parseConfig reads and parses an OpenCode config file, populating the result.
func (d *Detector) parseConfig(path string, result *DetectionResult) {
	data, err := os.ReadFile(path)
	if err != nil {
		result.Error = fmt.Sprintf("read config: %v", err)
		return
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		result.Error = fmt.Sprintf("parse config: %v", err)
		return
	}

	// Check agent name
	if agentName, ok := raw["agent_name"].(string); ok {
		result.AgentName = agentName
		result.IsInitialized = true
	}

	// Check agent role
	if agentRole, ok := raw["agent_role"].(string); ok {
		result.AgentRole = agentRole
	}

	// Check skills
	if skills, ok := raw["skills"]; ok {
		result.HasSkills = true
		if arr, ok := skills.([]any); ok {
			result.SkillCount = len(arr)
		}
	}

	// Check self-init (agent block indicates self-init capability)
	if _, ok := raw["agent"]; ok {
		result.HasSelfInit = true
	}
}
