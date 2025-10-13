package bridge

import (
	"encoding/hex"
	"fmt"
	"os"
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

			splitHeight, err := cmd.Flags().GetInt64("split-height")
			if err != nil {
				return err
			}

			if splitHeight <= 0 {
				splitHeight = 1000 // Default split height
			}

			// Create error log file
			timestamp := time.Now().Format("2006-01-02_15-04-05")
			errorLogFileName := fmt.Sprintf("bridge_errors_%d_%d_%s.log", startHeight, endHeight, timestamp)
			errorLogFile, err := os.OpenFile(errorLogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("failed to create error log file: %s", err)
			}
			defer errorLogFile.Close()

			// Create summary log file
			summaryLogFileName := fmt.Sprintf("bridge_summary_%d_%d_%s.log", startHeight, endHeight, timestamp)
			summaryLogFile, err := os.OpenFile(summaryLogFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("failed to create summary log file: %s", err)
			}
			defer summaryLogFile.Close()

			// Write header to summary log
			headerMsg := "=== Bridge Transaction Processing Summary ===\n"
			headerMsg += fmt.Sprintf("Start Height: %d\n", startHeight)
			headerMsg += fmt.Sprintf("End Height: %d\n", endHeight)
			headerMsg += fmt.Sprintf("Split Height: %d\n", splitHeight)
			headerMsg += fmt.Sprintf("Started at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			headerMsg += "============================================\n\n"
			if _, err := summaryLogFile.WriteString(headerMsg); err != nil {
				log.Error().Err(err).Msg("failed to write header to summary log file")
			}

			log.Info().Str("error_log", errorLogFileName).Str("summary_log", summaryLogFileName).Msg("log files created")

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
			for rangeStart := startHeight; rangeStart < endHeight; rangeStart += splitHeight {
				rangeEnd := rangeStart + splitHeight
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
							errorMsg := fmt.Sprintf("[%s] Height: %d | TxHash: %s | MsgIndex: %d | Error: %v\n",
								time.Now().Format("2006-01-02 15:04:05"),
								tx.Height,
								txHash,
								index,
								err,
							)

							// Write to error log file
							if _, writeErr := errorLogFile.WriteString(errorMsg); writeErr != nil {
								log.Error().Err(writeErr).Msg("failed to write to error log file")
							}

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

				// Write summary for this range
				summaryMsg := fmt.Sprintf("--- Range: %d to %d ---\n", rangeStart, rangeEnd)
				summaryMsg += fmt.Sprintf("Timestamp: %s\n", time.Now().Format("2006-01-02 15:04:05"))
				summaryMsg += fmt.Sprintf("Duration: %s\n", rangeDuration.String())
				summaryMsg += fmt.Sprintf("Transactions Found: %d\n", len(txs))
				summaryMsg += fmt.Sprintf("Messages Processed: %d\n", rangeMessagesProcessed)
				summaryMsg += fmt.Sprintf("Blocks Inserted: %d\n", rangeBlocksInserted)
				summaryMsg += fmt.Sprintf("Errors Encountered: %d\n", rangeErrorCount)
				if rangeMessagesProcessed > 0 {
					successRate := float64(rangeMessagesProcessed-rangeErrorCount) / float64(rangeMessagesProcessed) * 100
					summaryMsg += fmt.Sprintf("Success Rate: %.2f%%\n", successRate)
				} else {
					summaryMsg += "Success Rate: N/A (no messages processed)\n"
				}
				summaryMsg += "\n"

				if _, err := summaryLogFile.WriteString(summaryMsg); err != nil {
					log.Error().Err(err).Msg("failed to write to summary log file")
				}

				log.Info().Int64("range_start", rangeStart).Int64("range_end", rangeEnd).
					Int("txs", len(txs)).Int("errors", rangeErrorCount).
					Str("duration", rangeDuration.String()).
					Msg("finished processing height range")
			}

			// Write final summary
			rangeCount := (endHeight - startHeight + splitHeight - 1) / splitHeight

			finalSummary := "============================================\n"
			finalSummary += "=== Final Summary ===\n"
			finalSummary += fmt.Sprintf("Completed at: %s\n", time.Now().Format("2006-01-02 15:04:05"))
			finalSummary += fmt.Sprintf("Total Ranges Processed: %d\n", rangeCount)
			finalSummary += fmt.Sprintf("Total Transactions Processed: %d\n", totalTxsProcessed)
			finalSummary += fmt.Sprintf("Total Blocks Inserted: %d\n", totalBlocksInserted)
			finalSummary += fmt.Sprintf("Total Errors Encountered: %d\n", totalErrorsEncountered)
			finalSummary += "============================================\n"

			if _, err := summaryLogFile.WriteString(finalSummary); err != nil {
				log.Error().Err(err).Msg("failed to write final summary to log file")
			}

			log.Info().
				Str("error_log", errorLogFileName).
				Str("summary_log", summaryLogFileName).
				Int("total_txs", totalTxsProcessed).
				Int("total_errors", totalErrorsEncountered).
				Msg("processing complete")

			return nil
		},
	}

	cmd.Flags().Int64("start-height", 0, "Start height for filtering transactions (inclusive, 0 means no lower limit)")
	cmd.Flags().Int64("end-height", 0, "End height for filtering transactions (exclusive, 0 means no upper limit)")
	cmd.Flags().Int64("split-height", 100000, "Height range to split processing into chunks (default: 1000)")

	return cmd
}
