package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// opencodeSchemaURL is the schema URL for OpenCode config files.
const opencodeSchemaURL = "https://opencode.ai/config.json"

// invalidOpenCodeKeys are keys that MUST NOT appear in the final opencode.json.
// If present in an existing config, they are stripped and the original is backed up.
var invalidOpenCodeKeys = map[string]bool{
	"providers":     true,
	"provider.list": true,
	"app.agents":    true,
}

// syncOpenCodeConfig generates or merges the Plan-AI MCP entry into the
// OpenCode config file (opencode.json). It writes directly to opencode.json
// under the mcp.plan-ai key and does NOT depend on mcp-registry.json.
//
// Strategy:
//  1. Read existing opencode.json (handle JSONC comments)
//  2. Strip invalid keys (providers, provider.list, app.agents)
//  3. Ensure $schema is present
//  4. Merge mcp.plan-ai with the server configuration
//  5. Write back as clean JSON
func syncOpenCodeConfig(ocDir, binDir string) error {
	configPath := filepath.Join(ocDir, "opencode.json")
	if err := os.MkdirAll(ocDir, 0755); err != nil {
		return fmt.Errorf("mkdir opencode dir: %w", err)
	}

	// Determine project root for env
	projectRoot := os.Getenv("PLAN_AI_PROJECT_ROOT")
	if projectRoot == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd: %w", err)
		}
		projectRoot = cwd
	}

	// Determine MCP command
	mcpCmd := "plan-ai-mcp-server"
	if binDir != "" {
		mcpCmd = filepath.Join(binDir, binNameMCP)
	}

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
	// Strip invalid keys
	for key := range invalidOpenCodeKeys {
		if _, exists := cfg[key]; exists {
			delete(cfg, key)
			didStrip = true
		}
	}

	// If we stripped invalid keys from an existing config, back up the original
	if didStrip && configExists {
		backupPath := configPath + ".stripped." + timeNowFilename()
		data, _ := os.ReadFile(configPath)
		if err := os.WriteFile(backupPath, data, 0644); err == nil {
			fmt.Fprintf(os.Stderr, "backed up original config to %s (stripped invalid keys)\n", backupPath)
		}
	}

	// Ensure $schema
	if _, ok := cfg["$schema"]; !ok {
		cfg["$schema"] = opencodeSchemaURL
	}

	// Ensure mcp section exists
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

	// Set mcp.plan-ai
	mcpSection["plan-ai"] = map[string]any{
		"type":    "local",
		"enabled": true,
		"command": []string{mcpCmd},
		"env": map[string]string{
			"PLAN_AI_PROJECT_ROOT": projectRoot,
		},
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(configPath, out, 0644)
}

// removePlanAIFromOpenCodeConfig removes the Plan-AI MCP entry from the
// OpenCode config while preserving all other entries.
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

		// Write back
		out, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}
		if err := os.WriteFile(path, out, 0644); err != nil {
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
