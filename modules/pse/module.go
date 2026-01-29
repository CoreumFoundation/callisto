package pse

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/juno/v6/modules"
)

var (
	_ modules.BlockModule = &Module{}
	_ modules.Module      = &Module{}
)

// Module represent x/pse module
type Module struct {
	cdc          codec.Codec
	db           *database.Db
	bech32Prefix string
}

// NewModule returns a new Module instance
func NewModule(cdc codec.Codec, db *database.Db, bech32Prefix string) *Module {
	return &Module{
		cdc:          cdc,
		db:           db,
		bech32Prefix: bech32Prefix,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "pse"
}
