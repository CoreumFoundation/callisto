package dex

import (
	"encoding/json"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/rs/zerolog/log"

	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

// HandleGenesis implements GenesisModule
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "dex").Msg("parsing genesis")

	// Read the genesis state
	var genState dextypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[dextypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while unmarshaling DEX state: %s", err)
	}

	// Save the params
	err = m.db.SaveDEXParams(types.NewDEXParams(genState.Params, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis DEX params: %s", err)
	}

	return nil
}
