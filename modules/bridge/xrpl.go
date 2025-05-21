package bridge

import (
	"fmt"
	"strconv"

	"github.com/samber/lo"

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

// XrplMsgHandler implements the TxHandler interface for XRPL bridge messages.
type XrplMsgHandler struct {
	smartContractAddress string
	height               uint64
	txHash               string
	msgIndex             int
	msgEvents            sdk.StringEvents
	db                   DbHandler
	bridgesource.Source
}

// NewXrplMsgHandler returns a new XrplMsgHandler instance.
func NewXrplMsgHandler(
	smartContractAddress string,
	height uint64,
	txHash string,
	msgIndex int,
	msgEvents sdk.StringEvents,
	db DbHandler,
	source bridgesource.Source,
) *XrplMsgHandler {
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

// HandleMsg processes all relevant events for the XRPL bridge.
func (h *XrplMsgHandler) HandleMsg() error {
	for _, msgEvent := range h.msgEvents {
		if msgEvent.Type != wasmtypes.WasmModuleEventType {
			continue
		}
		actionAttr, found := eventsutil.FindAttributeByKey(msgEvent, sdk.AttributeKeyAction)
		if !found {
			continue
		}
		switch actionAttr.Value {
		case string(relayertypes.ExecSendToXRPL):
			if err := h.handleCoreumToXrplEvent(msgEvent); err != nil {
				return fmt.Errorf("extracting coreum to xrpl transaction: %w", err)
			}
		case string(relayertypes.ExecMethodSaveEvidence):
			if err := h.handleSaveEvidenceEvent(msgEvent); err != nil {
				return fmt.Errorf("handling save evidence: %w", err)
			}
		}
	}
	return nil
}

// handleCoreumToXrplEvent extracts and saves a Coreum-to-XRPL transaction from the event.
func (h *XrplMsgHandler) handleCoreumToXrplEvent(event sdk.StringEvent) error {
	parsedEvent, err := eventsutil.BuildAttributesMap(event,
		[]string{sdk.AttributeKeySender, recipientAttribute, coinAttribute},
		[]string{operationUniqueIDAttribute},
	)
	if err != nil {
		return err
	}

	operationUniqueID := parsedEvent[operationUniqueIDAttribute]
	if operationUniqueID == "" {
		operationID, err := h.Source.GetOutgoingPendingOperationSequence(
			h.smartContractAddress, parsedEvent[recipientAttribute], h.height,
		)
		if err != nil {
			return err
		}
		operationUniqueID = strconv.FormatUint(uint64(operationID), 10)
	}

	parsedCoin, err := sdk.ParseCoinNormalized(parsedEvent[coinAttribute])
	if err != nil {
		return fmt.Errorf("parsing coins: %w", err)
	}

	transaction := types.NewBridgeTransaction(
		&operationUniqueID,
		lo.ToPtr(int64(h.height)),
		h.txHash,
		types.ChainCoreum,
		types.ChainXRPL,
		lo.ToPtr(parsedEvent[sdk.AttributeKeySender]),
		parsedEvent[recipientAttribute],
		parsedCoin.Denom,
		parsedCoin.Amount.String(),
	)

	if _, err := h.db.SaveBridgeTransaction(&transaction); err != nil {
		return fmt.Errorf("saving transaction: %w", err)
	}
	return nil
}

// handleSaveEvidenceEvent processes and saves bridge evidence from the event.
func (h *XrplMsgHandler) handleSaveEvidenceEvent(event sdk.StringEvent) error {
	operationType, found := eventsutil.FindAttributeByKey(event, operationTypeAttribute)
	if found && operationType.Value != OperationTypeCoreumToXrplTransfer {
		return nil
	}

	evt, err := eventsutil.BuildAttributesMap(event,
		[]string{sdk.AttributeKeySender, thresholdReachedAttribute},
		nil,
	)
	if err != nil {
		return err
	}

	thresholdReached, err := strconv.ParseBool(evt[thresholdReachedAttribute])
	if err != nil {
		return err
	}

	evidence := types.NewBridgeEvidence(
		h.height,
		h.txHash,
		h.msgIndex,
		evt[sdk.AttributeKeySender],
		thresholdReached,
	)

	var transaction types.BridgeTransaction

	if operationType.Value == OperationTypeCoreumToXrplTransfer {
		toXrpl, err := eventsutil.BuildAttributesMap(event,
			nil,
			[]string{operationUniqueIDAttribute, operationIdAttribute},
		)
		if err != nil {
			return err
		}

		operationUniqueID := toXrpl[operationUniqueIDAttribute]
		if operationUniqueID == "" {
			operationUniqueID = toXrpl[operationIdAttribute]
			if operationUniqueID == "" {
				return fmt.Errorf("neither operation id nor operation unique id found")
			}
		}

		transaction, err = h.db.GetBridgeTransaction(operationUniqueID)
		if err != nil {
			return err
		}

		if evidence.ThresholdReached {
			parsed, err := eventsutil.BuildAttributesMap(event,
				[]string{transactionResultAttribute, txHashAttribute},
				nil,
			)
			if err != nil {
				return err
			}
			// the actual payment or rejection happens when the transaction threshold is reached
			// so we store the result of the transaction whether it was accepted or rejected
			// this hash is xrpl transaction hash
			evidence.SetFinalProps(parsed[txHashAttribute], types.BridgeTxResultToStr[parsed[transactionResultAttribute]])
		}
	} else {
		toCoreum, err := eventsutil.BuildAttributesMap(event,
			[]string{hashAttribute, recipientAttribute, issuerAttribute, currencyAttribute, amountAttribute},
			nil,
		)
		if err != nil {
			return err
		}

		denom := fmt.Sprintf("%s-%s", toCoreum[issuerAttribute], toCoreum[currencyAttribute])
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
			// the actual payment happens when the transaction threshold is reached
			// this happens if the contract minted or sent the token to the recipient.
			// this hash is coreum transaction hash
			evidence.SetFinalProps(h.txHash, types.BridgeTxResultAccepted)
		}
	}

	transactionID, err := h.db.SaveBridgeTransaction(&transaction)
	if err != nil {
		return fmt.Errorf("saving transaction: %w", err)
	}

	evidence.TransactionId = transactionID
	if _, err := h.db.SaveBridgeEvidence(&evidence); err != nil {
		return fmt.Errorf("saving evidence: %w", err)
	}

	return nil
}
