/* ---- Transaction ---- */
CREATE TABLE
    bridge_transaction (
        id SERIAL NOT NULL PRIMARY KEY,
        operation_id INTEGER NULL DEFAULT NULL,
        user_initiated_height BIGINT NULL,
        user_initiated_hash TEXT NOT NULL,
        settlement_hash TEXT NULL DEFAULT NULL,
        final_evidence_hash TEXT NULL DEFAULT NULL,
        source_chain TEXT NOT NULL,
        destination_chain TEXT NOT NULL,
        sender TEXT NOT NULL,
        recipient TEXT NOT NULL,
        amount TEXT NOT NULL,
        result TEXT NULL DEFAULT NULL
    );
CREATE INDEX bridge_transaction_user_initiated_hash_idx ON bridge_transaction (user_initiated_hash);
CREATE INDEX bridge_transaction_sender_idx ON bridge_transaction (sender);
CREATE INDEX bridge_transaction_recipient_idx ON bridge_transaction (recipient);
CREATE INDEX bridge_transaction_final_evidence_hash_operation_id_idx ON bridge_transaction (final_evidence_hash, operation_id);

/* ---- Evidence ---- */
CREATE TABLE
    bridge_evidence (
        id SERIAL NOT NULL PRIMARY KEY,
        transaction_id BIGINT NOT NULL REFERENCES bridge_transaction (id),
        height BIGINT NOT NULL REFERENCES block (height),
        hash TEXT NOT NULL, --indexed
        relayer_address TEXT NOT NULL, --indexed
        threshold_reached BOOLEAN NOT NULL
    );
CREATE INDEX bridge_evidence_hash_idx ON bridge_evidence (hash);
CREATE INDEX bridge_evidence_relayer_address_idx ON bridge_evidence (relayer_address);