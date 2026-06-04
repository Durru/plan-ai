package opencode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "opencode-integration.json"

// DefaultConfig returns the default OpenCode integration configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:        true,
		Mode:           ModeTool,
		AutoDetect:     true,
		WarnOnConflict: true,
		ReadOnly:       true,
		DoctorChecks:   []string{"version", "config", "mcp"},
	}
}

// LoadConfig loads the integration config from the project's .opencode/ directory.
func LoadConfig(homeRoot string) (Config, error) {
	cfg := DefaultConfig()
	path := filepath.Join(homeRoot, configFileName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // defaults
		}
		return cfg, fmt.Errorf("read integration config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse integration config: %w", err)
	}

	return cfg, nil
}

// SaveConfig writes the integration config to the project's .opencode/ directory.
func SaveConfig(homeRoot string, cfg Config) error {
	path := filepath.Join(homeRoot, configFileName)

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
