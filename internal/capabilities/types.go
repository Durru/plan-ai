package capabilities

import "time"

// CapabilityType represents an abstract project skill — not a profile.
type CapabilityType string

const (
	CapVision         CapabilityType = "vision"
	CapResearch       CapabilityType = "research"
	CapPlanning       CapabilityType = "planning"
	CapArchitecture   CapabilityType = "architecture"
	CapDatabase       CapabilityType = "database"
	CapBackend        CapabilityType = "backend"
	CapFrontend       CapabilityType = "frontend"
	CapSecurity       CapabilityType = "security"
	CapTesting        CapabilityType = "testing"
	CapImpactAnalysis CapabilityType = "impact_analysis"
	CapValidation     CapabilityType = "validation"
)

// Capability defines a registered abstract skill.
type Capability struct {
	Type        CapabilityType `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
}

// WorkflowType maps capability to its common workflow type.
func (c Capability) WorkflowType() string {
	switch c.Type {
	case CapVision:
		return "vision"
	case CapResearch:
		return "research"
	case CapPlanning:
		return "planning"
	case CapValidation, CapImpactAnalysis:
		return "approval"
	default:
		return "planning"
	}
}
