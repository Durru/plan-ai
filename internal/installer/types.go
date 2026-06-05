// Package installer provides a Gentle-AI style installer for Plan-AI.
//
// It manages global state (~/.plan-ai/state.json), component-based
// install/sync/uninstall, preset selection, tool detection, backups,
// and OpenCode config integration — all testable with a temp HOME.
package installer

// Component names used throughout the installer.
const (
	CompIntent    = "intent"
	CompPlanning  = "planning"
	CompMCP       = "mcp"
	CompOpenCode  = "opencode-agent"
	CompDocs      = "docs"
	CompContext   = "context"
	CompAlignment = "alignment"
)

// AllComponents is the canonical list of all installable components.
var AllComponents = []string{CompIntent, CompPlanning, CompMCP, CompOpenCode, CompDocs, CompContext, CompAlignment}

// Preset definitions map presets to their component sets.
var Presets = map[string][]string{
	"full-plan-ai":   AllComponents,
	"ecosystem-only": {CompMCP, CompOpenCode, CompDocs},
	"minimal":        {CompMCP},
}

// ComponentDescriptions describes each component.
var ComponentDescriptions = map[string]string{
	CompIntent:    "Product Intent, discovery, and ambiguity analysis",
	CompPlanning:  "Master plans, specific plans, tasks, phases, and workflows",
	CompMCP:       "MCP server binary, protocol, tools, and tool registry",
	CompOpenCode:  "OpenCode agent registration, profiles, prompts, and workflows",
	CompDocs:      "Installation docs, quickstart, and OpenCode integration guides",
	CompContext:   "L0-L4 context generation and approved context management",
	CompAlignment: "Alignment checks, validation rules, and confidence scoring",
}

// State tracks what is installed globally.
type State struct {
	Version     string                    `json:"version"`
	InstalledAt string                    `json:"installed_at"`
	UpdatedAt   string                    `json:"updated_at"`
	Components  map[string]ComponentState `json:"components"`
	Preset      string                    `json:"preset"`
	BinDir      string                    `json:"bin_dir"`
	DataDir     string                    `json:"data_dir"`
	Tools       ToolsDetected             `json:"tools_detected"`
}

// ComponentState tracks whether a single component is installed.
type ComponentState struct {
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
}

// ToolsDetected records what tools were found on the system.
type ToolsDetected struct {
	OpenCode  bool `json:"opencode"`
	Git       bool `json:"git"`
	Go        bool `json:"go"`
	MCPBinary bool `json:"plan_ai_mcp_server"`
}

// InstallOptions controls install, sync, and init behaviour.
type InstallOptions struct {
	DryRun     bool
	Preset     string
	Components []string // only used when Preset == "custom"
	BinDir     string   // where to install binaries (defaults to $HOME/.local/bin)
	AllowReal  bool     // allow writing to real ~/.config/opencode
}

// DoctorReport is the result of a doctor check.
type DoctorReport struct {
	StateExists         bool
	StateValid          bool
	Tools               ToolsDetected
	ComponentsInstalled int
	ComponentsTotal     int
	OpenCodeValid       bool
	OpenCodeConfigPath  string
	DataDir             string
	BinDir              string
	Preset              string
	GlobalDBExists      bool
	ProjectDBExists     bool
}
