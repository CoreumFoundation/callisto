package database

import (
	"github.com/forbole/callisto/v4/types"
)

// SaveOutgoingTransfer saves the outgoing transfer to the database
func (db *Db) SaveOutgoingTransfer(tx types.BridgeTransaction) error {
	stmt := `
	INSERT INTO 
	bridge_transaction (user_initiated_height, user_initiated_hash, source_chain, destination_chain, sender, recipient, amount, operation_id) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`
	_, err := db.SQL.Exec(
		stmt,
		tx.UserInitiatedHeight,
		tx.UserInitiatedHash,
		tx.SourceChain,
		tx.DestinationChain,
		tx.Sender,
		tx.Recipient,
		tx.Amount,
		tx.OperationID,
	)
	if err != nil {
		return err
	}
	return nil
}

// GetOutgoingPendingTransaction saves the outgoing pending transaction to the database
func (db *Db) GetOutgoingPendingTransaction(operationID uint32) (*types.BridgeTransaction, error) {
	stmt := `
	SELECT user_initiated_hash, source_chain, destination_chain, sender, recipient, amount, operation_id
	FROM bridge_transaction
	WHERE operation_id = $1 AND final_evidence_hash IS NULL
`

	row := db.SQL.QueryRow(stmt, operationID)

	var bridgeTx types.BridgeTransaction
	err := row.Scan(
		&bridgeTx.UserInitiatedHash,
		&bridgeTx.SourceChain,
		&bridgeTx.DestinationChain,
		&bridgeTx.Sender,
		&bridgeTx.Recipient,
		&bridgeTx.Amount,
		&bridgeTx.OperationID,
	)

	if err != nil {
		return nil, err
	}

	return &bridgeTx, nil
}

// SaveOutgoingPendingEvidence saves the outgoing pending evidence to the database
func (db *Db) SaveOutgoingPendingEvidence(evidence types.BridgeEvidence, operationId uint32) error {
	stmt := `
	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached) 
	VALUES ((SELECT id FROM bridge_transaction WHERE final_evidence_hash IS NULL AND operation_id = $4), $1, $2, $3, FALSE)
`
	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.RelayerAddress,
		operationId,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveOutgoingFinalEvidence saves the outgoing final evidence to the database
func (db *Db) SaveOutgoingFinalEvidence(evidence types.BridgeEvidence, operationId uint32, txResult types.BridgeTxResult, settlementHash string) error {
	stmt := `
	WITH updated_transaction AS (
		UPDATE bridge_transaction
			SET 
				settlement_hash = $5,
				final_evidence_hash = $2,
				result = $6
			WHERE final_evidence_hash IS NULL AND operation_id = $4
			RETURNING id
	)
	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
	SELECT ut.id, $1, $2, $3, TRUE
	FROM updated_transaction ut
	WHERE ut.id IS NOT NULL;
`
	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.RelayerAddress,
		operationId,
		settlementHash,
		txResult,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveIncomingPendingTxAndEvidence saves the incoming pending transaction and evidence to the database
func (db *Db) SaveIncomingPendingTxAndEvidence(tx types.BridgeTransaction, evidence types.BridgeEvidence) error {
	stmt := `
	WITH selected_tx AS (
		SELECT id
		FROM bridge_transaction
		WHERE source_chain = $2 AND user_initiated_hash = $1
    ), inserted_tx AS (
		INSERT INTO bridge_transaction (user_initiated_hash, source_chain, destination_chain, sender, recipient, amount)
		SELECT $1, $2, $3, $4, $5, $6
		WHERE NOT EXISTS (SELECT 1 FROM selected_tx)
		RETURNING id
	)
	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
	SELECT it.id, $7, $8, $9, FALSE
	FROM (SELECT id FROM inserted_tx UNION ALL SELECT id FROM selected_tx) it;
`

	_, err := db.SQL.Exec(
		stmt,
		tx.UserInitiatedHash,
		tx.SourceChain,
		tx.DestinationChain,
		tx.Sender,
		tx.Recipient,
		tx.Amount,
		evidence.Height,
		evidence.Hash,
		evidence.RelayerAddress,
	)

	if err != nil {
		return err
	}
	return nil
}

// SaveIncomingFinalTxAndEvidence saves the incoming final transaction and evidence to the database
func (db *Db) SaveIncomingFinalTxAndEvidence(evidence types.BridgeEvidence, sourceChain types.Chain, sourceHash string, txResult types.BridgeTxResult) error {
	stmt := `
	WITH updated_transaction AS (
		UPDATE bridge_transaction
		SET settlement_hash = $2, final_evidence_hash = $2, result = $6
		WHERE source_chain = $4 AND user_initiated_hash = $5
		RETURNING id
	)
	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
	SELECT ut.id, $1, $2, $3, TRUE
	FROM updated_transaction ut;
`

	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.RelayerAddress,
		sourceChain,
		sourceHash,
		txResult,
	)

	if err != nil {
		return err
	}
	return nil
}
