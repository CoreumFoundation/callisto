package database

import (
	"github.com/forbole/callisto/v4/types"
)

// SaveBridgeTransaction saves the bridge transaction to the database.
// It returns the ID of the transaction if it was inserted, or the existing ID if it was already present.
func (db *Db) SaveBridgeTransaction(tx *types.BridgeTransaction) (int64, error) {
	stmt := `
	INSERT INTO bridge_transaction (operation_unique_id, height, user_initiated_hash, source_chain, destination_chain, sender, recipient, denom, amount)
	VALUES ($1::TEXT, $2::BIGINT, $3, $4, $5, $6::TEXT, $7, $8, $9)
	ON CONFLICT (user_initiated_hash) 
	DO UPDATE SET 
		operation_unique_id = EXCLUDED.operation_unique_id,
		height = EXCLUDED.height,
		source_chain = EXCLUDED.source_chain,
		destination_chain = EXCLUDED.destination_chain,
		sender = EXCLUDED.sender,
		recipient = EXCLUDED.recipient,
		denom = EXCLUDED.denom,
		amount = EXCLUDED.amount
	RETURNING id
`
	var id int64
	err := db.SQL.QueryRow(
		stmt,
		tx.OperationUniqueID,
		tx.Height,
		tx.UserInitiatedHash,
		tx.SourceChain,
		tx.DestinationChain,
		tx.Sender,
		tx.Recipient,
		tx.Denom,
		tx.Amount,
	).Scan(&id)
	return id, err
}

// GetBridgeTransaction retrieves a bridge transaction by its operation unique ID.
// It returns the transaction if found, or an error if not found.
func (db *Db) GetBridgeTransaction(operationUniqueID string) (types.BridgeTransaction, error) {
	stmt := `
	SELECT operation_unique_id, height, user_initiated_hash, source_chain, destination_chain, sender, recipient, denom, amount
	FROM bridge_transaction
	WHERE operation_unique_id = $1::TEXT
`

	row := db.SQL.QueryRow(stmt, operationUniqueID)

	var bridgeTx types.BridgeTransaction
	err := row.Scan(
		&bridgeTx.OperationUniqueID,
		&bridgeTx.Height,
		&bridgeTx.UserInitiatedHash,
		&bridgeTx.SourceChain,
		&bridgeTx.DestinationChain,
		&bridgeTx.Sender,
		&bridgeTx.Recipient,
		&bridgeTx.Denom,
		&bridgeTx.Amount,
	)

	if err != nil {
		return types.BridgeTransaction{}, err
	}

	return bridgeTx, nil
}

// SaveBridgeEvidence saves the bridge evidence to the database.
// It returns the ID of the evidence if it was inserted, or the existing ID if it was already present.
// The evidence is identified by its hash and relayer address.
func (db *Db) SaveBridgeEvidence(evidence *types.BridgeEvidence) (int64, error) {
	stmt := `
	INSERT INTO bridge_evidence (transaction_id, height, hash, msg_index, relayer_address, threshold_reached, settlement_hash, result)
	VALUES ($1, $2, $3, $4, $5, $6, $7::TEXT, $8)
	ON CONFLICT (hash, msg_index) DO UPDATE SET
		transaction_id = EXCLUDED.transaction_id,
		height = EXCLUDED.height,
		threshold_reached = EXCLUDED.threshold_reached,
		settlement_hash = EXCLUDED.settlement_hash,
		result = EXCLUDED.result
	RETURNING id
`
	var id int64
	err := db.SQL.QueryRow(
		stmt,
		evidence.TransactionId,
		evidence.Height,
		evidence.Hash,
		evidence.MsgIndex,
		evidence.RelayerAddress,
		evidence.ThresholdReached,
		evidence.SettlementHash,
		evidence.Result,
	).Scan(&id)
	return id, err
}
