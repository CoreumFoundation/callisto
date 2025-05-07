/* ---- Transaction ---- */
CREATE TABLE
    bridge_transaction (
        id SERIAL NOT NULL PRIMARY KEY,
        init_height BIGINT NOT NULL REFERENCES block (height),
        final_height BIGINT NULL DEFAULT NULL REFERENCES block (height),
        init_hash TEXT NOT NULL,
        final_hash TEXT NULL DEFAULT NULL,
        counterparty TEXT NOT NULL,
        counterparty_hash TEXT NULL DEFAULT NULL,
        sender TEXT NOT NULL,
        recipient TEXT NOT NULL,
        amount TEXT NOT NULL,
        direction TEXT NOT NULL,
        result TEXT NULL DEFAULT NULL,
        operation_ids INTEGER[] NULL DEFAULT NULL
    );
CREATE INDEX bridge_transaction_counterparty_counterparty_hash_idx ON bridge_transaction (counterparty, counterparty_hash);
CREATE INDEX bridge_transaction_init_height_idx ON bridge_transaction (init_height);
CREATE INDEX bridge_transaction_final_height_idx ON bridge_transaction (final_height);
CREATE INDEX bridge_transaction_init_hash_idx ON bridge_transaction (init_hash);
CREATE INDEX bridge_transaction_counterparty_idx ON bridge_transaction (counterparty);
CREATE INDEX bridge_transaction_counterparty_hash_idx ON bridge_transaction (counterparty_hash);
CREATE INDEX bridge_transaction_direction_idx ON bridge_transaction (direction);

/* ---- Evidence ---- */
CREATE TABLE
    bridge_evidence (
        id SERIAL NOT NULL PRIMARY KEY,
        tx_id BIGINT NOT NULL REFERENCES bridge_transaction (id),
        height BIGINT NOT NULL REFERENCES block (height),
        hash TEXT NOT NULL,
        sender TEXT NOT NULL,
        threshold_reached BOOLEAN NOT NULL
    );
CREATE INDEX bridge_evidence_height_idx ON bridge_evidence (height);
CREATE INDEX bridge_evidence_hash_idx ON bridge_evidence (hash);
CREATE INDEX bridge_evidence_threshold_reached_idx ON bridge_evidence (threshold_reached);