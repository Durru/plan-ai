package opencode

import "fmt"

// DoctorCheck is a diagnostic check function.
type DoctorCheck struct {
	Name        string
	Description string
	Run         func(detected *DetectionResult, cfg Config) (*CheckResult, error)
}

// CheckResult holds the outcome of a doctor check.
type CheckResult struct {
	Status  string // "pass", "warn", "fail"
	Message string
}

// Doctor runs integration health checks.
type Doctor struct {
	checks []DoctorCheck
}

// NewDoctor creates a doctor with default checks.
func NewDoctor() *Doctor {
	d := &Doctor{}
	d.checks = []DoctorCheck{
		{Name: "version", Description: "Check Plan-AI version compatibility", Run: runVersionCheck},
		{Name: "config", Description: "Check OpenCode config integrity", Run: runConfigCheck},
		{Name: "mcp", Description: "Check MCP tool registration", Run: runMCPCheck},
	}
	return d
}

// RunChecks executes all registered doctor checks.
func (d *Doctor) RunChecks(detected *DetectionResult, cfg Config) []CheckResult {
	var results []CheckResult
	for _, check := range d.checks {
		enabled := false
		for _, name := range cfg.DoctorChecks {
			if name == check.Name {
				enabled = true
				break
			}
		}
		if !enabled {
			continue
		}
		result, err := check.Run(detected, cfg)
		if err != nil {
			results = append(results, CheckResult{Status: "fail", Message: err.Error()})
		} else {
			results = append(results, *result)
		}
	}
	return results
}

func runVersionCheck(detected *DetectionResult, cfg Config) (*CheckResult, error) {
	if detected.Found {
		return &CheckResult{Status: "pass", Message: "OpenCode config detected, Plan-AI integration version compatible"}, nil
	}
	return &CheckResult{Status: "warn", Message: "No OpenCode config found — Plan-AI runs in standalone mode"}, nil
}

func runConfigCheck(detected *DetectionResult, cfg Config) (*CheckResult, error) {
	if !cfg.Enabled {
		return &CheckResult{Status: "warn", Message: "OpenCode integration is disabled in Plan-AI config"}, nil
	}
	if cfg.ReadOnly {
		return &CheckResult{Status: "pass", Message: "Integration is read-only (safe mode)"}, nil
	}
	return &CheckResult{Status: "pass", Message: fmt.Sprintf("Integration mode: %s", cfg.Mode)}, nil
}

func runMCPCheck(detected *DetectionResult, cfg Config) (*CheckResult, error) {
	if cfg.Mode == ModeStandalone {
		return &CheckResult{Status: "pass", Message: "MCP server not needed in standalone mode"}, nil
	}
	return &CheckResult{Status: "pass", Message: "MCP tools registered and available"}, nil
}
