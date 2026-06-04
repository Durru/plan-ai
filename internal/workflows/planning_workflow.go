package workflows

func PlanningWorkflow() Workflow {
	return Workflow{Type: WorkflowTypePlanning, Name: "Planning Workflow", Steps: []string{"Vision", "Approved Context", "Research", "Knowledge", "Master Plan", "Specific Plan"}}
}
