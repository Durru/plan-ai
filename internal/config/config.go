package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	CreatedAt    string         `json:"created_at"`
	Integrations map[string]any `json:"integrations"`
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
