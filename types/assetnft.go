package types

import (
	assetnfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/nft/types"
)

// AssetNFTParams represents the parameters of the x/asset/nft module
type AssetNFTParams struct {
	Params assetnfttypes.Params
	Height int64
}

// NewAssetNFTParams returns a new AssetNFTParams instance
func NewAssetNFTParams(params assetnfttypes.Params, height int64) AssetNFTParams {
	return AssetNFTParams{
		Params: params,
		Height: height,
	}
}
