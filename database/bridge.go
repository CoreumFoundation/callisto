package database

import (
	"github.com/forbole/callisto/v4/types"
	"github.com/lib/pq"
)

// SaveOutgoingTransfer saves the outgoing transfer to the database
func (db *Db) SaveOutgoingTransfer(tx types.BridgeTransaction) error {
	stmt := `
	INSERT INTO 
	bridge_transaction (init_height, init_hash, counterparty, sender, recipient, amount, direction, operation_ids) 
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

	_, err := db.SQL.Exec(
		stmt,
		tx.InitHeight,
		tx.InitHash,
		tx.Counterparty,
		tx.Sender,
		tx.Recipient,
		tx.Amount,
		tx.Direction,
		pq.Array(tx.OperationIDs),
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveOutgoingPendingTransaction saves the outgoing pending transaction to the database
func (db *Db) GetOutgoingPendingTransaction(operationIDs []uint32) (*types.BridgeTransaction, error) {
	stmt := `
	SELECT init_height, init_hash, counterparty, sender, recipient, amount, direction, operation_ids 
	FROM bridge_transaction
	WHERE  operation_ids && $1 AND final_height IS NULL
`

	row := db.SQL.QueryRow(stmt, pq.Array(operationIDs))

	var bridgeTx types.BridgeTransaction
	err := row.Scan(
		&bridgeTx.InitHeight,
		&bridgeTx.InitHash,
		&bridgeTx.Counterparty,
		&bridgeTx.Sender,
		&bridgeTx.Recipient,
		&bridgeTx.Amount,
		&bridgeTx.Direction,
		&bridgeTx.OperationIDs,
	)

	if err != nil {
		return nil, err
	}

	return &bridgeTx, nil
}

// SaveOutgoingPendingEvidence saves the outgoing pending evidence to the database
func (db *Db) SaveOutgoingPendingEvidence(evidence types.BridgeEvidence, operationId uint32) error {
	stmt := `
	INSERT INTO bridge_evidence (tx_id, height, hash, sender, threshold_reached) 
	VALUES ((SELECT id FROM bridge_transaction WHERE final_height IS NULL AND operation_ids && $4), $1, $2, $3, FALSE)
`
	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.Relayer,
		pq.Array([]uint32{operationId}),
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveOutgoingFinalEvidence saves the outgoing final evidence to the database
func (db *Db) SaveOutgoingFinalEvidence(evidence types.BridgeEvidence, operationId uint32, txResult types.BridgeTxResult, counterpartyHash string) error {
	stmt := `
	WITH updated_transaction AS (
		UPDATE bridge_transaction
			SET final_height = $1,
				final_hash = $2,
				counterparty_hash = $5,
				result = $6
			WHERE operation_ids && $4 AND final_height IS NULL
			RETURNING id
	)
	INSERT INTO bridge_evidence (tx_id, height, hash, sender, threshold_reached)
	SELECT ut.id, $1, $2, $3, TRUE
	FROM updated_transaction ut
	WHERE ut.id IS NOT NULL;
`
	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.Relayer,
		pq.Array([]uint32{operationId}),
		counterpartyHash,
		txResult,
	)
	if err != nil {
		return err
	}
	return nil
}

// SaveIncomingPendingTxAndEvidence saves the incoming pending transaction and evidence to the database
func (db *Db) SaveIncomingPendingTxAndEvidence(tx types.BridgeTransaction, relayer string) error {
	stmt := `
	WITH selected_tx AS (
		SELECT id
		FROM bridge_transaction
		WHERE counterparty = $3 AND counterparty_hash = $4
    ), inserted_tx AS (
		INSERT INTO bridge_transaction (init_height, init_hash, counterparty, counterparty_hash, sender, recipient, amount, direction)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE NOT EXISTS (SELECT 1 FROM selected_tx)
		RETURNING id
	)
	INSERT INTO bridge_evidence (tx_id, height, hash, sender, threshold_reached)
	SELECT it.id, $1, $2, $9, FALSE
	FROM (SELECT id FROM inserted_tx UNION ALL SELECT id FROM selected_tx) it;
`

	_, err := db.SQL.Exec(
		stmt,
		tx.InitHeight,
		tx.InitHash,
		tx.Counterparty,
		tx.CounterpartyHash,
		tx.Sender,
		tx.Recipient,
		tx.Amount,
		tx.Direction,
		relayer,
	)

	if err != nil {
		return err
	}
	return nil
}

// SaveIncomingFinalTxAndEvidence saves the incoming final transaction and evidence to the database
func (db *Db) SaveIncomingFinalTxAndEvidence(evidence types.BridgeEvidence, counterparty types.Counterparty, counterpartyHash string, txResult types.BridgeTxResult) error {
	stmt := `
	WITH updated_transaction AS (
		UPDATE bridge_transaction
		SET final_height = $1, final_hash = $2, result = $6
		WHERE counterparty = $4 AND counterparty_hash = $5
		RETURNING id
	)
	INSERT INTO bridge_evidence (tx_id, height, hash, sender, threshold_reached)
	SELECT ut.id, $1, $2, $3, TRUE
	FROM updated_transaction ut;
`

	_, err := db.SQL.Exec(
		stmt,
		evidence.Height,
		evidence.Hash,
		evidence.Relayer,
		counterparty,
		counterpartyHash,
		txResult,
	)

	if err != nil {
		return err
	}
	return nil
}
