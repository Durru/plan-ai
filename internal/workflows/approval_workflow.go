package workflows

func ApprovalWorkflow() Workflow {
	return Workflow{Type: WorkflowTypeApproval, Name: "Approval Workflow", Steps: []string{"Draft", "Review", "Approved/Rejected"}}
}
