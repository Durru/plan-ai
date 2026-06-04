package continuous

// ApprovalService handles the approval workflow for plan update proposals.
type ApprovalService struct {
	repo PlanUpdateProposalRepository
}

// NewApprovalService creates a new ApprovalService.
func NewApprovalService(repo PlanUpdateProposalRepository) *ApprovalService {
	return &ApprovalService{repo: repo}
}

// RequestApproval moves a proposal from draft to pending_approval.
func (s *ApprovalService) RequestApproval(proposalID string) (PlanUpdateProposal, error) {
	proposal, err := s.repo.GetProposal(proposalID)
	if err != nil {
		return PlanUpdateProposal{}, err
	}
	if proposal.Status != ProposalDraft {
		return proposal, nil // already requested or further along
	}
	if err := s.repo.UpdateProposalStatus(proposalID, ProposalPendingApproval); err != nil {
		return PlanUpdateProposal{}, err
	}
	proposal.Status = ProposalPendingApproval
	return proposal, nil
}
