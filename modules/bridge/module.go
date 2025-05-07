package bridge

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	bridgesource "github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/juno/v6/modules"
)

var (
	_ modules.MessageModule = &Module{}
	_ modules.Module        = &Module{}
)

// Module represent bridge module
type Module struct {
	bridgesource.Source
	cdc codec.Codec
	db  *database.Db
}

// NewModule returns a new Module instance
func NewModule(source bridgesource.Source, cdc codec.Codec, db *database.Db) *Module {
	return &Module{
		Source: source,
		cdc:    cdc,
		db:     db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "bridge"
}
