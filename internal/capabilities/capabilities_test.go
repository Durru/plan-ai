package capabilities_test

import (
	"testing"

	"github.com/plan-ai/plan-ai/internal/capabilities"
)

func TestNewDefaultRegistry(t *testing.T) {
	r := capabilities.NewDefaultRegistry()
	caps := r.ListCapabilities()
	// We expect a fixed set: vision, research, planning, architecture, database, backend, frontend, security, testing, impact_analysis, validation
	if len(caps) != 11 {
		t.Fatalf("got %d capabilities, want 11", len(caps))
	}

	// Verify sorted by type
	for i := 1; i < len(caps); i++ {
		if caps[i].Type <= caps[i-1].Type {
			t.Errorf("capabilities not sorted: %s <= %s", caps[i].Type, caps[i-1].Type)
		}
	}

	// Spot-check known capabilities
	check := func(ct capabilities.CapabilityType) {
		c, err := r.GetCapability(ct)
		if err != nil {
			t.Errorf("GetCapability(%q): %v", ct, err)
		}
		if c.Name == "" {
			t.Errorf("capability %q has empty name", ct)
		}
	}
	check(capabilities.CapVision)
	check(capabilities.CapResearch)
	check(capabilities.CapPlanning)
	check(capabilities.CapArchitecture)
	check(capabilities.CapBackend)
	check(capabilities.CapFrontend)
	check(capabilities.CapTesting)
	check(capabilities.CapSecurity)
	check(capabilities.CapImpactAnalysis)
	check(capabilities.CapValidation)
	check(capabilities.CapDatabase)
}

func TestRegisterAndGetCapability(t *testing.T) {
	r := capabilities.NewRegistry()

	if err := r.RegisterCapability(capabilities.Capability{Type: "custom", Name: "Custom Skill", Description: "A custom registered skill"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	c, err := r.GetCapability("custom")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c.Name != "Custom Skill" {
		t.Errorf("name = %q", c.Name)
	}

	// Get nonexistent
	if _, err := r.GetCapability("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent capability")
	}
}

func TestRegisterCapabilityValidation(t *testing.T) {
	r := capabilities.NewRegistry()

	if err := r.RegisterCapability(capabilities.Capability{Type: "", Name: "No Type"}); err == nil {
		t.Fatal("expected error for empty type")
	}
	if err := r.RegisterCapability(capabilities.Capability{Type: "notype", Name: ""}); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestListCapabilities(t *testing.T) {
	r := capabilities.NewRegistry()
	if len(r.ListCapabilities()) != 0 {
		t.Fatal("new empty registry should have 0 capabilities")
	}

	r.RegisterCapability(capabilities.Capability{Type: "b", Name: "B"})
	r.RegisterCapability(capabilities.Capability{Type: "a", Name: "A"})
	r.RegisterCapability(capabilities.Capability{Type: "c", Name: "C"})

	list := r.ListCapabilities()
	if len(list) != 3 {
		t.Fatalf("len = %d, want 3", len(list))
	}
	// Must be sorted
	if list[0].Type != "a" || list[1].Type != "b" || list[2].Type != "c" {
		t.Errorf("order: %v", list)
	}
}

func TestRegisterCapabilityOverwritesExisting(t *testing.T) {
	r := capabilities.NewRegistry()
	r.RegisterCapability(capabilities.Capability{Type: "existing", Name: "Old"})
	r.RegisterCapability(capabilities.Capability{Type: "existing", Name: "New"})

	c, _ := r.GetCapability("existing")
	if c.Name != "New" {
		t.Errorf("name = %q, want New", c.Name)
	}
}

func TestWorkflowTypeMapping(t *testing.T) {
	tests := []struct {
		ct   capabilities.CapabilityType
		want string
	}{
		{capabilities.CapVision, "vision"},
		{capabilities.CapResearch, "research"},
		{capabilities.CapPlanning, "planning"},
		{capabilities.CapValidation, "approval"},
		{capabilities.CapImpactAnalysis, "approval"},
		{capabilities.CapDatabase, "planning"},
		{capabilities.CapSecurity, "planning"},
	}
	for _, tt := range tests {
		c := capabilities.Capability{Type: tt.ct}
		if got := c.WorkflowType(); got != tt.want {
			t.Errorf("WorkflowType(%q) = %q, want %q", tt.ct, got, tt.want)
		}
	}
}
