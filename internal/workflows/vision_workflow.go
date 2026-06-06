package workflows

func VisionWorkflow() Workflow {
	return Workflow{
		Type: WorkflowTypeVision,
		Name: "Vision Workflow",
		Steps: []string{
			"detect_intent",
			"create_discovery",
			"approve_intent",
		},
	}
}
