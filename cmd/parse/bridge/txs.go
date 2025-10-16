package bridge

import (
	"encoding/hex"
	"fmt"
	"sort"
	"time"

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

			endHeight, err := cmd.Flags().GetInt64("end-height")
			if err != nil {
				return err
			}

			checkSize, err := cmd.Flags().GetInt64("check-size")
			if err != nil {
				return err
			}

			if checkSize <= 0 {
				checkSize = 1000 // Default check size
			}

			// Log processing summary header
			log.Info().Msg("=== Bridge Transaction Processing Summary ===")
			log.Info().Int64("start_height", startHeight).Int64("end_height", endHeight).Int64("check_size", checkSize).Msg("starting bridge transaction processing")

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

			// Track overall statistics
			totalTxsProcessed := 0
			totalErrorsEncountered := 0
			totalBlocksInserted := 0

			// Split the height range into chunks
			for rangeStart := startHeight; rangeStart < endHeight; rangeStart += checkSize {
				rangeEnd := rangeStart + checkSize
				if rangeEnd > endHeight {
					rangeEnd = endHeight
				}

				rangeStartTime := time.Now()

				log.Info().Int64("range_start", rangeStart).Int64("range_end", rangeEnd).
					Msg("processing height range")

				// Track statistics for this range
				rangeErrorCount := 0
				rangeBlocksInserted := 0
				rangeMessagesProcessed := 0

				// Collect all the transactions for this range
				var txs []*tmctypes.ResultTx

				// Get all the MsgSendToXrpl txs
				query := fmt.Sprintf(
					"tx.height >= %d AND tx.height < %d AND wasm._contract_address='%s' AND wasm.action='send_to_xrpl'",
					rangeStart,
					rangeEnd,
					bridgeCfg.ContractAddress,
				)
				sendToXrplTxs, err := utils.QueryTxs(parseCtx.Node, query)
				if err != nil {
					return err
				}
				txs = append(txs, sendToXrplTxs...)

				// Get all the MsgSaveEvidence txs
				query = fmt.Sprintf(
					"tx.height >= %d AND tx.height < %d AND wasm._contract_address='%s' AND wasm.action='save_evidence'",
					rangeStart,
					rangeEnd,
					bridgeCfg.ContractAddress,
				)
				saveEvidenceTxs, err := utils.QueryTxs(parseCtx.Node, query)
				if err != nil {
					return err
				}
				txs = append(txs, saveEvidenceTxs...)

				log.Info().Int64("range_start", rangeStart).Int64("range_end", rangeEnd).
					Int("tx_count", len(txs)).Msg("found transactions in range")

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
						rangeMessagesProcessed++
						err = bridgeModule.HandleMsg(index, transaction.Body.Messages[index], transaction)
						if err != nil {
							rangeErrorCount++
							txHash := hex.EncodeToString(tx.Tx.Hash())

							log.Error().Int64("height", tx.Height).Int("msg_index", index).
								Str("tx_hash", txHash).
								Err(err).Msg("error while handling bridge module message")
							continue
						}
					}
				}

				// Calculate range duration
				rangeDuration := time.Since(rangeStartTime)

				// Update overall statistics
				totalTxsProcessed += len(txs)
				totalErrorsEncountered += rangeErrorCount
				totalBlocksInserted += rangeBlocksInserted

				// Log summary for this range
				log.Info().Msg("--- Range Summary ---")
				if rangeMessagesProcessed > 0 {
					successRate := float64(rangeMessagesProcessed-rangeErrorCount) / float64(rangeMessagesProcessed) * 100
					log.Info().Int64("range_start", rangeStart).Int64("range_end", rangeEnd).
						Int("txs_found", len(txs)).Int("messages_processed", rangeMessagesProcessed).
						Int("blocks_inserted", rangeBlocksInserted).Int("errors", rangeErrorCount).
						Float64("success_rate", successRate).
						Str("duration", rangeDuration.String()).
						Msg("finished processing height range")
				} else {
					log.Info().Int64("range_start", rangeStart).Int64("range_end", rangeEnd).
						Int("txs_found", len(txs)).Int("messages_processed", rangeMessagesProcessed).
						Int("blocks_inserted", rangeBlocksInserted).Int("errors", rangeErrorCount).
						Str("success_rate", "N/A").
						Str("duration", rangeDuration.String()).
						Msg("finished processing height range")
				}
			}

			// Log final summary
			rangeCount := (endHeight - startHeight + checkSize - 1) / checkSize

			log.Info().Msg("============================================")
			log.Info().Msg("=== Final Summary ===")
			log.Info().
				Int64("total_ranges_processed", rangeCount).
				Int("total_txs_processed", totalTxsProcessed).
				Int("total_blocks_inserted", totalBlocksInserted).
				Int("total_errors_encountered", totalErrorsEncountered).
				Msg("processing complete")
			log.Info().Msg("============================================")

			return nil
		},
	}

	cmd.Flags().Int64("start-height", 0, "Start height for filtering transactions (inclusive, 0 means no lower limit)")
	cmd.Flags().Int64("end-height", 0, "End height for filtering transactions (exclusive, 0 means no upper limit)")
	cmd.Flags().Int64("check-size", 100000, "Size of the height range to check (default: 100000)")

	return cmd
}
