package bridge

import (
	"fmt"
	"strconv"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgesource "github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/callisto/v4/types"
	"github.com/forbole/callisto/v4/utils/events"
	juno "github.com/forbole/juno/v6/types"
)

// XrplMsgHandler is a struct that implements the TxHandler interface
// for handling messages related to the XRPL bridge.
type XrplMsgHandler struct {
	smartContractAddress string
	height               uint64
	msg                  juno.Message
	tx                   *juno.Transaction
	db                   DbHandler
	bridgesource.Source
}

// NewXrplMsgHandler creates a new XrplMsgHandler instance
func NewXrplMsgHandler(smartContractAddress string, height uint64, msg juno.Message, tx *juno.Transaction, db DbHandler, source bridgesource.Source) *XrplMsgHandler {
	return &XrplMsgHandler{
		smartContractAddress: smartContractAddress,
		height:               height,
		msg:                  msg,
		tx:                   tx,
		db:                   db,
		Source:               source,
	}
}

// HandleMsg handles the message for the XrplMsgHandler
func (h *XrplMsgHandler) HandleMsg() error {
	events := juno.FindEventsByType(h.tx.Events, "wasm")
	for _, event := range events {
		action, err := juno.FindAttributeByKey(event, "action")
		if err != nil {
			return fmt.Errorf("error while getting action attribute: %s", err)
		}

		var transaction types.BridgeTransaction
		var evidence types.BridgeEvidence
		switch action.Value {
		case "send_to_xrpl":
			transaction, err = h.extractCoreumToXrplTransaction(event)
			if err != nil {
				return err
			}
		case "save_evidence":
			transaction, evidence, err = h.handleSaveEvidence(event)
			if err != nil {
				return err
			}
		default:
			continue
		}

		transactionId, err := h.db.SaveBridgeTransaction(transaction)
		if err != nil {
			return fmt.Errorf("error while saving transaction: %s", err)
		}

		if evidence.Hash != "" {
			evidence.TransactionId = transactionId
			_, err = h.db.SaveBridgeEvidence(evidence)
			if err != nil {
				return fmt.Errorf("error while saving evidence: %s", err)
			}
		}
	}
	return nil
}

// extractCoreumToXrplTransaction extracts the coreum to xrpl transaction from the event
// It returns the transaction and an error if any
func (h *XrplMsgHandler) extractCoreumToXrplTransaction(event abci.Event) (types.BridgeTransaction, error) {
	sender, err := juno.FindAttributeByKey(event, "sender")
	if err != nil {
		return types.BridgeTransaction{}, fmt.Errorf("error while getting sender attribute: %s", err)
	}
	recipient, err := juno.FindAttributeByKey(event, "recipient")
	if err != nil {
		return types.BridgeTransaction{}, fmt.Errorf("error while getting recipient attribute: %s", err)
	}
	coin, err := juno.FindAttributeByKey(event, "coin")
	if err != nil {
		return types.BridgeTransaction{}, fmt.Errorf("error while getting coin attribute: %s", err)
	}

	operationUniqueIdAttr, err := juno.FindAttributeByKey(event, "operation_unique_id")
	if err != nil && err.Error() != events.JunoAttributeNotFoundError("operation_unique_id", event) {
		return types.BridgeTransaction{}, fmt.Errorf("error while getting operation type attribute: %s", err)
	}
	operationUniqueId := operationUniqueIdAttr.Value
	if operationUniqueId == "" {
		// legacy operation id query
		operationId, err := h.Source.GetOutgoingPendingOperationID(h.smartContractAddress, recipient.Value, h.height)
		if err != nil {
			return types.BridgeTransaction{}, fmt.Errorf("error while getting operation id: %s", err)
		}

		operationUniqueId = strconv.FormatUint(uint64(operationId), 10)
	}

	pendingTx, err := h.db.GetBridgeTransaction(operationUniqueId)
	if err != nil && !strings.Contains(err.Error(), "sql: no rows in result set") {
		return types.BridgeTransaction{}, fmt.Errorf("error while getting pending transaction: %s", err)
	}
	if pendingTx.ID != 0 {
		return types.BridgeTransaction{}, fmt.Errorf("pending transaction already exists for operation unique id %v", operationUniqueId)
	}

	parsedCoin, err := sdk.ParseCoinNormalized(coin.Value)
	if err != nil {
		return types.BridgeTransaction{}, fmt.Errorf("error while parsing coins: %s", err)
	}

	heightInt64 := int64(h.height)
	transaction := types.NewBridgeTransaction(
		&operationUniqueId,
		&heightInt64,
		h.tx.TxHash,
		types.ChainCoreum,
		types.ChainXRPL,
		nil,
		&sender.Value,
		recipient.Value,
		parsedCoin.Denom,
		parsedCoin.Amount.String(),
	)
	return transaction, nil
}

// handleSaveEvidence handles the save evidence event and returns the transaction and evidence
// If the event is not a save evidence event, it returns an empty transaction and evidence
func (h *XrplMsgHandler) handleSaveEvidence(event abci.Event) (types.BridgeTransaction, types.BridgeEvidence, error) {
	operationType, err := juno.FindAttributeByKey(event, "operation_type")
	if err != nil && err.Error() != events.JunoAttributeNotFoundError("operation_type", event) {
		return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting operation type attribute: %s", err)
	}

	if operationType.Value != "" && operationType.Value != "coreum_to_xrpl_transfer" {
		return types.BridgeTransaction{}, types.BridgeEvidence{}, nil
	}

	relayerAcc, err := juno.FindAttributeByKey(event, "sender")
	if err != nil {
		return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting sender attribute: %s", err)
	}

	thresholdReached, err := juno.FindAttributeByKey(event, "threshold_reached")
	if err != nil {
		return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting threshold reached attribute: %s", err)
	}
	thresholdReachedValue, err := strconv.ParseBool(thresholdReached.Value)
	if err != nil {
		return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while parsing threshold reached value: %s", err)
	}

	evidence := types.NewBridgeEvidence(
		h.height,
		h.tx.TxHash,
		relayerAcc.Value,
		thresholdReachedValue,
	)

	if operationType.Value == "coreum_to_xrpl_transfer" {
		operationUniqueIdAttr, err := juno.FindAttributeByKey(event, "operation_unique_id")
		if err != nil && err.Error() != events.JunoAttributeNotFoundError("operation_unique_id", event) {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting operation type attribute in coreum_to_xrpl_transfer: %s", err)
		}
		operationUniqueId := operationUniqueIdAttr.Value
		if operationUniqueId == "" {
			operationIdAttr, err := juno.FindAttributeByKey(event, "operation_id")
			if err != nil {
				return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting operation id in coreum_to_xrpl_transfer: %s", err)
			}
			operationUniqueId = operationIdAttr.Value
		}

		transaction, err := h.db.GetBridgeTransaction(operationUniqueId)
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting bridge transaction: %s", err)
		}

		if evidence.ThresholdReached {
			transactionResult, err := juno.FindAttributeByKey(event, "transaction_result")
			if err != nil {
				return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting transaction result attribute: %s", err)
			}

			// threshold reached, so the transaction hash of this evidence is the actual payment hash
			xrplTxHash, err := juno.FindAttributeByKey(event, "tx_hash")
			if err != nil {
				return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting tx hash attribute: %s", err)
			}

			evidence.SetFinalProps(xrplTxHash.Value, types.BridgeTxResultToStr[transactionResult.Value])
		}
		return transaction, evidence, nil
	} else {
		xrplTxHash, err := juno.FindAttributeByKey(event, "hash")
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting hash attribute: %s", err)
		}
		recipient, err := juno.FindAttributeByKey(event, "recipient")
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting recipient attribute: %s", err)
		}

		issuer, err := juno.FindAttributeByKey(event, "issuer")
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting issuer attribute: %s", err)
		}
		currency, err := juno.FindAttributeByKey(event, "currency")
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting currency attribute: %s", err)
		}
		amount, err := juno.FindAttributeByKey(event, "amount")
		if err != nil {
			return types.BridgeTransaction{}, types.BridgeEvidence{}, fmt.Errorf("error while getting amount attribute: %s", err)
		}

		transaction := types.NewBridgeTransaction(
			nil,
			nil,
			xrplTxHash.Value,
			types.ChainXRPL,
			types.ChainCoreum,
			&issuer.Value,
			nil,
			recipient.Value,
			currency.Value,
			amount.Value,
		)

		if evidence.ThresholdReached {
			evidence.SetFinalProps(xrplTxHash.Value, types.BridgeTxResultAccepted)
		}

		return transaction, evidence, nil
	}

}
