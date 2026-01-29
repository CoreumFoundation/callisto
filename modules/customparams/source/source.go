package source

import (
	customparamstypes "github.com/tokenize-x/tx-chain/v6/x/customparams/types"
)

type Source interface {
	GetParams(height int64) (customparamstypes.StakingParams, error)
}
