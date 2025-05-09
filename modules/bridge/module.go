package bridge

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	bridgesource "github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/juno/v6/modules"
	"github.com/forbole/juno/v6/types/config"
)

var (
	_ modules.MessageModule = &Module{}
	_ modules.Module        = &Module{}
)

// Module represent bridge module
type Module struct {
	cfg *Config
	bridgesource.Source
	cdc codec.Codec
	db  *database.Db
}

// NewModule returns a new Module instance
func NewModule(cfg config.Config, source bridgesource.Source, cdc codec.Codec, db *database.Db) *Module {
	bz, err := cfg.GetBytes()
	if err != nil {
		panic(err)
	}

	bridgeCfg, err := ParseConfig(bz)
	if err != nil {
		panic(err)
	}

	return &Module{
		cfg:    bridgeCfg,
		Source: source,
		cdc:    cdc,
		db:     db,
	}
}

// Name implements modules.Module
func (m *Module) Name() string {
	return "bridge"
}
