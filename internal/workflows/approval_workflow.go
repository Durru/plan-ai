package workflows

func ApprovalWorkflow() Workflow {
	return Workflow{
		Type: WorkflowTypeApproval,
		Name: "Approval Workflow",
		Steps: []string{
			"check_requirements",
			"validate",
			"approve_reject",
		},
	}
}
