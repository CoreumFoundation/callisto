package bridge

import (
	"fmt"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/forbole/callisto/v4/utils"
	juno "github.com/forbole/juno/v6/types"
	"github.com/rs/zerolog/log"
)

var msgFilter = map[string]bool{
	"/cosmwasm.wasm.v1.MsgExecuteContract": true,
}

// HandleMsgExec implements modules.AuthzMessageModule
func (m *Module) HandleMsgExec(index, _ int, executedMsg juno.Message, tx *juno.Transaction) error {
	return m.HandleMsg(index, executedMsg, tx)
}

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(
	_ int, msg juno.Message, tx *juno.Transaction,
) error {
	if _, ok := msgFilter[msg.GetType()]; !ok {
		return nil
	}

	cosmosMsg := utils.UnpackMessage(m.cdc, msg.GetBytes(), &wasmtypes.MsgExecuteContract{})

	// TODO: move this to a configuration
	if cosmosMsg.Contract != types.BridgeContractAddress {
		return nil
	}

	log.Debug().Str("module", "bridge").Str("hash", tx.TxHash).Uint64("height", tx.Height).Msg(fmt.Sprintf("handling bridge message %s", msg.GetType()))

	err := m.addCoreumToXRPLTransfer(tx.Height, msg, tx)
	if err != nil {
		fmt.Printf("Error when adding Coreum to XRPL transfer, error: %s", err)
	}
	return nil
}

// addCoreumToXRPLTransfer
func (m *Module) addCoreumToXRPLTransfer(height uint64, _ juno.Message, tx *juno.Transaction) error {
	events := juno.FindEventsByType(tx.Events, "wasm")
	for _, event := range events {
		action, err := juno.FindAttributeByKey(event, "action")
		if err != nil {
			return fmt.Errorf("error while getting action attribute: %s", err)
		}
		switch action.Value {
		case "send_to_xrpl":
			sender, err := juno.FindAttributeByKey(event, "sender")
			if err != nil {
				return fmt.Errorf("error while getting sender attribute: %s", err)
			}
			recipient, err := juno.FindAttributeByKey(event, "recipient")
			if err != nil {
				return fmt.Errorf("error while getting recipient attribute: %s", err)
			}
			coin, err := juno.FindAttributeByKey(event, "coin")
			if err != nil {
				return fmt.Errorf("error while getting coin attribute: %s", err)
			}

			operationIds, err := m.Source.GetSendToXRPLOperationIDs(recipient.Value, height)
			if err != nil {
				return fmt.Errorf("error while getting operation id: %s", err)
			}

			// check if the operation id already exists
			pendingTx, err := m.db.GetOutgoingPendingTransaction(operationIds)
			if err != nil && !strings.Contains(err.Error(), "sql: no rows in result set") {
				return fmt.Errorf("error while getting pending transaction: %s", err)
			}
			if pendingTx != nil {
				return fmt.Errorf("pending transaction already exists for operation id %v", operationIds)
			}

			err = m.db.SaveOutgoingTransfer(types.NewOutgoingPendingBridgeTransaction(
				tx.TxHash,
				height,
				types.Counterparty_XRPL,
				sender.Value,
				recipient.Value,
				coin.Value,
				types.BridgeTxDir_Outgoing,
				operationIds,
			))
			if err != nil {
				return fmt.Errorf("error while saving coreum to xrpl transaction: %s", err)
			}
		case "save_evidence":
			operationType, err := juno.FindAttributeByKey(event, "operation_type")
			if err != nil {
				// xrpl tp coreum transfer does not have operation type
				if err.Error() != "no attribute with key operation_type found inside event with type wasm" {
					return fmt.Errorf("error while getting operation type attribute: %s", err)
				}
			}

			if operationType.Value != "" && operationType.Value != "coreum_to_xrpl_transfer" {
				continue
			}

			relayerAcc, err := juno.FindAttributeByKey(event, "sender")
			if err != nil {
				return fmt.Errorf("error while getting sender attribute: %s", err)
			}

			threshold_reached, err := juno.FindAttributeByKey(event, "threshold_reached")
			if err != nil {
				return fmt.Errorf("error while getting threshold reached attribute: %s", err)
			}
			threshold_reached_value, err := strconv.ParseBool(threshold_reached.Value)
			if err != nil {
				return fmt.Errorf("error while parsing threshold reached value: %s", err)
			}

			evidence := types.NewBridgeEvidence(
				height,
				tx.TxHash,
				relayerAcc.Value,
				threshold_reached_value,
			)

			if operationType.Value == "coreum_to_xrpl_transfer" {
				// match operation Ids and save them
				operationId, err := juno.FindAttributeByKey(event, "operation_id")
				if err != nil {
					return fmt.Errorf("error while getting operation id attribute: %s", err)
				}

				operationIdInt, err := strconv.ParseUint(operationId.Value, 10, 32)
				if err != nil {
					return fmt.Errorf("error while parsing operation id: %s", err)
				}

				if threshold_reached_value {
					transactionResult, err := juno.FindAttributeByKey(event, "transaction_result")
					if err != nil {
						return fmt.Errorf("error while getting transaction result attribute: %s", err)
					}
					xrplTxHash, err := juno.FindAttributeByKey(event, "tx_hash")
					if err != nil {
						return fmt.Errorf("error while getting tx hash attribute: %s", err)
					}

					err = m.db.SaveOutgoingFinalEvidence(
						evidence,
						uint32(operationIdInt),
						types.BridgeTxResultToStr[transactionResult.Value],
						xrplTxHash.Value,
					)
					if err != nil {
						return fmt.Errorf("error while saving evidence for operation finalization: %s", err)
					}
				} else {
					// TODO: check if the evidences come after finalization
					err = m.db.SaveOutgoingPendingEvidence(
						evidence,
						uint32(operationIdInt),
					)

					if err != nil {
						return fmt.Errorf("error while saving evidence for pending operation: %s", err)
					}
				}

			} else {
				// xrpl to coreum transfer
				xrplHash, err := juno.FindAttributeByKey(event, "hash")
				if err != nil {
					return fmt.Errorf("error while getting hash attribute: %s", err)
				}
				recipient, err := juno.FindAttributeByKey(event, "recipient")
				if err != nil {
					return fmt.Errorf("error while getting recipient attribute: %s", err)
				}

				if threshold_reached_value {
					err = m.db.SaveIncomingFinalTxAndEvidence(
						evidence,
						types.Counterparty_XRPL,
						xrplHash.Value,
						types.BridgeTxResult_ACCEPTED,
					)
					if err != nil {
						return fmt.Errorf("error while saving xrpl to coreum transaction result: %s", err)
					}
				} else {
					issuer, err := juno.FindAttributeByKey(event, "issuer")
					if err != nil {
						return fmt.Errorf("error while getting issuer attribute: %s", err)
					}
					currency, err := juno.FindAttributeByKey(event, "currency")
					if err != nil {
						return fmt.Errorf("error while getting currency attribute: %s", err)
					}
					amount, err := juno.FindAttributeByKey(event, "amount")
					if err != nil {
						return fmt.Errorf("error while getting amount attribute: %s", err)
					}

					err = m.db.SaveIncomingPendingTxAndEvidence(
						types.NewIncomingPendingBridgeTransaction(
							tx.TxHash,
							height,
							types.Counterparty_XRPL,
							xrplHash.Value,
							issuer.Value,
							recipient.Value,
							strings.Join([]string{amount.Value, currency.Value}, ""),
							types.BridgeTxDir_Incoming,
						),
						evidence.Relayer,
					)
					if err != nil {
						return fmt.Errorf("error while saving xrpl to coreum transaction: %s", err)
					}
				}
				if err != nil {
					return fmt.Errorf("error while saving xrpl to coreum transaction evidence: %s", err)
				}
			}
		}

	}
	return nil
}
