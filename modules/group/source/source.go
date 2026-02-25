package source

import (
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
)

type Source interface {
	GetProposal(proposalID uint64, height int64) (grouptypes.QueryProposalResponse, error)
}
