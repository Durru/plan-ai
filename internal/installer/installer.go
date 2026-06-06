package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/plan-ai/plan-ai/internal/config"
)

// stateVersion is the current schema version for state.json.
const stateVersion = "1"

// binNameMCP is the CLI binary that serves MCP through `plan-ai mcp serve`.
const binNameMCP = config.BinNameMCP

// backupTimeFormat is used for backup file naming.
const backupTimeFormat = "20060102-150405"

// Installer manages the global Plan-AI installation lifecycle.
type Installer struct {
	HomeDir string // effective $HOME
	DataDir string // ~/.plan-ai
	State   *State // parsed state.json (nil until LoadState or Install)
}

// NewInstaller creates an installer rooted at homeDir.
func NewInstaller(homeDir string) *Installer {
	return &Installer{
		HomeDir: homeDir,
		DataDir: filepath.Join(homeDir, ".plan-ai"),
	}
}

// ── state management ────────────────────────────────────

// LoadState reads state.json from DataDir into State.
func (inst *Installer) LoadState() error {
	path := filepath.Join(inst.DataDir, "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("parse state: %w", err)
	}
	inst.State = &s
	return nil
}

// SaveState persists State to state.json.
func (inst *Installer) SaveState() error {
	if inst.State == nil {
		return fmt.Errorf("state is nil")
	}
	path := filepath.Join(inst.DataDir, "state.json")
	if err := os.MkdirAll(inst.DataDir, 0755); err != nil {
		return fmt.Errorf("mkdir data dir: %w", err)
	}
	data, err := json.MarshalIndent(inst.State, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return writeFileAtomically(path, data, 0644)
}

// ── tool detection ──────────────────────────────────────

// DetectTools scans PATH and common locations for required tools.
func (inst *Installer) DetectTools() ToolsDetected {
	t := ToolsDetected{}

	// Check PATH for executables
	pathDirs := filepath.SplitList(os.Getenv("PATH"))
	// Add common locations
	pathDirs = append(pathDirs,
		filepath.Join(inst.HomeDir, ".local", "bin"),
		"/usr/local/bin",
		"/usr/bin",
	)

	seen := map[string]bool{}
	for _, dir := range pathDirs {
		cleaned := filepath.Clean(dir)
		if seen[cleaned] {
			continue
		}
		seen[cleaned] = true

		entries, err := os.ReadDir(cleaned)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			info, err := e.Info()
			if err != nil {
				continue
			}
			// Check it's executable
			if info.Mode()&0111 == 0 {
				continue
			}
			switch name {
			case "opencode":
				t.OpenCode = true
			case "git":
				t.Git = true
			case "go":
				t.Go = true
			case binNameMCP:
				t.MCPBinary = true
			}
		}
	}

	// Also detect OpenCode via config dir
	if !t.OpenCode {
		if ocDir := openCodeConfigDir(inst.HomeDir); ocDir != "" {
			candidates := []string{
				filepath.Join(ocDir, "opencode.json"),
				filepath.Join(ocDir, "opencode.jsonc"),
			}
			for _, c := range candidates {
				if _, err := os.Stat(c); err == nil {
					t.OpenCode = true
					break
				}
			}
		}
	}

	return t
}

// ── install ─────────────────────────────────────────────

// Install performs a global Plan-AI installation.
// It creates state.json, detects tools, generates OpenCode integration
// artifacts (with backup), and records the selected preset.
func (inst *Installer) Install(opts InstallOptions) error {
	if opts.DryRun {
		return nil
	}

	// Resolve preset components
	components := resolveComponents(opts)

	// Detect tools
	tools := inst.DetectTools()

	// Build state
	now := timeNowUTC()
	compState := make(map[string]ComponentState, len(AllComponents))
	for _, c := range AllComponents {
		installed := false
		for _, selected := range components {
			if c == selected {
				installed = true
				break
			}
		}
		compState[c] = ComponentState{
			Installed: installed,
			Version:   stateVersion,
		}
	}

	inst.State = &State{
		Version:     stateVersion,
		InstalledAt: now,
		UpdatedAt:   now,
		Components:  compState,
		Preset:      opts.Preset,
		BinDir:      opts.BinDir,
		DataDir:     inst.DataDir,
		Tools:       tools,
	}

	// Backup existing OpenCode config before modifying
	if err := inst.backupOpenCodeConfig(); err != nil {
		return fmt.Errorf("backup opencode config: %w", err)
	}

	// Generate OpenCode integration artifacts if opencode-agent is selected
	if compState[CompOpenCode].Installed || compState[CompMCP].Installed {
		if err := inst.syncOpenCodeConfig(opts); err != nil {
			return fmt.Errorf("sync opencode config: %w", err)
		}
	}

	return inst.SaveState()
}

// ── sync ────────────────────────────────────────────────

// Sync re-applies the current installation state, ensuring all artifacts
// are present and up-to-date. It is idempotent.
func (inst *Installer) Sync(opts InstallOptions) error {
	if inst.State == nil {
		if err := inst.LoadState(); err != nil {
			return fmt.Errorf("cannot sync: no state found — run install first: %w", err)
		}
	}

	// Update tools
	inst.State.Tools = inst.DetectTools()
	inst.State.UpdatedAt = timeNowUTC()
	inst.State.DataDir = inst.DataDir

	// Re-apply OpenCode integration if applicable
	if inst.State.Components[CompOpenCode].Installed || inst.State.Components[CompMCP].Installed {
		if err := inst.backupOpenCodeConfig(); err != nil {
			return fmt.Errorf("backup opencode config: %w", err)
		}
		if err := inst.syncOpenCodeConfig(opts); err != nil {
			return fmt.Errorf("sync opencode config: %w", err)
		}
	}

	return inst.SaveState()
}

// SyncPlan describes what a Sync operation would change.
type SyncPlan struct {
	WillSync    bool     // true if there's something to sync
	FilesPlan  []string // planned file changes
	OpenCode   string   // "add" | "update" | "remove" | "none"
	OpenCodeNew string  // new opencode config content (empty if no change)
}

// PreviewSync generates a plan of what Sync would change, without writing anything.
// Returns nil plan (not error) if there's nothing to sync.
func (inst *Installer) PreviewSync() (*SyncPlan, error) {
	if inst.State == nil {
		if err := inst.LoadState(); err != nil {
			return nil, nil // nothing to sync
		}
	}

	plan := &SyncPlan{WillSync: false}

	needsOC := inst.State.Components[CompOpenCode].Installed || inst.State.Components[CompMCP].Installed
	if !needsOC {
		return plan, nil
	}

	ocDir := openCodeConfigDir(inst.HomeDir)
	if ocDir == "" {
		return plan, nil
	}

	currentPath := openCodeConfigPath(ocDir)
	plan.FilesPlan = append(plan.FilesPlan, currentPath)

	// Determine action: add, update, or none
	if _, err := os.Stat(currentPath); err == nil {
		plan.OpenCode = "update"
	} else {
		plan.OpenCode = "add"
	}

	// Generate the planned content in-memory
	binDir := ""
	if inst.State != nil {
		binDir = inst.State.BinDir
	}
	content, err := generateOpenCodeConfigContent(ocDir, binDir)
	if err != nil {
		return nil, fmt.Errorf("generate preview: %w", err)
	}
	plan.OpenCodeNew = string(content)
	plan.WillSync = true

	return plan, nil
}

// ── uninstall ───────────────────────────────────────────

// Uninstall removes components from state and optionally cleans up
// OpenCode integration artifacts. If components is nil or empty,
// it removes everything (full uninstall).
func (inst *Installer) Uninstall(components []string) error {
	if inst.State == nil {
		if err := inst.LoadState(); err != nil {
			return fmt.Errorf("nothing to uninstall — no state found")
		}
	}

	if len(components) == 0 {
		// Full uninstall — surgically remove Plan-AI's footprint:
		//   1. Strip the plan-ai MCP entry from the opencode config
		//      (with a backup so a user mistake is recoverable).
		//   2. Remove the global data dir (state, db, subdirs).
		if err := inst.backupOpenCodeConfig(); err != nil {
			return fmt.Errorf("backup before full uninstall: %w", err)
		}
		if err := inst.removePlanAIFromOpenCode(); err != nil {
			return fmt.Errorf("remove plan-ai from opencode: %w", err)
		}
		if err := os.RemoveAll(inst.DataDir); err != nil {
			return fmt.Errorf("remove data dir: %w", err)
		}
		inst.State = nil
		return nil
	}

	// Partial uninstall — mark components as not installed
	for _, c := range components {
		if cs, ok := inst.State.Components[c]; ok {
			cs.Installed = false
			inst.State.Components[c] = cs
		}
	}
	inst.State.UpdatedAt = timeNowUTC()

	// Clean up OpenCode config if removing mcp or opencode-agent
	removeOC := false
	for _, c := range components {
		if c == CompMCP || c == CompOpenCode {
			removeOC = true
			break
		}
	}
	if removeOC {
		if err := inst.backupOpenCodeConfig(); err != nil {
			return fmt.Errorf("backup before uninstall: %w", err)
		}
		if err := inst.removePlanAIFromOpenCode(); err != nil {
			return fmt.Errorf("remove plan-ai from opencode: %w", err)
		}
	}

	return inst.SaveState()
}

// ── doctor ──────────────────────────────────────────────

// Doctor runs health checks and returns a report.
func (inst *Installer) Doctor() *DoctorReport {
	r := &DoctorReport{
		DataDir: inst.DataDir,
		Tools:   inst.DetectTools(),
	}

	// Check state
	statePath := filepath.Join(inst.DataDir, "state.json")
	if _, err := os.Stat(statePath); err == nil {
		r.StateExists = true
	}

	if inst.State != nil || r.StateExists {
		if err := inst.LoadState(); err == nil {
			r.StateValid = true
			r.BinDir = inst.State.BinDir
			r.Preset = inst.State.Preset
			for _, cs := range inst.State.Components {
				r.ComponentsTotal++
				if cs.Installed {
					r.ComponentsInstalled++
				}
			}
		}
	}

	// Check OpenCode — must have mcp.plan-ai (old format) or mcpServers.plan-ai (new format)
	ocDir := openCodeConfigDir(inst.HomeDir)
	if ocDir != "" {
		candidates := []string{
			filepath.Join(ocDir, "opencode.json"),
			filepath.Join(ocDir, "opencode.jsonc"),
			filepath.Join(ocDir, "config.json"),
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				r.OpenCodeConfigPath = path
				if hasMCPPlanAI(path) || hasMCPPlanAINewFormat(path) {
					r.OpenCodeValid = true
					break
				}
			}
		}
		// If no candidate was valid, keep the path of the last existing one
		if !r.OpenCodeValid {
			for _, path := range candidates {
				if _, err := os.Stat(path); err == nil {
					r.OpenCodeConfigPath = path
					break
				}
			}
		}
	}

	// Check DB files
	globalDB := config.GlobalDBPath(inst.HomeDir)
	if _, err := os.Stat(globalDB); err == nil {
		r.GlobalDBExists = true
	}

	// Issue detection: missing registered binary, duplicate MCP entries,
	// stale config (state says installed but mcp-srv not in OpenCode).
	if inst.State != nil {
		if r.BinDir != "" {
			bin := filepath.Join(r.BinDir, config.BinNameMCP)
			if _, err := os.Stat(bin); err != nil {
				r.Issues = append(r.Issues, DoctorIssue{
					Severity: "warn",
					Code:     "registered_binary_missing",
					Message:  fmt.Sprintf("state records MCP binary at %s but it is not on disk", bin),
				})
			}
		}
	}

	if r.OpenCodeConfigPath != "" {
		// Duplicate detection: count plan-ai MCP entries in the opencode
		// config (handles both legacy `mcp.plan-ai` and new `mcpServers.plan-ai`).
		oldCount, newCount, _ := countPlanAIEntries(r.OpenCodeConfigPath)
		if oldCount+newCount > 1 {
			r.Issues = append(r.Issues, DoctorIssue{
				Severity: "warn",
				Code:     "duplicate_opencode_registration",
				Message:  fmt.Sprintf("opencode config at %s has %d plan-ai entries (expected 1)", r.OpenCodeConfigPath, oldCount+newCount),
			})
		}
		if r.StateExists && !r.OpenCodeValid {
			r.Issues = append(r.Issues, DoctorIssue{
				Severity: "warn",
				Code:     "stale_state_opencode_missing",
				Message:  fmt.Sprintf("state exists at %s but opencode config %s has no plan-ai entry", statePath, r.OpenCodeConfigPath),
			})
		}
	}

	return r
}

// ── project init ────────────────────────────────────────

// InitProject initializes a project after global install.
func (inst *Installer) InitProject(projectRoot string, opts InstallOptions) error {
	if inst.State == nil {
		if err := inst.LoadState(); err != nil {
			return fmt.Errorf("global install required: no state found at %s", inst.DataDir)
		}
	}

	// Create project config
	projCfgPath := config.ProjectConfigPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(projCfgPath), 0755); err != nil {
		return fmt.Errorf("mkdir project config dir: %w", err)
	}

	// Only write if not exists (idempotent)
	if _, err := os.Stat(projCfgPath); os.IsNotExist(err) {
		projCfg := config.ProjectConfig{
			Version:      stateVersion,
			ProjectName:  filepath.Base(projectRoot),
			ProjectRoot:  projectRoot,
			ProjectDB:    config.ProjectDBPath(projectRoot),
			CreatedAt:    timeNowUTC(),
			Integrations: map[string]any{},
		}
		if err := config.SaveProjectConfig(projCfgPath, projCfg); err != nil {
			return fmt.Errorf("save project config: %w", err)
		}
	}

	// Run project DB migrations if DB exists
	dbPath := config.ProjectDBPath(projectRoot)
	if _, err := os.Stat(dbPath); err == nil {
		db, openErr := openDB(dbPath)
		if openErr != nil {
			return fmt.Errorf("open project db: %w", openErr)
		}
		defer db.Close()
		if migrateErr := runMigrations(db); migrateErr != nil {
			return fmt.Errorf("project migrations: %w", migrateErr)
		}
	}

	return nil
}

// ── internal helpers ────────────────────────────────────

// resolveComponents maps a preset + custom components to a flat list.
func resolveComponents(opts InstallOptions) []string {
	if opts.Preset == "custom" {
		return opts.Components
	}
	if comps, ok := Presets[opts.Preset]; ok {
		return comps
	}
	// Default to full
	return Presets["full-plan-ai"]
}

// timeNowUTC returns the current UTC time as RFC3339.
func timeNowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// openCodeConfigDir returns the OPENCODE_CONFIG_DIR env var or
// the default ~/.config/opencode. When homeOverride is non-empty it
// is used in place of os.UserHomeDir() so callers can stay consistent
// with the installer's HomeDir (important for tests and sandboxed HOME).
func openCodeConfigDir(homeOverride string) string {
	if d := os.Getenv("OPENCODE_CONFIG_DIR"); d != "" {
		return d
	}
	home := homeOverride
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return ""
		}
	}
	return filepath.Join(home, ".config", "opencode")
}

// backupOpenCodeConfig copies existing opencode config files to DataDir/backups/.
func (inst *Installer) backupOpenCodeConfig() error {
	ocDir := openCodeConfigDir(inst.HomeDir)
	if ocDir == "" {
		return nil
	}

	// Check if any config exists
	candidates := []string{"opencode.json", "opencode.jsonc"}
	hasConfig := false
	for _, name := range candidates {
		if _, err := os.Stat(filepath.Join(ocDir, name)); err == nil {
			hasConfig = true
			break
		}
	}
	if !hasConfig {
		return nil // nothing to backup
	}

	backupDir := filepath.Join(inst.DataDir, "backups")
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("mkdir backups: %w", err)
	}

	timestamp := time.Now().UTC().Format(backupTimeFormat)
	for _, name := range candidates {
		src := filepath.Join(ocDir, name)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}
		dst := filepath.Join(backupDir, fmt.Sprintf("%s.%s.bak", name, timestamp))
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("write backup %s: %w", dst, err)
		}
	}

	return nil
}

// syncOpenCodeConfig generates or merges the Plan-AI MCP entry into the
// OpenCode config. It uses the existing opencode.SetupService under the hood.
func (inst *Installer) syncOpenCodeConfig(opts InstallOptions) error {
	ocDir := openCodeConfigDir(inst.HomeDir)
	if ocDir == "" {
		return fmt.Errorf("cannot determine OpenCode config dir")
	}

	if !opts.AllowReal {
		// Refuse to write to the user's actual ~/.config/opencode. We
		// compare against the OS-level real home (getpwuid_r) so that a
		// sandboxed or test HOME does not trip the safety check.
		if u, err := user.Current(); err == nil {
			realOCDir := filepath.Join(u.HomeDir, ".config", "opencode")
			if ocDir == realOCDir {
				return fmt.Errorf("refusing to write to %s without --allow-real-opencode flag", ocDir)
			}
		}
	}

	// Use the existing opencode.SetupService for the actual config generation
	// We'll call a helper function defined in sync.go
	return syncOpenCodeConfig(inst.HomeDir, opts.BinDir)
}

// removePlanAIFromOpenCode strips Plan-AI entries from the OpenCode config.
func (inst *Installer) removePlanAIFromOpenCode() error {
	ocDir := openCodeConfigDir(inst.HomeDir)
	if ocDir == "" {
		return nil
	}

	return removePlanAIFromOpenCodeConfig(ocDir)
}

// openDB is a type alias for opening a SQLite database.
// It's extracted so tests can replace it.
var openDB = func(path string) (interface{ Close() error }, error) {
	// This will be replaced with the real store.Open call.
	// For now we just check the file exists.
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("db not found: %s", path)
	}
	return &nopCloser{}, nil
}

type nopCloser struct{}

func (n *nopCloser) Close() error { return nil }

// runMigrations is a placeholder for the real migration logic.
var runMigrations = func(db interface{}) error {
	return nil // no-op for now
}
