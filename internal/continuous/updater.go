package continuous

import "fmt"

// Updater applies approved plan updates.
type Updater struct {
	repo PlanUpdateProposalRepository
}

// NewUpdater creates a new Updater.
func NewUpdater(repo PlanUpdateProposalRepository) *Updater {
	return &Updater{repo: repo}
}

// Approve marks a proposal as approved.
func (u *Updater) Approve(proposalID string) (PlanUpdateProposal, error) {
	proposal, err := u.repo.GetProposal(proposalID)
	if err != nil {
		return PlanUpdateProposal{}, fmt.Errorf("get proposal: %w", err)
	}
	if proposal.Status != ProposalDraft && proposal.Status != ProposalPendingApproval {
		return proposal, fmt.Errorf("proposal %s is in status %s, cannot approve", proposalID, proposal.Status)
	}
	if err := u.repo.UpdateProposalStatus(proposalID, ProposalApproved); err != nil {
		return PlanUpdateProposal{}, err
	}
	proposal.Status = ProposalApproved
	return proposal, nil
}

// Reject marks a proposal as rejected.
func (u *Updater) Reject(proposalID string) (PlanUpdateProposal, error) {
	proposal, err := u.repo.GetProposal(proposalID)
	if err != nil {
		return PlanUpdateProposal{}, fmt.Errorf("get proposal: %w", err)
	}
	if err := u.repo.UpdateProposalStatus(proposalID, ProposalRejected); err != nil {
		return PlanUpdateProposal{}, err
	}
	proposal.Status = ProposalRejected
	return proposal, nil
}

// Apply marks a proposal as applied.
func (u *Updater) Apply(proposalID string) (PlanUpdateProposal, error) {
	proposal, err := u.repo.GetProposal(proposalID)
	if err != nil {
		return PlanUpdateProposal{}, fmt.Errorf("get proposal: %w", err)
	}
	if proposal.Status != ProposalApproved {
		return proposal, fmt.Errorf("proposal %s is not approved (status: %s)", proposalID, proposal.Status)
	}
	if err := u.repo.UpdateProposalStatus(proposalID, ProposalApplied); err != nil {
		return PlanUpdateProposal{}, err
	}
	proposal.Status = ProposalApplied
	return proposal, nil
}
