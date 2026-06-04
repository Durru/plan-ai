package workflows

func ResearchWorkflow() Workflow {
	return Workflow{Type: WorkflowTypeResearch, Name: "Research Workflow", Steps: []string{"Topic", "Research", "Knowledge"}}
}
