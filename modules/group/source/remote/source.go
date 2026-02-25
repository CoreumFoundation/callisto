package remote

import (
	groupsource "github.com/forbole/callisto/v4/modules/group/source"
	"github.com/forbole/juno/v6/node/remote"

	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
)

var _ groupsource.Source = &Source{}

// Source implements groupsource.Source using a remote node
type Source struct {
	*remote.Source
	groupClient grouptypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, dexClient grouptypes.QueryClient) *Source {
	return &Source{
		Source:      source,
		groupClient: dexClient,
	}
}

// GetProposal implements groupsource.Source
func (s Source) GetProposal(proposalID uint64, height int64) (grouptypes.QueryProposalResponse, error) {
	res, err := s.groupClient.Proposal(remote.GetHeightRequestContext(s.Ctx, height), &grouptypes.QueryProposalRequest{
		ProposalId: proposalID,
	})
	if err != nil {
		return grouptypes.QueryProposalResponse{}, err
	}

	return *res, nil
}
