package types

import (
	feemodeltypes "github.com/tokenize-x/tx-chain/v6/x/feemodel/types"
)

// FeeModelParams represents the parameters of the x/feemodel module
type FeeModelParams struct {
	Params feemodeltypes.Params
	Height int64
}

// NewFeeModelParams returns a new FeeModelParams instance
func NewFeeModelParams(params feemodeltypes.Params, height int64) FeeModelParams {
	return FeeModelParams{
		Params: params,
		Height: height,
	}
}
