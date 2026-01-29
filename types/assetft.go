package types

import (
	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

// AssetFTParams represents the parameters of the x/asset/ft module
type AssetFTParams struct {
	Params assetfttypes.Params
	Height int64
}

// NewAssetFTParams returns a new AssetFTParams instance
func NewAssetFTParams(params assetfttypes.Params, height int64) AssetFTParams {
	return AssetFTParams{
		Params: params,
		Height: height,
	}
}
