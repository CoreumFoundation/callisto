package group

import (
	"encoding/json"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/rs/zerolog/log"

	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
)

// HandleGenesis implements GenesisModule
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "dex").Msg("parsing genesis")

	// Read the genesis state
	var genState grouptypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[grouptypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while unmarshaling DEX state: %s", err)
	}

	return nil
}
