package opencode

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateProjectArtifacts creates the directory structure for OpenCode project
// foundation artifacts. Used when --opencode is set on plan-ai init.
func GenerateProjectArtifacts(homeRoot string) error {
	base := filepath.Join(homeRoot, ".config", "opencode")

	dirs := []string{
		filepath.Join(base, "profiles"),
		filepath.Join(base, "prompts"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	return nil
}
