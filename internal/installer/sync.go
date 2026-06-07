package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Durru/plan-ai/internal/config"
	"github.com/Durru/plan-ai/internal/opencode"
)

// opencodeSchemaURL is the schema URL for OpenCode config files.
const opencodeSchemaURL = "https://opencode.ai/config.json"

// invalidOpenCodeKeys are keys that MUST NOT appear in the final opencode.json.
// IMPORTANT: only strip keys that Plan-AI owns. OpenCode's own config keys
// (providers, provider.list, app.agents, config.get) are legitimate and must
// be preserved.
var invalidOpenCodeKeys = map[string]bool{}

// syncOpenCodeConfig is a thin wrapper that delegates to opencode.SetupMCPConfig.
func syncOpenCodeConfig(homeRoot, binDir string, allowReal bool) error {
	_, err := opencode.SetupMCPConfig(homeRoot, binDir, allowReal)
	return err
}

// generateOpenCodeConfigContent builds the opencode.json content in memory.
// It reads and merges with any existing config, strips invalid keys,
// and returns the final JSON bytes without writing to disk.
func generateOpenCodeConfigContent(ocDir, binDir string) ([]byte, error) {
	configPath := filepath.Join(ocDir, "opencode.json")

	// Determine project root for env
	projectRoot := os.Getenv("PLAN_AI_PROJECT_ROOT")
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("getwd: %w", err)
		}
		projectRoot = cwd
	}

	mcpCmd := planAIMCPCommand(binDir)

	// Read existing config if present
	var cfg map[string]any
	configExists := false
	if data, err := os.ReadFile(configPath); err == nil {
		configExists = true
		cleaned := cleanJSONCComments(data)
		if err := json.Unmarshal(cleaned, &cfg); err != nil {
			// Unparseable config — back it up and start fresh
			backupPath := configPath + ".invalid." + timeNowFilename()
			if err := os.WriteFile(backupPath, data, 0644); err == nil {
				fmt.Fprintf(os.Stderr, "backed up invalid opencode.json to %s\n", backupPath)
			}
			cfg = make(map[string]any)
		}
	} else {
		cfg = make(map[string]any)
	}

	didStrip := false
	for key := range invalidOpenCodeKeys {
		if _, exists := cfg[key]; exists {
			delete(cfg, key)
			didStrip = true
		}
	}

	if didStrip && configExists {
		backupPath := configPath + ".stripped." + timeNowFilename()
		data, _ := os.ReadFile(configPath)
		if err := os.WriteFile(backupPath, data, 0644); err == nil {
			fmt.Fprintf(os.Stderr, "backed up original config to %s (stripped invalid keys)\n", backupPath)
		}
	}

	if _, ok := cfg["$schema"]; !ok {
		cfg["$schema"] = opencodeSchemaURL
	}

	mcpRaw, ok := cfg["mcp"]
	if !ok || mcpRaw == nil {
		mcpRaw = make(map[string]any)
		cfg["mcp"] = mcpRaw
	}
	mcpSection, ok := mcpRaw.(map[string]any)
	if !ok {
		mcpSection = make(map[string]any)
		cfg["mcp"] = mcpSection
	}

	mcpSection["plan-ai"] = map[string]any{
		"type":    "local",
		"enabled": true,
		"command": mcpCmd,
		"env": map[string]string{
			"PLAN_AI_PROJECT_ROOT": projectRoot,
		},
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}
	return out, nil
}

func planAIMCPCommand(binDir string) []string {
	return config.MCPCommand(binDir)
}

// removePlanAIFromOpenCodeConfig removes the Plan-AI MCP entry from the
// OpenCode config while preserving all other entries.
func removePlanAIFromOpenCodeConfig(ocDir string) error {
	candidates := []string{
		filepath.Join(ocDir, "opencode.json"),
		filepath.Join(ocDir, "opencode.jsonc"),
		filepath.Join(ocDir, "config.json"),
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

		// Remove plan-ai from MCP (legacy format)
		if mcpRaw, ok := cfg["mcp"].(map[string]any); ok {
			delete(mcpRaw, "plan-ai")
			if len(mcpRaw) == 0 {
				delete(cfg, "mcp")
			} else {
				cfg["mcp"] = mcpRaw
			}
		}

		// Remove plan-ai from MCP (new format)
		if mcpRaw, ok := cfg["mcpServers"].(map[string]any); ok {
			delete(mcpRaw, "plan-ai")
			if len(mcpRaw) == 0 {
				delete(cfg, "mcpServers")
			} else {
				cfg["mcpServers"] = mcpRaw
			}
		}

		// Write back
		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}
		if err := writeFileAtomically(path, out, 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}

	return nil
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

// timeNowFilename returns a compact timestamp for backup filenames.
func timeNowFilename() string {
	return timeNowUTC()
}

// hasMCPPlanAI checks whether the opencode config at path contains mcp.plan-ai.
func hasMCPPlanAI(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
		return false
	}
	mcpRaw, ok := cfg["mcp"]
	if !ok {
		return false
	}
	mcpSection, ok := mcpRaw.(map[string]any)
	if !ok {
		return false
	}
	_, ok = mcpSection["plan-ai"]
	return ok
}

// countPlanAIEntries returns the number of plan-ai* MCP entries in the
// opencode config (old format, new format). Used by Doctor to detect
// duplicate registrations left behind by failed updates.
func countPlanAIEntries(path string) (oldCount, newCount int, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	var cfg map[string]any
	if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
		return 0, 0, err
	}
	if mcpRaw, ok := cfg["mcp"].(map[string]any); ok {
		oldCount = countKeysWithPrefix(mcpRaw, "plan-ai")
	}
	if mcpRaw, ok := cfg["mcpServers"].(map[string]any); ok {
		newCount = countKeysWithPrefix(mcpRaw, "plan-ai")
	}
	return oldCount, newCount, nil
}

// countKeysWithPrefix returns the number of map keys that start with prefix.
func countKeysWithPrefix(m map[string]any, prefix string) int {
	n := 0
	for k := range m {
		if strings.HasPrefix(k, prefix) {
			n++
		}
	}
	return n
}

// hasMCPPlanAINewFormat checks whether the config at path contains mcpServers.plan-ai.
func hasMCPPlanAINewFormat(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
		return false
	}
	mcpRaw, ok := cfg["mcpServers"]
	if !ok {
		return false
	}
	mcpSection, ok := mcpRaw.(map[string]any)
	if !ok {
		return false
	}
	_, ok = mcpSection["plan-ai"]
	return ok
}

// hasSchema checks whether the opencode config at path has a $schema field.
func hasSchema(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var cfg map[string]any
	if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
		return false
	}
	_, ok := cfg["$schema"]
	return ok
}

// hasInvalidKeys checks whether the opencode config at path contains any
// keys listed in invalidOpenCodeKeys.
func hasInvalidKeys(path string) (bool, []string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(cleanJSONCComments(data), &cfg); err != nil {
		return false, nil
	}
	var found []string
	for key := range invalidOpenCodeKeys {
		if _, exists := cfg[key]; exists {
			found = append(found, key)
		}
	}
	return len(found) > 0, found
}

// openCodeConfigPath returns the path to opencode.json in the config dir.
func openCodeConfigPath(ocDir string) string {
	// Prefer .json over .jsonc
	jsonPath := filepath.Join(ocDir, "opencode.json")
	if _, err := os.Stat(jsonPath); err == nil {
		return jsonPath
	}
	jsoncPath := filepath.Join(ocDir, "opencode.jsonc")
	if _, err := os.Stat(jsoncPath); err == nil {
		return jsoncPath
	}
	return jsonPath // default even if not existent
}
