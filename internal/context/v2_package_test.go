package context

import "testing"

func TestBuildSmartPackageSetsBudgetAndPriority(t *testing.T) {
	pkg := BuildSmartPackage("project", PackageImplementation, "qwen", "Implement checkout with tests", 1024)
	if pkg.TokenBudget != 1024 || pkg.Priority != 1 || pkg.ModelTarget != "qwen" {
		t.Fatalf("unexpected package: %#v", pkg)
	}
}
