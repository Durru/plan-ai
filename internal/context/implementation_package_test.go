package context

import "testing"

func TestBuildImplementationPackageProtectsRuntimePaths(t *testing.T) {
	pkg := BuildImplementationPackage("project", "plan", "opencode", "Build plan")
	if pkg.ModelTarget != "opencode" || len(pkg.FilesNotToTouch) == 0 || len(pkg.Validations) == 0 {
		t.Fatalf("unexpected implementation package: %#v", pkg)
	}
}
