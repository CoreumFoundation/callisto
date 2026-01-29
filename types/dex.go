package types

import (
	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

// DEXParams represents the parameters of the x/asset/ft module
type DEXParams struct {
	Params dextypes.Params
	Height int64
}

// NewDEXParams returns a new DEXParams instance
func NewDEXParams(params dextypes.Params, height int64) DEXParams {
	return DEXParams{
		Params: params,
		Height: height,
	}
}
