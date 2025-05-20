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

const (
	senderAttribute            = "sender"
	recipientAttribute         = "recipient"
	coinAttribute              = "coin"
	operationTypeAttribute     = "operation_type"
	thresholdReachedAttribute  = "threshold_reached"
	transactionResultAttribute = "transaction_result"
	txHashAttribute            = "tx_hash"
	hashAttribute              = "hash"
	issuerAttribute            = "issuer"
	currencyAttribute          = "currency"
	amountAttribute            = "amount"
	operationIdAttribute       = "operation_id"
	operationUniqueIdAttribute = "operation_unique_id"
)

const (
	OperationTypeCoreumToXrplTransfer = "coreum_to_xrpl_transfer"
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
	wasmEvents := juno.FindEventsByType(h.tx.Events, "wasm")
	for _, event := range wasmEvents {
		action, err := juno.FindAttributeByKey(event, "action")
		if err != nil {
			return fmt.Errorf("error while getting action attribute: %s", err)
		}

		var transaction *types.BridgeTransaction
		var evidence *types.BridgeEvidence
		switch action.Value {
		case "send_to_xrpl":
			transaction, err = h.extractCoreumToXrplTransaction(event)
			if err != nil {
				return fmt.Errorf("error while extracting coreum to xrpl transaction: %s", err)
			}
		case "save_evidence":
			transaction, evidence, err = h.extractXrplToCoreumTransactionAndEvidence(event)
			if err != nil {
				return fmt.Errorf("error while handling save evidence: %s", err)
			}
		default:
			continue
		}

		// if the transaction is nil, it means that the event is not relevant
		// it happens if the event is save_evidence and operation_type is not
		// coreum_to_xrpl_transfer or empty which is xrpl to coreum evidence.
		if transaction == nil {
			continue
		}

		transactionId, err := h.db.SaveBridgeTransaction(transaction)
		if err != nil {
			return fmt.Errorf("error while saving transaction: %s", err)
		}

		if evidence != nil {
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
func (h *XrplMsgHandler) extractCoreumToXrplTransaction(event abci.Event) (*types.BridgeTransaction, error) {
	sender, err := juno.FindAttributeByKey(event, senderAttribute)
	if err != nil {
		return nil, err
	}
	recipient, err := juno.FindAttributeByKey(event, recipientAttribute)
	if err != nil {
		return nil, err
	}
	coin, err := juno.FindAttributeByKey(event, coinAttribute)
	if err != nil {
		return nil, err
	}

	operationUniqueIdAttr, err := juno.FindAttributeByKey(event, operationUniqueIdAttribute)
	if err != nil && err.Error() != events.JunoAttributeNotFoundError(operationUniqueIdAttribute, event) {
		return nil, err
	}
	operationUniqueID := operationUniqueIdAttr.Value
	if operationUniqueID == "" {
		// legacy operation id query
		operationId, err := h.Source.GetOutgoingPendingOperationSequence(h.smartContractAddress, recipient.Value, h.height)
		if err != nil {
			return nil, err
		}

		operationUniqueID = strconv.FormatUint(uint64(operationId), 10)
	}

	pendingTx, err := h.db.GetBridgeTransaction(operationUniqueID)
	if err != nil && !strings.Contains(err.Error(), "sql: no rows in result set") {
		return nil, fmt.Errorf("error while getting pending transaction: %s", err)
	}
	if pendingTx.ID != 0 {
		return nil, fmt.Errorf("pending transaction already exists for operation unique id %v", operationUniqueID)
	}

	parsedCoin, err := sdk.ParseCoinNormalized(coin.Value)
	if err != nil {
		return nil, fmt.Errorf("error while parsing coins: %s", err)
	}

	heightInt64 := int64(h.height)
	transaction := types.NewBridgeTransaction(
		&operationUniqueID,
		&heightInt64,
		h.tx.TxHash,
		types.ChainCoreum,
		types.ChainXRPL,
		&sender.Value,
		recipient.Value,
		parsedCoin.Denom,
		parsedCoin.Amount.String(),
	)
	return &transaction, nil
}

// extractXrplToCoreumTransactionAndEvidence handles the save evidence event and returns the transaction and evidence
// If the event is not a save evidence event, it returns an empty transaction and evidence
func (h *XrplMsgHandler) extractXrplToCoreumTransactionAndEvidence(event abci.Event) (*types.BridgeTransaction, *types.BridgeEvidence, error) {
	operationType, err := juno.FindAttributeByKey(event, operationTypeAttribute)
	if err != nil && err.Error() != events.JunoAttributeNotFoundError(operationTypeAttribute, event) {
		return nil, nil, err
	}

	if operationType.Value != "" && operationType.Value != OperationTypeCoreumToXrplTransfer {
		return nil, nil, nil
	}

	relayerAcc, err := juno.FindAttributeByKey(event, senderAttribute)
	if err != nil {
		return nil, nil, err
	}

	thresholdReached, err := juno.FindAttributeByKey(event, thresholdReachedAttribute)
	if err != nil {
		return nil, nil, err
	}
	thresholdReachedValue, err := strconv.ParseBool(thresholdReached.Value)
	if err != nil {
		return nil, nil, err
	}

	evidence := types.NewBridgeEvidence(
		h.height,
		h.tx.TxHash,
		relayerAcc.Value,
		thresholdReachedValue,
	)

	if operationType.Value == OperationTypeCoreumToXrplTransfer {
		operationUniqueIdAttr, err := juno.FindAttributeByKey(event, operationUniqueIdAttribute)
		if err != nil && err.Error() != events.JunoAttributeNotFoundError(operationUniqueIdAttribute, event) {
			return nil, nil, err
		}
		operationUniqueId := operationUniqueIdAttr.Value
		if operationUniqueId == "" {
			operationIdAttr, err := juno.FindAttributeByKey(event, operationIdAttribute)
			if err != nil {
				return nil, nil, err
			}
			operationUniqueId = operationIdAttr.Value
		}

		transaction, err := h.db.GetBridgeTransaction(operationUniqueId)
		if err != nil {
			return nil, nil, err
		}

		if evidence.ThresholdReached {
			transactionResult, err := juno.FindAttributeByKey(event, transactionResultAttribute)
			if err != nil {
				return nil, nil, err
			}

			// threshold reached, so the transaction hash of this evidence is the actual payment hash
			xrplTxHash, err := juno.FindAttributeByKey(event, txHashAttribute)
			if err != nil {
				return nil, nil, err
			}

			evidence.SetFinalProps(xrplTxHash.Value, types.BridgeTxResultToStr[transactionResult.Value])
		}
		return &transaction, &evidence, nil
	} else {
		xrplTxHash, err := juno.FindAttributeByKey(event, txHashAttribute)
		if err != nil {
			return nil, nil, err
		}
		recipient, err := juno.FindAttributeByKey(event, recipientAttribute)
		if err != nil {
			return nil, nil, err
		}

		issuer, err := juno.FindAttributeByKey(event, issuerAttribute)
		if err != nil {
			return nil, nil, err
		}
		currency, err := juno.FindAttributeByKey(event, currencyAttribute)
		if err != nil {
			return nil, nil, err
		}
		amount, err := juno.FindAttributeByKey(event, amountAttribute)
		if err != nil {
			return nil, nil, err
		}

		// concat the issuer and currency to create the denom alias,
		// this will store the issuer and currency in the denom field
		denom := issuer.Value + "-" + currency.Value

		transaction := types.NewBridgeTransaction(
			nil,
			nil,
			xrplTxHash.Value,
			types.ChainXRPL,
			types.ChainCoreum,
			nil,
			recipient.Value,
			denom,
			amount.Value,
		)

		if evidence.ThresholdReached {
			evidence.SetFinalProps(xrplTxHash.Value, types.BridgeTxResultAccepted)
		}

		return &transaction, &evidence, nil
	}
}
