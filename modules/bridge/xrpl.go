package bridge

import (
	"fmt"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	relayertypes "github.com/CoreumFoundation/coreumbridge-xrpl/relayer/coreum"

	bridgesource "github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/callisto/v4/types"
	eventsutil "github.com/forbole/callisto/v4/utils/events"
)

const (
	OperationTypeCoreumToXrplTransfer = "coreum_to_xrpl_transfer"
)

const (
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
	operationUniqueIDAttribute = "operation_unique_id"
)

// XrplMsgHandler is a struct that implements the TxHandler interface
// for handling messages related to the XRPL bridge.
type XrplMsgHandler struct {
	smartContractAddress string
	height               uint64
	txHash               string
	msgIndex             int
	msgEvents            sdk.StringEvents
	db                   DbHandler
	bridgesource.Source
}

// NewXrplMsgHandler creates a new XrplMsgHandler instance
func NewXrplMsgHandler(smartContractAddress string, height uint64, txHash string, msgIndex int, msgEvents sdk.StringEvents, db DbHandler, source bridgesource.Source) *XrplMsgHandler {
	return &XrplMsgHandler{
		smartContractAddress: smartContractAddress,
		height:               height,
		txHash:               txHash,
		msgIndex:             msgIndex,
		msgEvents:            msgEvents,
		db:                   db,
		Source:               source,
	}
}

// HandleMsg handles the message for the XrplMsgHandler
func (h *XrplMsgHandler) HandleMsg() error {
	for _, msgEvent := range h.msgEvents {
		if msgEvent.Type != wasmtypes.WasmModuleEventType {
			continue
		}

		for _, attr := range msgEvent.Attributes {
			if attr.Key == sdk.AttributeKeyAction {
				switch attr.Value {
				case string(relayertypes.ExecSendToXRPL):
					if err := h.handleCoreumToXrplEvent(msgEvent); err != nil {
						return fmt.Errorf("error while extracting coreum to xrpl transaction: %w", err)
					}
				case string(relayertypes.ExecMethodSaveEvidence):
					if err := h.handleSaveEvidenceEvent(msgEvent); err != nil {
						return fmt.Errorf("error while handling save evidence: %w", err)
					}
				}
			}
		}

	}

	return nil
}

// handleCoreumToXrplEvent extracts the coreum to xrpl transaction from the event
// It returns the transaction and an error if any
func (h *XrplMsgHandler) handleCoreumToXrplEvent(event sdk.StringEvent) error {
	parsedEvent, err := eventsutil.FindEventMap(event, []string{
		sdk.AttributeKeySender,
		recipientAttribute,
		coinAttribute,
	}, []string{
		operationUniqueIDAttribute,
	})
	if err != nil {
		return err
	}

	operationUniqueID := parsedEvent[operationUniqueIDAttribute]
	if operationUniqueID == "" {
		// legacy operation id query
		operationId, err := h.Source.GetOutgoingPendingOperationSequence(h.smartContractAddress, parsedEvent[recipientAttribute], h.height)
		if err != nil {
			return err
		}
		operationIdStr := strconv.FormatUint(uint64(operationId), 10)
		operationUniqueID = operationIdStr
	}

	parsedCoin, err := sdk.ParseCoinNormalized(parsedEvent[coinAttribute])
	if err != nil {
		return fmt.Errorf("error while parsing coins: %s", err)
	}

	heightInt64 := int64(h.height)
	parsedSender := parsedEvent[sdk.AttributeKeySender]
	transaction := types.NewBridgeTransaction(
		&operationUniqueID,
		&heightInt64,
		h.txHash,
		types.ChainCoreum,
		types.ChainXRPL,
		&parsedSender,
		parsedEvent[recipientAttribute],
		parsedCoin.Denom,
		parsedCoin.Amount.String(),
	)

	_, err = h.db.SaveBridgeTransaction(&transaction)
	if err != nil {
		return fmt.Errorf("error while saving transaction: %s", err)
	}

	return nil
}

// handleSaveEvidenceEvent handles the save evidence event and returns the transaction and evidence
// If the event is not a save evidence event, it returns an empty transaction and evidence
func (h *XrplMsgHandler) handleSaveEvidenceEvent(event sdk.StringEvent) error {
	operationType, found := eventsutil.FindAttributeByKey(event, operationTypeAttribute)
	if found && operationType.Value != OperationTypeCoreumToXrplTransfer {
		return nil
	}

	evt, err := eventsutil.FindEventMap(event, []string{
		sdk.AttributeKeySender,
		thresholdReachedAttribute,
	}, []string{})
	if err != nil {
		return err
	}

	thresholdReachedValue, err := strconv.ParseBool(evt[thresholdReachedAttribute])
	if err != nil {
		return err
	}

	var transaction types.BridgeTransaction
	evidence := types.NewBridgeEvidence(
		h.height,
		h.txHash,
		h.msgIndex,
		evt[sdk.AttributeKeySender],
		thresholdReachedValue,
	)

	if operationType.Value == OperationTypeCoreumToXrplTransfer {
		toXrpl, err := eventsutil.FindEventMap(event, []string{
			operationUniqueIDAttribute,
			operationIdAttribute,
		}, []string{})
		if err != nil {
			return err
		}

		operationUniqueID := toXrpl[operationUniqueIDAttribute]
		if operationUniqueID == "" {
			operationID, ok := toXrpl[operationIdAttribute]
			if !ok {
				return fmt.Errorf("nor operation id nor operation unique id found")
			}
			operationUniqueID = operationID
		}

		transaction, err = h.db.GetBridgeTransaction(operationUniqueID)
		if err != nil {
			return err
		}

		if evidence.ThresholdReached {
			parsedThresholdReachedEvent, err := eventsutil.FindEventMap(event, []string{
				transactionResultAttribute,
				hashAttribute,
			}, []string{})
			if err != nil {
				return err
			}

			transactionResult := parsedThresholdReachedEvent[transactionResultAttribute]
			evidence.SetFinalProps(parsedThresholdReachedEvent[hashAttribute], types.BridgeTxResultToStr[transactionResult])
		}

	} else {
		toCoreum, err := eventsutil.FindEventMap(event, []string{
			hashAttribute,
			recipientAttribute,
			issuerAttribute,
			currencyAttribute,
			amountAttribute,
		}, []string{})
		if err != nil {
			return err
		}

		// concat the issuer and currency to create the denom alias,
		// this will store the issuer and currency in the denom field
		denom := toCoreum[issuerAttribute] + "-" + toCoreum[currencyAttribute]

		transaction = types.NewBridgeTransaction(
			nil,
			nil,
			toCoreum[hashAttribute],
			types.ChainXRPL,
			types.ChainCoreum,
			nil,
			toCoreum[recipientAttribute],
			denom,
			toCoreum[amountAttribute],
		)

		if evidence.ThresholdReached {
			evidence.SetFinalProps(toCoreum[hashAttribute], types.BridgeTxResultAccepted)
		}
	}

	transactionId, err := h.db.SaveBridgeTransaction(&transaction)
	if err != nil {
		return fmt.Errorf("error while saving transaction: %s", err)
	}

	evidence.TransactionId = transactionId
	_, err = h.db.SaveBridgeEvidence(&evidence)
	if err != nil {
		return fmt.Errorf("error while saving evidence: %s", err)
	}

	return nil
}
