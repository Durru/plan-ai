package research

import (
	"fmt"
)

// ApprovalChecker validates that a research entry can be approved.
// Requires: at least 1 finding, 1 source, 1 conclusion.
type ApprovalChecker struct {
	repo Repository
}

// NewApprovalChecker creates a new approval checker.
func NewApprovalChecker(repo Repository) *ApprovalChecker {
	return &ApprovalChecker{repo: repo}
}

// CanApprove checks whether a research entry can be approved.
func (c *ApprovalChecker) CanApprove(id string) error {
	entry, err := c.repo.GetEntry(id)
	if err != nil {
		return fmt.Errorf("research entry %q not found: %w", id, err)
	}
	if entry.Status == ResearchStatusApproved {
		return fmt.Errorf("research entry %q is already approved", id)
	}

	var errs ValidationErrors

	findings, err := c.repo.ListFindings(id)
	if err != nil {
		return err
	}
	if len(findings) == 0 {
		errs = append(errs, ValidationError{Field: "findings", Message: "at least one finding is required"})
	}

	sources, err := c.repo.ListSources(id)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		errs = append(errs, ValidationError{Field: "sources", Message: "at least one source is required"})
	}

	conclusions, err := c.repo.ListConclusions(id)
	if err != nil {
		return err
	}
	if len(conclusions) == 0 {
		errs = append(errs, ValidationError{Field: "conclusions", Message: "at least one conclusion is required"})
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}
