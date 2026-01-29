package source

import (
	feemodeltypes "github.com/tokenize-x/tx-chain/v6/x/feemodel/types"
)

type Source interface {
	GetParams(height int64) (feemodeltypes.Params, error)
}
