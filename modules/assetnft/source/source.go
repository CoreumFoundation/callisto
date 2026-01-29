package source

import (
	assetnfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/nft/types"
)

type Source interface {
	GetParams(height int64) (assetnfttypes.Params, error)
}
