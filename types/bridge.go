package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BridgeTxDir string

const (
	BridgeTxDir_UNKNOWN  BridgeTxDir = "UNKNOWN"
	BridgeTxDir_Outgoing BridgeTxDir = "Outgoing"
	BridgeTxDir_Incoming BridgeTxDir = "Incoming"
)

var StrToBridgeTxDir = map[string]BridgeTxDir{
	"UNKNOWN":  BridgeTxDir_UNKNOWN,
	"Outgoing": BridgeTxDir_Outgoing,
	"Incoming": BridgeTxDir_Incoming,
}

type BridgeTxResult string

const (
	BridgeTxResult_UNKNOWN  BridgeTxResult = "UNKNOWN"
	BridgeTxResult_ACCEPTED BridgeTxResult = "transaction_accepted"
	BridgeTxResult_REJECTED BridgeTxResult = "transaction_rejected"
	BridgeTxResult_INVALID  BridgeTxResult = "transaction_invalid"
)

var BridgeTxResultToStr = map[string]BridgeTxResult{
	"UNKNOWN":              BridgeTxResult_UNKNOWN,
	"transaction_accepted": BridgeTxResult_ACCEPTED,
	"transaction_rejected": BridgeTxResult_REJECTED,
	"transaction_invalid":  BridgeTxResult_INVALID,
}

type Counterparty string

const (
	Counterparty_UNKNOWN Counterparty = "UNKNOWN"
	Counterparty_XRPL    Counterparty = "XRPL"
)

var StrToCounterparty = map[string]Counterparty{
	"UNKNOWN": Counterparty_UNKNOWN,
	"XRPL":    Counterparty_XRPL,
}

type BridgeTransaction struct {
	ID               int64          `json:"id"`
	InitHeight       int64          `json:"init_height"`
	FinalHeight      int64          `json:"final_height"`
	InitHash         string         `json:"init_hash"`
	FinalHash        string         `json:"final_hash"`
	Counterparty     Counterparty   `json:"counterparty"`
	CounterpartyHash string         `json:"counterparty_hash"`
	Sender           string         `json:"sender"`
	Recipient        string         `json:"recipient"`
	Amount           string         `json:"amount"`
	Direction        BridgeTxDir    `json:"direction"`
	Result           BridgeTxResult `json:"result"`
	OperationIDs     []uint32       `json:"operation_ids"`
}

func NewOutgoingPendingBridgeTransaction(txHash string, height uint64, counterparty Counterparty, sender, recipient, amount string, direction BridgeTxDir, operationIDs []uint32) BridgeTransaction {
	return BridgeTransaction{
		InitHash:     txHash,
		InitHeight:   int64(height),
		Counterparty: counterparty,
		Sender:       sender,
		Recipient:    recipient,
		Amount:       amount,
		Direction:    direction,
		OperationIDs: operationIDs,
	}
}

func NewIncomingPendingBridgeTransaction(txHash string, height uint64, counterparty Counterparty, counterpartyHash, sender, recipient, amount string, direction BridgeTxDir) BridgeTransaction {
	return BridgeTransaction{
		InitHash:         txHash,
		InitHeight:       int64(height),
		Counterparty:     counterparty,
		CounterpartyHash: counterpartyHash,
		Sender:           sender,
		Recipient:        recipient,
		Amount:           amount,
		Direction:        direction,
	}
}

type BridgeEvidence struct {
	ID               int64  `json:"id"`
	BridgeTxID       int64  `json:"tx_id"`
	Height           int64  `json:"height"`
	Hash             string `json:"hash"`
	Relayer          string `json:"relayer"`
	ThresholdReached bool   `json:"threshold_reached"`
}

func NewBridgeEvidence(height uint64, hash string, relayer string, thresholdReached bool) BridgeEvidence {
	return BridgeEvidence{
		Height:           int64(height),
		Hash:             hash,
		Relayer:          relayer,
		ThresholdReached: thresholdReached,
	}
}

type Source interface {
	GetSendToXRPLOperationID(
		recipient string,
		height uint64,
	) (uint32, error)
}

// QueryMethod is contract query method.
type QueryMethod string

const QueryMethodPendingOperations QueryMethod = "pending_operations"

var Limit uint32 = 50

// Signature is a pair of the relayer provided the signature and signature string.
type Signature struct {
	RelayerCoreumAddress sdk.AccAddress `json:"relayer_coreum_address"`
	Signature            string         `json:"signature"`
}

// OperationTypeCoreumToXRPLTransfer is coreum to XRPL transfer operation type.
type OperationTypeCoreumToXRPLTransfer struct {
	Issuer    string       `json:"issuer"`
	Currency  string       `json:"currency"`
	Amount    sdkmath.Int  `json:"amount"`
	MaxAmount *sdkmath.Int `json:"max_amount,omitempty"`
	Recipient string       `json:"recipient"`
}

// OperationType is operation type.
type OperationType struct {
	CoreumToXRPLTransfer *OperationTypeCoreumToXRPLTransfer `json:"coreum_to_xrpl_transfer,omitempty"`
}

// Operation is contract operation which should be signed and executed.
type Operation struct {
	Version         uint32        `json:"version"`
	TicketSequence  uint32        `json:"ticket_sequence"`
	AccountSequence uint32        `json:"account_sequence"`
	Signatures      []Signature   `json:"signatures"`
	OperationType   OperationType `json:"operation_type"`
	XRPLBaseFee     uint32        `json:"xrpl_base_fee"`
}

// GetOperationID returns operation ID.
func (o Operation) GetOperationID() uint32 {
	if o.TicketSequence != 0 {
		return o.TicketSequence
	}

	return o.AccountSequence
}

type PagingUint32KeyRequest struct {
	StartAfterKey *uint32 `json:"start_after_key,omitempty"`
	Limit         *uint32 `json:"limit,omitempty"`
}
type PendingOperationsResponse struct {
	LastKey    uint32      `json:"last_key"`
	Operations []Operation `json:"operations"`
}
