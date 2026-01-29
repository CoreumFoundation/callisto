package assetft

import (
	"encoding/json"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/rs/zerolog/log"

	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

// HandleGenesis implements GenesisModule
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "assetft").Msg("parsing genesis")

	// Read the genesis state
	var genState assetfttypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[assetfttypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while unmarshaling assetft state: %s", err)
	}

	// Save the params
	err = m.db.SaveAssetFTParams(types.NewAssetFTParams(genState.Params, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis assetft params: %s", err)
	}

	return nil
}
