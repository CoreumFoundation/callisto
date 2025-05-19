package database

import (
	"github.com/forbole/callisto/v4/types"
)

// SaveBridgeTransaction saves the bridge transaction to the database.
// It returns the ID of the transaction if it was inserted, or the existing ID if it was already present.
func (db *Db) SaveBridgeTransaction(tx types.BridgeTransaction) (int64, error) {
	stmt := `
	WITH selected_tx AS (
		SELECT id
		FROM bridge_transaction
		WHERE user_initiated_hash = $3 AND ($1::TEXT IS NULL OR operation_unique_id = $1::TEXT)
	), inserted_tx AS (
		INSERT INTO bridge_transaction (operation_unique_id, height, user_initiated_hash, source_chain, destination_chain, issuer, sender, recipient, denom, amount)
		SELECT $1::TEXT, $2::BIGINT, $3, $4, $5, $6::TEXT, $7::TEXT, $8, $9, $10
		WHERE NOT EXISTS (SELECT 1 FROM selected_tx)
		RETURNING id
	)
	SELECT id FROM inserted_tx UNION ALL SELECT id FROM selected_tx	
`
	var id int64
	err := db.SQL.QueryRow(
		stmt,
		tx.OperationID,
		tx.Height,
		tx.UserInitiatedHash,
		tx.SourceChain,
		tx.DestinationChain,
		tx.Issuer,
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
	SELECT operation_unique_id, height, user_initiated_hash, source_chain, destination_chain, issuer, sender, recipient, denom, amount
	FROM bridge_transaction
	WHERE operation_unique_id = $1::TEXT
`

	row := db.SQL.QueryRow(stmt, operationUniqueID)

	var bridgeTx types.BridgeTransaction
	err := row.Scan(
		&bridgeTx.OperationID,
		&bridgeTx.Height,
		&bridgeTx.UserInitiatedHash,
		&bridgeTx.SourceChain,
		&bridgeTx.DestinationChain,
		&bridgeTx.Issuer,
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
func (db *Db) SaveBridgeEvidence(evidence types.BridgeEvidence) (int64, error) {
	stmt := `
	WITH selected_ev AS (
		SELECT id
		FROM bridge_evidence
		WHERE transaction_id = $1 and relayer_address = $4
    ), inserted_ev AS (
		INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached, settlement_hash, result) 
		select $1, $2, $3, $4, $5, $6::TEXT, $7
		WHERE NOT EXISTS (SELECT 1 FROM selected_ev)
		RETURNING id
	)
	SELECT id FROM inserted_ev UNION ALL SELECT id FROM selected_ev
`
	var id int64
	err := db.SQL.QueryRow(
		stmt,
		evidence.TransactionId,
		evidence.Height,
		evidence.Hash,
		evidence.RelayerAddress,
		evidence.ThresholdReached,
		evidence.SettlementHash,
		evidence.Result,
	).Scan(&id)
	return id, err
}
