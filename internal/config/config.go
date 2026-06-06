package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type GlobalConfig struct {
	Version      string         `json:"version"`
	InstalledAt  string         `json:"installed_at"`
	GlobalDir    string         `json:"global_dir"`
	GlobalDB     string         `json:"global_db"`
	Integrations map[string]any `json:"integrations"`
}

type ProjectConfig struct {
	Version      string         `json:"version"`
	ProjectName  string         `json:"project_name"`
	ProjectRoot  string         `json:"project_root"`
	ProjectDB    string         `json:"project_db"`
	Mode         string         `json:"mode"`
	CreatedAt    string         `json:"created_at"`
	Integrations map[string]any `json:"integrations"`
}

// ProjectModeExternal stores project data under the global Plan-AI home
// (default). ProjectModeLocal stores project data inside the project's
// working tree at <root>/.plan-ai (opt-in via `plan-ai init --local`).
const (
	ProjectModeExternal = "external"
	ProjectModeLocal    = "local"
)

// ProjectSlug converts an absolute project root into a filesystem-safe slug
// used for the directory name under ~/.plan-ai/projects/. The slug is stable
// for any given root path so the same project always resolves to the same
// external directory.
func ProjectSlug(rootPath string) string {
	cleaned := filepath.Clean(rootPath)
	if cleaned == string(filepath.Separator) || cleaned == "." {
		return "project_root"
	}
	slug := strings.ReplaceAll(cleaned, string(filepath.Separator), "__")
	slug = strings.ReplaceAll(slug, ":", "_")
	for strings.Contains(slug, "__") {
		slug = strings.ReplaceAll(slug, "__", "_")
	}
	slug = strings.Trim(slug, "_")
	if slug == "" {
		return "project_root"
	}
	return "project_" + slug
}

func MCPRegistryPath(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "mcp-registry.json")
}

func GlobalDir(homeDir string) string {
	return filepath.Join(homeDir, ".plan-ai")
}

func GlobalConfigPath(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "config.json")
}

func GlobalDBPath(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "global.db")
}

func GlobalCacheDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "cache")
}

func GlobalLogsDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "logs")
}

func GlobalSkillsDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "skills")
}

func GlobalDataDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "data")
}

func GlobalBackupsDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "backups")
}

func ProjectDir(projectDir string) string {
	return filepath.Join(projectDir, ".plan-ai")
}

func ProjectConfigPath(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "config.json")
}

func ProjectDBPath(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "project.db")
}

func ProjectCacheDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "cache")
}

func ProjectSnapshotsDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "snapshots")
}

func ProjectExportsDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "exports")
}

func ProjectDocsDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "docs")
}

func ProjectLocksDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "locks")
}

func ProjectBackupsDir(projectDir string) string {
	return filepath.Join(ProjectDir(projectDir), "backups")
}

// GlobalProjectsDir is the parent directory under the global Plan-AI home
// that hosts one subdirectory per registered external project.
func GlobalProjectsDir(homeDir string) string {
	return filepath.Join(GlobalDir(homeDir), "projects")
}

// ExternalProjectDir returns the external per-project directory for the
// project identified by projectSlug (the result of ProjectSlug(rootPath)).
func ExternalProjectDir(homeDir, projectSlug string) string {
	return filepath.Join(GlobalProjectsDir(homeDir), projectSlug)
}

// ExternalProjectDBPath returns the SQLite path for an external project.
func ExternalProjectDBPath(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "project.db")
}

// ExternalProjectConfigPath returns the per-project config.json path for an
// external project.
func ExternalProjectConfigPath(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "config.json")
}

// ExternalProjectCacheDir returns the cache directory for an external project.
func ExternalProjectCacheDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "cache")
}

// ExternalProjectSnapshotsDir returns the snapshots directory for an external project.
func ExternalProjectSnapshotsDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "snapshots")
}

// ExternalProjectExportsDir returns the exports directory for an external project.
func ExternalProjectExportsDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "exports")
}

// ExternalProjectDocsDir returns the docs directory for an external project.
func ExternalProjectDocsDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "docs")
}

// ExternalProjectLocksDir returns the locks directory for an external project.
func ExternalProjectLocksDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "locks")
}

// ExternalProjectBackupsDir returns the backups directory for an external project.
func ExternalProjectBackupsDir(homeDir, projectSlug string) string {
	return filepath.Join(ExternalProjectDir(homeDir, projectSlug), "backups")
}

func LoadGlobalConfig(path string) (GlobalConfig, error) {
	var cfg GlobalConfig
	err := loadJSON(path, &cfg)
	return cfg, err
}

func SaveGlobalConfig(path string, cfg GlobalConfig) error {
	return saveJSON(path, cfg)
}

func LoadProjectConfig(path string) (ProjectConfig, error) {
	var cfg ProjectConfig
	err := loadJSON(path, &cfg)
	return cfg, err
}

func SaveProjectConfig(path string, cfg ProjectConfig) error {
	return saveJSON(path, cfg)
}

func loadJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func saveJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}
