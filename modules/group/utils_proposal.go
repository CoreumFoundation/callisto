package group

import (
	"fmt"

	"github.com/forbole/callisto/v4/types"
)

func (m *Module) GetProposal(proposalID uint64, height int64) (*types.GroupProposal, error) {
	groupProposal, err := m.source.GetProposal(proposalID, height)
	if err != nil {
		return nil, fmt.Errorf("error while getting staking pool snapshot: %s", err)
	}

	proposal := groupProposal.Proposal

	res := types.NewGroupProposal(
		proposal.Id,
		proposal.Title,
		proposal.Summary,
		proposal.Metadata,
		proposal.Messages,
		proposal.Status.String(),
		proposal.SubmitTime,
		proposal.VotingPeriodEnd,
		proposal.Proposers,
		proposal.GroupPolicyAddress,
		proposal.GroupVersion,
		proposal.GroupPolicyVersion,
	)

	return &res, nil
}
