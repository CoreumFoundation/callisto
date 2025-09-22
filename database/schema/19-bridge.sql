/* ---- Evidence ---- */
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'bridge_evidence'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'bridge_evidence' AND column_name = 'tx_on_chain_unique_key'
    ) THEN
        DROP TABLE bridge_evidence CASCADE;
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS
    bridge_evidence (
        id SERIAL NOT NULL PRIMARY KEY,
        tx_on_chain_unique_key TEXT NOT NULL,
        height BIGINT NOT NULL REFERENCES block (height),
        hash TEXT NOT NULL,
        msg_index BIGINT NOT NULL,
        relayer_address TEXT NOT NULL,
        threshold_reached BOOLEAN NOT NULL,
        settlement_hash TEXT NULL DEFAULT NULL,
        result TEXT NULL DEFAULT NULL
    );

CREATE INDEX IF NOT EXISTS bridge_evidence_tx_on_chain_unique_key_idx ON bridge_evidence (tx_on_chain_unique_key);
CREATE INDEX IF NOT EXISTS bridge_evidence_hash_idx ON bridge_evidence (hash);
CREATE INDEX IF NOT EXISTS bridge_evidence_relayer_address_idx ON bridge_evidence (relayer_address);
CREATE UNIQUE INDEX IF NOT EXISTS bridge_evidence_hash_msg_index_idx ON bridge_evidence (hash, msg_index); -- for SaveBridgeEvidence

/* ---- Transaction ---- */
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'bridge_transaction'
    ) AND NOT EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'bridge_transaction' AND column_name = 'on_chain_unique_key'
    ) THEN
        DROP TABLE bridge_transaction CASCADE;
    END IF;
END $$;
CREATE TABLE IF NOT EXISTS
    bridge_transaction (
        id SERIAL NOT NULL PRIMARY KEY,
        on_chain_unique_key TEXT NOT NULL UNIQUE,
        operation_unique_id TEXT NULL DEFAULT NULL,
        height BIGINT NULL,
        user_initiated_hash TEXT NOT NULL,
        msg_index BIGINT NOT NULL,
        source_chain TEXT NOT NULL,
        destination_chain TEXT NOT NULL,
        sender TEXT NULL DEFAULT NULL,
        recipient TEXT NOT NULL,
        denom TEXT NOT NULL,
        amount TEXT NOT NULL
    );
CREATE INDEX IF NOT EXISTS bridge_transaction_on_chain_unique_key_idx ON bridge_transaction (on_chain_unique_key);
CREATE INDEX IF NOT EXISTS bridge_transaction_sender_idx ON bridge_transaction (sender);
CREATE INDEX IF NOT EXISTS bridge_transaction_recipient_idx ON bridge_transaction (recipient);
CREATE UNIQUE INDEX IF NOT EXISTS bridge_transaction_user_initiated_hash_msg_index_idx ON bridge_transaction (user_initiated_hash, msg_index); -- for SaveBridgeTransaction
