package assetnft

import (
	"encoding/json"
	"fmt"

	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/rs/zerolog/log"

	assetnfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/nft/types"
)

// HandleGenesis implements GenesisModule
func (m *Module) HandleGenesis(doc *tmtypes.GenesisDoc, appState map[string]json.RawMessage) error {
	log.Debug().Str("module", "assetnft").Msg("parsing genesis")

	// Read the genesis state
	var genState assetnfttypes.GenesisState
	err := m.cdc.UnmarshalJSON(appState[assetnfttypes.ModuleName], &genState)
	if err != nil {
		return fmt.Errorf("error while unmarshaling assetnft state: %s", err)
	}

	// Save the params
	err = m.db.SaveAssetNFTParams(types.NewAssetNFTParams(genState.Params, doc.InitialHeight))
	if err != nil {
		return fmt.Errorf("error while storing genesis assetnft params: %s", err)
	}

	return nil
}
