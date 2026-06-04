package opencode

// DetectionResult describes what was found about a project's OpenCode setup.
type DetectionResult struct {
	Found         bool   `json:"found"`                 // Is OpenCode config present?
	ConfigPath    string `json:"config_path,omitempty"` // Path to opencode.json[c]
	IsInitialized bool   `json:"is_initialized"`        // Has opencode init been run?
	AgentName     string `json:"agent_name,omitempty"`  // Configured agent name
	AgentRole     string `json:"agent_role,omitempty"`  // Agent role description
	HasSkills     bool   `json:"has_skills"`            // Has configured skills
	SkillCount    int    `json:"skill_count"`           // Number of installed skills
	HasSelfInit   bool   `json:"has_self_init"`         // Has self-init configuration
	Error         string `json:"error,omitempty"`       // Detection error message
}

// IntegrationMode describes how Plan-AI integrates with OpenCode.
type IntegrationMode string

const (
	ModeStandalone IntegrationMode = "standalone" // Plan-AI runs independently
	ModeTool       IntegrationMode = "tool"       // Plan-AI exposes tools for OpenCode MCP
	ModeHybrid     IntegrationMode = "hybrid"     // Both modes active
)

// Config holds the OpenCode integration configuration.
type Config struct {
	Enabled        bool            `json:"enabled"`
	Mode           IntegrationMode `json:"mode"`
	AutoDetect     bool            `json:"auto_detect"`      // Auto-detect opencode config
	WarnOnConflict bool            `json:"warn_on_conflict"` // Warn if Plan-AI config conflicts with OC
	ReadOnly       bool            `json:"read_only"`        // Read-only OC integration
	DoctorChecks   []string        `json:"doctor_checks"`    // Enabled doctor checks
}
