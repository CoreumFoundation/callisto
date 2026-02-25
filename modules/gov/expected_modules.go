package gov

import (
	"github.com/forbole/callisto/v4/types"
)

type AuthModule interface {
	UpdateParams(height int64) error
}

type DistrModule interface {
	UpdateParams(height int64) error
}

type MintModule interface {
	UpdateParams(height int64) error
	UpdateInflation() error
}

type SlashingModule interface {
	UpdateParams(height int64) error
}

type StakingModule interface {
	GetStakingPoolSnapshot(height int64) (*types.PoolSnapshot, error)
	UpdateParams(height int64) error
}

type FeeModelModule interface {
	UpdateParams(height int64) error
}

type CustomParamsModule interface {
	UpdateParams(height int64) error
}

type AssetFTModule interface {
	UpdateParams(height int64) error
}

type AssetNFTModule interface {
	UpdateParams(height int64) error
}

type DEXModule interface {
	UpdateParams(height int64) error
}

type GroupModule interface {
	GetProposal(proposalID uint64, height int64) (*types.GroupProposal, error)
}
