package workflows

func PlanningWorkflow() Workflow {
	return Workflow{
		Type: WorkflowTypePlanning,
		Name: "Planning Workflow",
		Steps: []string{
			"load_approved_context",
			"create_master_plan",
			"create_specific_plan",
			"approve_plans",
		},
	}
}
