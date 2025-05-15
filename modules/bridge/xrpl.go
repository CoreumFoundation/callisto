package bridge

import (
	"fmt"
	"strconv"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	bridgesource "github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/callisto/v4/types"
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

		switch action.Value {
		case "send_to_xrpl":
			if err := h.handleCoreumToXrplPending(event); err != nil {
				return err
			}
		case "save_evidence":
			if err := h.handleSaveEvidence(event); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleCoreumToXrplPending handles the outgoing pending send event
// and saves the outgoing transfer to the database
func (h *XrplMsgHandler) handleCoreumToXrplPending(event abci.Event) error {
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

	operationUniqueId, err := juno.FindAttributeByKey(event, "operation_unique_id")
	if err != nil && err.Error() != "no attribute with key operation_unique_id found inside event with type wasm" {
		return fmt.Errorf("error while getting operation type attribute: %s", err)
	}

	var operationId uint32
	if operationUniqueId.Value == "" {
		// legacy operation id query
		operationId, err = h.Source.GetOutgoingPendingOperationID(h.smartContractAddress, recipient.Value, h.height)
		if err != nil {
			return fmt.Errorf("error while getting operation id: %s", err)
		}
	} else {
		parts := strings.Split(operationUniqueId.Value, "-")
		if len(parts) != 2 {
			return fmt.Errorf("invalid operation unique id format: %s", operationUniqueId.Value)
		}
		operationIdInt, err := strconv.ParseUint(parts[1], 10, 32)
		if err != nil {
			return fmt.Errorf("error while parsing operation id: %s", err)
		}
		operationId = uint32(operationIdInt)
	}

	pendingTx, err := h.db.GetOutgoingPendingTransaction(operationId)
	if err != nil && !strings.Contains(err.Error(), "sql: no rows in result set") {
		return fmt.Errorf("error while getting pending transaction: %s", err)
	}
	if pendingTx != nil {
		return fmt.Errorf("pending transaction already exists for operation id %v", operationId)
	}

	return h.db.SaveOutgoingTransfer(types.NewOutgoingPendingBridgeTransaction(
		h.height,
		h.tx.TxHash,
		types.ChainCoreum,
		types.ChainXRPL,
		sender.Value,
		recipient.Value,
		coin.Value,
		operationId,
	))
}

// handleSaveEvidence handles the save_evidence event
// and saves the evidence to the database
func (h *XrplMsgHandler) handleSaveEvidence(event abci.Event) error {
	operationType, err := juno.FindAttributeByKey(event, "operation_type")
	if err != nil && err.Error() != "no attribute with key operation_type found inside event with type wasm" {
		return fmt.Errorf("error while getting operation type attribute: %s", err)
	}

	if operationType.Value != "" && operationType.Value != "coreum_to_xrpl_transfer" {
		return nil
	}

	relayerAcc, err := juno.FindAttributeByKey(event, "sender")
	if err != nil {
		return fmt.Errorf("error while getting sender attribute: %s", err)
	}

	thresholdReached, err := juno.FindAttributeByKey(event, "threshold_reached")
	if err != nil {
		return fmt.Errorf("error while getting threshold reached attribute: %s", err)
	}
	thresholdReachedValue, err := strconv.ParseBool(thresholdReached.Value)
	if err != nil {
		return fmt.Errorf("error while parsing threshold reached value: %s", err)
	}

	evidence := types.NewBridgeEvidence(
		h.height,
		h.tx.TxHash,
		relayerAcc.Value,
		thresholdReachedValue,
	)

	if operationType.Value == "coreum_to_xrpl_transfer" {
		return h.handleCoreumToXrplEvidence(event, evidence)
	}
	return h.handleXrplToCoreumEvidence(event, evidence)
}

// handleCoreumToXrplEvidence handles the outgoing transfer evidence
// and saves the evidence to the database
func (h *XrplMsgHandler) handleCoreumToXrplEvidence(event abci.Event, evidence types.BridgeEvidence) error {
	operationId, err := juno.FindAttributeByKey(event, "operation_id")
	if err != nil {
		return fmt.Errorf("error while getting operation id attribute: %s", err)
	}

	operationIdInt, err := strconv.ParseUint(operationId.Value, 10, 32)
	if err != nil {
		return fmt.Errorf("error while parsing operation id: %s", err)
	}

	if evidence.ThresholdReached {
		transactionResult, err := juno.FindAttributeByKey(event, "transaction_result")
		if err != nil {
			return fmt.Errorf("error while getting transaction result attribute: %s", err)
		}

		// threshold reached, so the transaction hash of this evidence is the actual payment hash
		xrplTxHash, err := juno.FindAttributeByKey(event, "tx_hash")
		if err != nil {
			return fmt.Errorf("error while getting tx hash attribute: %s", err)
		}

		return h.db.SaveOutgoingFinalEvidence(
			evidence,
			uint32(operationIdInt),
			types.BridgeTxResultToStr[transactionResult.Value],
			xrplTxHash.Value,
		)
	}

	return h.db.SaveOutgoingPendingEvidence(
		evidence,
		uint32(operationIdInt),
	)
}

// handleXrplToCoreumEvidence handles the incoming evidence
// and saves the evidence to the database
func (h *XrplMsgHandler) handleXrplToCoreumEvidence(event abci.Event, evidence types.BridgeEvidence) error {
	xrplHash, err := juno.FindAttributeByKey(event, "hash")
	if err != nil {
		return fmt.Errorf("error while getting hash attribute: %s", err)
	}
	recipient, err := juno.FindAttributeByKey(event, "recipient")
	if err != nil {
		return fmt.Errorf("error while getting recipient attribute: %s", err)
	}

	if evidence.ThresholdReached {
		return h.db.SaveIncomingFinalizedTxAndEvidence(
			evidence,
			types.ChainXRPL,
			xrplHash.Value,
			types.BridgeTxResultAccepted,
		)
	}

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

	return h.db.SaveIncomingPendingTxAndEvidence(
		types.NewIncomingPendingBridgeTransaction(
			xrplHash.Value,
			types.ChainXRPL,
			types.ChainCoreum,
			issuer.Value,
			recipient.Value,
			strings.Join([]string{amount.Value, currency.Value}, ""),
		),
		evidence,
	)
}
