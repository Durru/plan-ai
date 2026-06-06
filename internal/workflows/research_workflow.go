package workflows

func ResearchWorkflow() Workflow {
	return Workflow{
		Type: WorkflowTypeResearch,
		Name: "Research Workflow",
		Steps: []string{
			"find_reusable",
			"create_research",
			"approve_research",
			"promote_to_knowledge",
		},
	}
}
