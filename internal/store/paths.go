package store

import (
	"os"
	"path/filepath"

	"github.com/Durru/plan-ai/internal/config"
)

// ResolveHomeRoot preserves the current PLAN_AI_HOME contract: the value is
// the home root and .plan-ai is appended by config.GlobalDir.
func ResolveHomeRoot() (string, error) {
	if home := os.Getenv("PLAN_AI_HOME"); home != "" {
		return filepath.Abs(home)
	}
	return os.UserHomeDir()
}

func ResolveProjectRoot() (string, error) {
	if root := os.Getenv("PLAN_AI_PROJECT_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	return os.Getwd()
}

func GlobalStorePath(homeRoot string) string { return config.GlobalDBPath(homeRoot) }

func ProjectStorePath(projectRoot string) string { return config.ProjectDBPath(projectRoot) }
