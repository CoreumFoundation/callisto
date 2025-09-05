package database

import (
	"github.com/forbole/callisto/v4/types"
)

// SaveBridgeTransaction saves and overwrites the bridge transaction to the database.
func (db *Db) SaveBridgeTransaction(tx *types.BridgeTransaction) (int64, error) {
	stmt := `
	INSERT INTO bridge_transaction (operation_unique_id, on_chain_unique_key, height, user_initiated_hash, msg_index, source_chain, destination_chain, sender, recipient, denom, amount)
	VALUES ($1::TEXT, $2, $3::BIGINT, $4, $5, $6, $7::TEXT, $8, $9, $10, $11)
	ON CONFLICT (user_initiated_hash, msg_index) 
	DO UPDATE SET 
		operation_unique_id = EXCLUDED.operation_unique_id,
		on_chain_unique_key = EXCLUDED.on_chain_unique_key,
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
		tx.OnChainUniqueKey,
		tx.Height,
		tx.UserInitiatedHash,
		tx.MsgIndex,
		tx.SourceChain,
		tx.DestinationChain,
		tx.Sender,
		tx.Recipient,
		tx.Denom,
		tx.Amount,
	).Scan(&id)
	return id, err
}

// SaveBridgeEvidence saves and overwrites the bridge evidence to the database.
func (db *Db) SaveBridgeEvidence(evidence *types.BridgeEvidence) (int64, error) {
	stmt := `
	INSERT INTO bridge_evidence (tx_on_chain_unique_key, height, hash, msg_index, relayer_address, threshold_reached, settlement_hash, result)
	VALUES ($1, $2, $3, $4, $5, $6, $7::TEXT, $8)
	ON CONFLICT (hash, msg_index) DO UPDATE SET
		tx_on_chain_unique_key = EXCLUDED.tx_on_chain_unique_key,
		height = EXCLUDED.height,
		threshold_reached = EXCLUDED.threshold_reached,
		settlement_hash = EXCLUDED.settlement_hash,
		result = EXCLUDED.result
	RETURNING id
`
	var id int64
	err := db.SQL.QueryRow(
		stmt,
		evidence.TxOnChainUniqueKey,
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
