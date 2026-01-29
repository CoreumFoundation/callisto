package feemodel

import (
	"encoding/json"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/rs/zerolog/log"

	feemodeltypes "github.com/tokenize-x/tx-chain/v6/x/feemodel/types"
)

// HandleGenesis implements GenesisModule
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "feemodel").Msg("parsing genesis")

	// Read the genesis state
	var genState feemodeltypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[feemodeltypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while unmarshaling feemodel state: %s", err)
	}

	// Save the params
	err = m.db.SaveFeeModelParams(types.NewFeeModelParams(genState.Params, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis feemodel params: %s", err)
	}

	return nil
}
