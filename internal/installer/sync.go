package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/plan-ai/plan-ai/internal/opencode"
)

// syncOpenCodeConfig generates or merges Plan-AI artifacts into the
// OpenCode config directory. It uses SetupService for the heavy lifting.
func syncOpenCodeConfig(ocDir, binDir string) error {
	// Determine project root from cwd or env
	projectRoot := os.Getenv("PLAN_AI_PROJECT_ROOT")
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		projectRoot = cwd
	}

	svc := opencode.NewSetupService()
	result, err := svc.Run(ocDir, projectRoot)
	if err != nil {
		return fmt.Errorf("setup service: %w", err)
	}

	// Update the MCP command path if binDir is specified
	if binDir != "" && result.MCPRegistryPath != "" {
		if err := updateMCPCommandPath(result.MCPRegistryPath, binDir); err != nil {
			return fmt.Errorf("update mcp command path: %w", err)
		}
	}

	_ = result // artifacts are on disk
	return nil
}

// removePlanAIFromOpenCodeConfig removes Plan-AI entries from the OpenCode
// config file while preserving all other entries.
func removePlanAIFromOpenCodeConfig(ocDir string) error {
	candidates := []string{
		filepath.Join(ocDir, "opencode.json"),
		filepath.Join(ocDir, "opencode.jsonc"),
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read %s: %w", path, err)
		}

		var cfg map[string]any
		if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
			continue // skip unparseable
		}

		// Remove plan-ai from MCP
		if mcpRaw, ok := cfg["mcp"].(map[string]any); ok {
			delete(mcpRaw, "plan-ai")
			if len(mcpRaw) == 0 {
				delete(cfg, "mcp")
			} else {
				cfg["mcp"] = mcpRaw
			}
		}

		// Remove plan-ai from agents if present
		if agentsRaw, ok := cfg["agents"].(map[string]any); ok {
			delete(agentsRaw, "plan-ai")
			if len(agentsRaw) == 0 {
				delete(cfg, "agents")
			} else {
				cfg["agents"] = agentsRaw
			}
		}

		// Remove plan-ai skills
		if skills, ok := cfg["skills"].([]any); ok {
			filtered := make([]any, 0, len(skills))
			for _, s := range skills {
				if sStr, ok := s.(string); ok && sStr == "plan-ai" {
					continue
				}
				filtered = append(filtered, s)
			}
			if len(filtered) == 0 {
				delete(cfg, "skills")
			} else {
				cfg["skills"] = filtered
			}
		}

		// Remove agent plan-ai if agent_name is plan-ai
		if name, ok := cfg["agent_name"].(string); ok && name == "plan-ai" {
			delete(cfg, "agent_name")
			delete(cfg, "agent_role")
			delete(cfg, "agent")
		}

		// Write back only if we actually removed something
		if len(cfg) > 0 {
			out, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal config: %w", err)
			}
			if err := os.WriteFile(path, out, 0644); err != nil {
				return fmt.Errorf("write %s: %w", path, err)
			}
		}
	}

	return nil
}

// updateMCPCommandPath updates the command path in the MCP registry file
// to point to the correct binDir.
func updateMCPCommandPath(registryPath, binDir string) error {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return fmt.Errorf("read registry: %w", err)
	}

	var registry map[string]any
	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("parse registry: %w", err)
	}

	// Update command to use full path
	mcpBin := filepath.Join(binDir, binNameMCP)
	registry["command"] = []string{mcpBin}

	out, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal registry: %w", err)
	}
	return os.WriteFile(registryPath, out, 0644)
}

// cleanJSONCComments strips // and /* */ comments from JSONC data.
func cleanJSONCComments(data []byte) []byte {
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


