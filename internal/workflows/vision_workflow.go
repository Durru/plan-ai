package workflows

func VisionWorkflow() Workflow {
	return Workflow{Type: WorkflowTypeVision, Name: "Vision Workflow", Steps: []string{"Input", "Vision", "Approval"}}
}
