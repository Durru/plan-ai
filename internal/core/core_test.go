package core

import "testing"

func TestNewAppStoresVersionAndRootDir(t *testing.T) {
	app := NewApp("v-test", "/workspace/project")

	if app.Version != "v-test" {
		t.Fatalf("Version = %q, want %q", app.Version, "v-test")
	}

	if app.RootDir != "/workspace/project" {
		t.Fatalf("RootDir = %q, want %q", app.RootDir, "/workspace/project")
	}
}
