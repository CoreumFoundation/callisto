package bridge

import (
	"encoding/hex"
	"fmt"
	"sort"

	tmctypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/bridge"
	modulestypes "github.com/forbole/callisto/v4/modules/types"
	"github.com/forbole/callisto/v4/utils"
	parsecmdtypes "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/forbole/juno/v6/types/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// txsCmd returns the Cobra command allowing re-scan bridge transactions
func txsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txs",
		Short: "Parse all the bridge transactions and overwrite the existing ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			startHeight, err := cmd.Flags().GetInt64("start-height")
			if err != nil {
				return err
			}

			bz, err := config.Cfg.GetBytes()
			if err != nil {
				return err
			}
			bridgeCfg, err := bridge.ParseConfig(bz)
			if err != nil {
				return err
			}

			sources, err := modulestypes.BuildSources(config.Cfg.Node, utils.GetCodec())
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)

			// Build bridge module
			bridgeModule := bridge.NewModule(config.Cfg, sources.BridgeSource, utils.GetCodec(), db)

			// Get the accounts
			// Collect all the transactions
			var txs []*tmctypes.ResultTx

			// Get all the MsgSendToXrpl txs
			query := fmt.Sprintf(
				"tx.height >= %d AND wasm._contract_address='%s' AND wasm.action='send_to_xrpl'",
				startHeight,
				bridgeCfg.ContractAddress,
			)
			sendToXrplTxs, err := utils.QueryTxs(parseCtx.Node, query)
			if err != nil {
				return err
			}
			txs = append(txs, sendToXrplTxs...)

			// Get all the MsgSaveEvidence txs
			query = fmt.Sprintf(
				"tx.height >= %d AND wasm._contract_address='%s' AND wasm.action='save_evidence'",
				startHeight,
				bridgeCfg.ContractAddress,
			)
			saveEvidenceTxs, err := utils.QueryTxs(parseCtx.Node, query)
			if err != nil {
				return err
			}
			txs = append(txs, saveEvidenceTxs...)

			// Sort the txs based on their ascending height
			sort.Slice(txs, func(i, j int) bool {
				return txs[i].Height < txs[j].Height
			})

			for _, tx := range txs {
				log.Debug().Int64("height", tx.Height).Msg("parsing transaction")
				transaction, err := parseCtx.Node.Tx(hex.EncodeToString(tx.Tx.Hash()))
				if err != nil {
					return err
				}

				for index := range transaction.Body.Messages {
					err = bridgeModule.HandleMsg(index, transaction.Body.Messages[index], transaction)
					if err != nil {
						return fmt.Errorf("error while handling bridge module message: %s", err)
					}
				}
			}

			return nil
		},
	}

	cmd.Flags().Int64("start-height", 0, "Start height for filtering transactions (inclusive, 0 means no lower limit)")

	return cmd
}
