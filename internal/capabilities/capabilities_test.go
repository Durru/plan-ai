package capabilities_test

import (
	"database/sql"
	"testing"

	"github.com/plan-ai/plan-ai/internal/capabilities"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if _, err := db.Exec(capabilitiesV2Schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
	return db
}

const capabilitiesV2Schema = `
CREATE TABLE IF NOT EXISTS capabilities_v2 (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  schema_info TEXT NOT NULL DEFAULT '{}',
  version TEXT NOT NULL DEFAULT '1.0',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);`

func TestNewDefaultRegistry(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	r := capabilities.NewDefaultRegistry(db)
	caps := r.ListCapabilities()
	if len(caps) != 11 {
		t.Fatalf("got %d capabilities, want 11", len(caps))
	}

	for i := 1; i < len(caps); i++ {
		if caps[i].Type <= caps[i-1].Type {
			t.Errorf("capabilities not sorted: %s <= %s", caps[i].Type, caps[i-1].Type)
		}
	}

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
	db := openTestDB(t)
	defer db.Close()
	r := capabilities.NewRegistry(db)

	if err := r.RegisterCapability(capabilities.Capability{Type: "custom", Name: "custom", Description: "A custom registered skill"}); err != nil {
		t.Fatalf("register: %v", err)
	}

	c, err := r.GetCapability("custom")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c.Name != "custom" {
		t.Errorf("name = %q", c.Name)
	}

	if _, err := r.GetCapability("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent capability")
	}
}

func TestRegisterCapabilityValidation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	r := capabilities.NewRegistry(db)

	if err := r.RegisterCapability(capabilities.Capability{Type: "", Name: "No Type"}); err == nil {
		t.Fatal("expected error for empty type")
	}
}

func TestListCapabilities(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	r := capabilities.NewRegistry(db)
	// After auto-seeding, registry has 11 default capabilities
	base := len(r.ListCapabilities())
	if base != 11 {
		t.Fatalf("seeded registry should have 11 capabilities, got %d", base)
	}

	r.RegisterCapability(capabilities.Capability{Type: "b", Name: "b", Enabled: true})
	r.RegisterCapability(capabilities.Capability{Type: "a", Name: "a", Enabled: true})
	r.RegisterCapability(capabilities.Capability{Type: "c", Name: "c", Enabled: true})

	list := r.ListCapabilities()
	if len(list) != base+3 {
		t.Fatalf("len = %d, want %d", len(list), base+3)
	}
	// Verify sorting within the new items
	found := []string{}
	for _, c := range list {
		found = append(found, c.Name)
	}
	// a, b, c must be in order among themselves
	idxA, idxB, idxC := -1, -1, -1
	for i, name := range found {
		switch name {
		case "a":
			idxA = i
		case "b":
			idxB = i
		case "c":
			idxC = i
		}
	}
	if idxA >= idxB || idxB >= idxC {
		t.Errorf("custom items not sorted: a=%d b=%d c=%d", idxA, idxB, idxC)
	}
}

func TestRegisterCapabilityOverwritesExisting(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	r := capabilities.NewRegistry(db)
	r.RegisterCapability(capabilities.Capability{Type: "existing", Name: "existing"})
	r.RegisterCapability(capabilities.Capability{Type: "existing", Name: "existing", Description: "Updated"})

	c, _ := r.GetCapability("existing")
	if c.Description != "Updated" {
		t.Errorf("description = %q, want Updated", c.Description)
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
