package source

import (
	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

type Source interface {
	GetParams(height int64) (dextypes.Params, error)
}
