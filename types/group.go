package types

import (
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
)

// GovProposal contains the data of the x/gov module parameters
type GovProposal struct {
	*grouptypes.Proposal
	Height int64 `json:"height" ymal:"height"`
}

func NewGovProposal(proposal *grouptypes.Proposal, height int64) *GovProposal {
	return &GovProposal{
		Proposal: proposal,
		Height:   height,
	}
}

// --------------------------------------------------------------------------------------------------------------------

// GroupProposal represents a single group proposal
type GroupProposal struct {
	ID                 uint64
	Title              string
	Summary            string
	Metadata           string
	Messages           []*codectypes.Any
	Status             string
	SubmitTime         time.Time
	VotingPeriodEnd    time.Time
	Proposers          []string
	GroupPolicyAddress string
	GroupVersion       uint64
	GroupPolicyVersion uint64
}

// NewGroupProposal return a new group Proposal instance
func NewGroupProposal(
	proposalID uint64,
	title string,
	summary string,
	metadata string,
	messages []*codectypes.Any,
	status string,
	submitTime time.Time,
	votingPeriodEnd time.Time,
	proposers []string,
	groupPolicyAddress string,
	groupVersion uint64,
	groupPolicyVersion uint64,
) GroupProposal {
	return GroupProposal{
		ID:                 proposalID,
		Title:              title,
		Summary:            summary,
		Metadata:           metadata,
		Messages:           messages,
		Status:             status,
		SubmitTime:         submitTime,
		VotingPeriodEnd:    votingPeriodEnd,
		Proposers:          proposers,
		GroupPolicyAddress: groupPolicyAddress,
		GroupVersion:       groupVersion,
		GroupPolicyVersion: groupPolicyVersion,
	}
}
