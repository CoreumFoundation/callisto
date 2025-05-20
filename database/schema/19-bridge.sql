/* ---- Transaction ---- */
CREATE TABLE
    bridge_transaction (
        id SERIAL NOT NULL PRIMARY KEY,
        operation_unique_id TEXT NULL DEFAULT NULL,
        height BIGINT NULL,
        user_initiated_hash TEXT NOT NULL,
        source_chain TEXT NOT NULL,
        destination_chain TEXT NOT NULL,
        sender TEXT NULL DEFAULT NULL,
        recipient TEXT NOT NULL,
        denom TEXT NOT NULL,
        amount TEXT NOT NULL
    );
CREATE UNIQUE INDEX bridge_transaction_user_initiated_hash_idx ON bridge_transaction (user_initiated_hash);
CREATE INDEX bridge_transaction_operation_unique_id_idx ON bridge_transaction (operation_unique_id); -- for GetBridgeTransaction
CREATE INDEX bridge_transaction_sender_idx ON bridge_transaction (sender);
CREATE INDEX bridge_transaction_recipient_idx ON bridge_transaction (recipient);


/* ---- Evidence ---- */
CREATE TABLE
    bridge_evidence (
        id SERIAL NOT NULL PRIMARY KEY,
        transaction_id BIGINT NOT NULL REFERENCES bridge_transaction (id),
        height BIGINT NOT NULL REFERENCES block (height),
        hash TEXT NOT NULL,
        msg_index BIGINT NOT NULL,
        relayer_address TEXT NOT NULL,
        threshold_reached BOOLEAN NOT NULL,
        settlement_hash TEXT NULL DEFAULT NULL,
        result TEXT NULL DEFAULT NULL
    );

CREATE INDEX bridge_evidence_hash_idx ON bridge_evidence (hash);
CREATE INDEX bridge_evidence_relayer_address_idx ON bridge_evidence (relayer_address);
CREATE UNIQUE INDEX bridge_evidence_hash_msg_index_idx ON bridge_evidence (hash, msg_index); -- for SaveBridgeEvidence