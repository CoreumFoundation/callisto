package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// BridgeTxResult is the result of the bridge transaction.
// It can be either accepted, rejected, or invalid.
type BridgeTxResult string

const (
	BridgeTxResultPending  BridgeTxResult = "pending"
	BridgeTxResultAccepted BridgeTxResult = "transaction_accepted"
	BridgeTxResultRejected BridgeTxResult = "transaction_rejected"
	BridgeTxResultInvalid  BridgeTxResult = "transaction_invalid"
)

var BridgeTxResultToStr = map[string]BridgeTxResult{
	"pending":              BridgeTxResultPending,
	"transaction_accepted": BridgeTxResultAccepted,
	"transaction_rejected": BridgeTxResultRejected,
	"transaction_invalid":  BridgeTxResultInvalid,
}

// Chain is the chain of the bridge transaction.
type Chain string

const (
	ChainUnknown Chain = "UNKNOWN"
	ChainCoreum  Chain = "Coreum"
	ChainXRPL    Chain = "XRPL"
)

var StrToChain = map[string]Chain{
	"UNKNOWN": ChainUnknown,
	"XRPL":    ChainXRPL,
	"Coreum":  ChainCoreum,
}

// BridgeTransaction is the structure of the bridge transaction.
type BridgeTransaction struct {
	// ID is the auto-generated serial ID of the transaction.
	ID int64 `json:"id"`
	// OperationUniqueID is the operation unique ID of the transaction (it might be null if there are no pending operations).
	OperationUniqueID *string `json:"operation_unique_id"`
	// Height is the height of the transaction when it is originated.
	Height *int64 `json:"height"`
	// UserInitiatedHash is the hash of the transaction when it is originated.
	UserInitiatedHash string `json:"user_initiated_hash"`
	// SourceChain is the source chain of the transfer origin.
	SourceChain Chain `json:"source_chain"`
	// DestinationChain is the destination chain of the transfer.
	DestinationChain Chain `json:"destination_chain"`
	// Sender is the sender address of the transfer.
	Sender *string `json:"sender"`
	// Recipient is the recipient address of the transfer.
	Recipient string `json:"recipient"`
	// Denom is the denomination of the amount.
	Denom string `json:"denom"`
	// Amount is the amount of the transfer.
	Amount string `json:"amount"`
}

// NewBridgeTransaction creates a bridge transaction.
func NewBridgeTransaction(
	operationUniqueID *string,
	height *int64,
	userInitiatedHash string,
	sourceChain, destinationChain Chain,
	sender *string, recipient, denom, amount string,
) BridgeTransaction {
	return BridgeTransaction{
		OperationUniqueID: operationUniqueID,
		Height:            height,
		UserInitiatedHash: userInitiatedHash,
		SourceChain:       sourceChain,
		DestinationChain:  destinationChain,
		Sender:            sender,
		Recipient:         recipient,
		Denom:             denom,
		Amount:            amount,
	}
}

// BridgeEvidence is the structure of the bridge evidence.
type BridgeEvidence struct {
	// ID is the auto-generated serial ID of the evidence.
	ID int64 `json:"id"`
	// TransactionId is the ID of the transaction.
	TransactionId int64 `json:"transaction_id"`
	// Height is the height of the evidence transaction.
	Height int64 `json:"height"`
	// Hash is the hash of the evidence transaction.
	Hash string `json:"hash"`
	// RelayerAddress is the address of the relayer.
	RelayerAddress string `json:"relayer_address"`
	// ThresholdReached is the flag indicating whether the threshold is reached which means transfer is finalized.
	ThresholdReached bool `json:"threshold_reached"`
	// SettlementHash is the hash of the actual fund transfer transaction.
	SettlementHash *string `json:"settlement_hash"`
	// Result is the result of the transaction.
	Result BridgeTxResult `json:"result"`
}

// NewBridgeEvidence creates a new bridge evidence.
func NewBridgeEvidence(
	height uint64,
	hash string,
	relayerAddress string,
	thresholdReached bool,
) BridgeEvidence {
	return BridgeEvidence{
		Height:           int64(height),
		Hash:             hash,
		RelayerAddress:   relayerAddress,
		ThresholdReached: thresholdReached,
		Result:           BridgeTxResultPending,
	}
}

func (e *BridgeEvidence) SetFinalProps(settlementHash string, result BridgeTxResult) {
	e.SettlementHash = &settlementHash
	e.Result = result
}

// *** the following code is a copy of the original code from the bridge relayer ***

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

// GetOperationSequence returns operation sequence.
func (o Operation) GetOperationSequence() uint32 {
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
