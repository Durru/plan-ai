package planning

import "testing"

func TestBuildPlanEvolutionBlueprintIncludesValidationSections(t *testing.T) {
	blueprint := BuildPlanEvolutionBlueprint("project", "Ship checkout", []string{"approved checkout"})
	if blueprint.Objective != "Ship checkout" || len(blueprint.Validations) == 0 || len(blueprint.Rollback) == 0 {
		t.Fatalf("unexpected blueprint: %#v", blueprint)
	}
	if blueprint.Scope[0] != "approved checkout" {
		t.Fatalf("scope not derived from approved inputs: %#v", blueprint.Scope)
	}
}
