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
		WHERE hash = $3 and relayer_address = $4
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

// func (db *Db) GetEvidencesByTxHash(txHash string) ([]*types.BridgeEvidence, error) {
// 	stmt := `
// 	SELECT id, transaction_id, height, hash, relayer_address, threshold_reached, settlement_hash, result
// 	FROM bridge_evidence
// 	WHERE hash = $1
// `
// 	rows, err := db.SQL.Query(stmt, txHash)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var evidences []*types.BridgeEvidence
// 	for rows.Next() {
// 		var evidence types.BridgeEvidence
// 		err := rows.Scan(
// 			&evidence.ID,
// 			&evidence.TransactionId,
// 			&evidence.Height,
// 			&evidence.Hash,
// 			&evidence.RelayerAddress,
// 			&evidence.ThresholdReached,
// 			&evidence.SettlementHash,
// 			&evidence.Result,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		evidences = append(evidences, &evidence)
// 	}
// 	return evidences, nil
// }

// // SaveOutgoingPendingEvidence saves the outgoing pending evidence to the database
// func (db *Db) SaveOutgoingPendingEvidence(evidence types.BridgeEvidence, operationId uint32) error {
// 	stmt := `
// 	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
// 	VALUES ((SELECT id FROM bridge_transaction WHERE final_evidence_hash IS NULL AND operation_unique_id = $4), $1, $2, $3, FALSE)
// `
// 	_, err := db.SQL.Exec(
// 		stmt,
// 		evidence.Height,
// 		evidence.Hash,
// 		evidence.RelayerAddress,
// 		operationId,
// 	)
// 	return err
// }

// // SaveOutgoingFinalEvidence saves the outgoing final evidence to the database
// func (db *Db) SaveOutgoingFinalEvidence(evidence types.BridgeEvidence, operationId uint32, txResult types.BridgeTxResult, settlementHash string) error {
// 	stmt := `
// 	WITH updated_transaction AS (
// 		UPDATE bridge_transaction
// 			SET
// 				settlement_hash = $5,
// 				final_evidence_hash = $2,
// 				result = $6
// 			WHERE final_evidence_hash IS NULL AND operation_unique_id = $4
// 			RETURNING id
// 	)
// 	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
// 	SELECT ut.id, $1, $2, $3, TRUE
// 	FROM updated_transaction ut
// 	WHERE ut.id IS NOT NULL;
// `
// 	_, err := db.SQL.Exec(
// 		stmt,
// 		evidence.Height,
// 		evidence.Hash,
// 		evidence.RelayerAddress,
// 		operationId,
// 		settlementHash,
// 		txResult,
// 	)
// 	return err
// }

// // SaveIncomingPendingTxAndEvidence saves the incoming pending transaction and evidence to the database
// func (db *Db) SaveIncomingPendingTxAndEvidence(tx types.BridgeTransaction, evidence types.BridgeEvidence) error {
// 	stmt := `
// 	WITH selected_tx AS (
// 		SELECT id
// 		FROM bridge_transaction
// 		WHERE source_chain = $2 AND user_initiated_hash = $1
//     ), inserted_tx AS (
// 		INSERT INTO bridge_transaction (user_initiated_hash, source_chain, destination_chain, sender, recipient, amount)
// 		SELECT $1, $2, $3, $4, $5, $6
// 		WHERE NOT EXISTS (SELECT 1 FROM selected_tx)
// 		RETURNING id
// 	)
// 	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
// 	SELECT it.id, $7, $8, $9, FALSE
// 	FROM (SELECT id FROM inserted_tx UNION ALL SELECT id FROM selected_tx) it;
// `

// 	_, err := db.SQL.Exec(
// 		stmt,
// 		tx.UserInitiatedHash,
// 		tx.SourceChain,
// 		tx.DestinationChain,
// 		tx.Sender,
// 		tx.Recipient,
// 		tx.Amount,
// 		evidence.Height,
// 		evidence.Hash,
// 		evidence.RelayerAddress,
// 	)
// 	return err
// }

// // SaveIncomingFinalizedTxAndEvidence saves the incoming final transaction and evidence to the database
// func (db *Db) SaveIncomingFinalizedTxAndEvidence(evidence types.BridgeEvidence, sourceChain types.Chain, sourceHash string, txResult types.BridgeTxResult) error {
// 	stmt := `
// 	WITH updated_transaction AS (
// 		UPDATE bridge_transaction
// 		SET settlement_hash = $2, final_evidence_hash = $2, result = $6
// 		WHERE source_chain = $4 AND user_initiated_hash = $5
// 		RETURNING id
// 	)
// 	INSERT INTO bridge_evidence (transaction_id, height, hash, relayer_address, threshold_reached)
// 	SELECT ut.id, $1, $2, $3, TRUE
// 	FROM updated_transaction ut;
// `

// 	_, err := db.SQL.Exec(
// 		stmt,
// 		evidence.Height,
// 		evidence.Hash,
// 		evidence.RelayerAddress,
// 		sourceChain,
// 		sourceHash,
// 		txResult,
// 	)
// 	return err
// }
