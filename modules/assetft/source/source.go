package source

import (
	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

type Source interface {
	GetParams(height int64) (assetfttypes.Params, error)
	GetDexSettings(height int64, denom string) (assetfttypes.DEXSettings, error)
}
