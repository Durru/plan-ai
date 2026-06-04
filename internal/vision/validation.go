package vision

func MissingInformation(d Draft) []string {
	var missing []string
	if d.Summary == "" {
		missing = append(missing, "main objective")
	}
	if len(d.TargetUsers) == 0 {
		missing = append(missing, "target users")
	}
	if d.ExpectedOutcome == "" && len(d.SuccessCriteria) == 0 {
		missing = append(missing, "expected outcome or success criteria")
	}
	if len(d.FunctionalGoals) == 0 {
		missing = append(missing, "functional goals")
	}
	return missing
}
