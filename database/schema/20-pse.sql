-- PSE Module Event Schema
-- Parent/child structure: one allocation per distribution_id/allocation_type, many transfers
-- Parent table for each distribution allocation
CREATE TABLE
    IF NOT EXISTS pse_distribution_allocation (
        distribution_id BIGINT NOT NULL, -- use scheduled_at as distribution_id
        allocation_type TEXT NOT NULL, -- pse_community | pse_foundation | pse_alliance | pse_partnership | pse_investors | pse_team
        clearing_account_address TEXT NOT NULL,  -- the address of module account holding funds for this allocation
        total_amount NUMERIC(78, 0) NOT NULL, -- total amount distributed in this batch
        community_pool_amount NUMERIC(78, 0) NOT NULL, -- amount sent to community pool (if any)
        total_score NUMERIC(78, 0) NOT NULL DEFAULT 0, -- present for pse_community, 0 for others
        scheduled_at BIGINT NOT NULL, -- scheduled distribution unix timestamp
        start_at_height BIGINT NOT NULL, -- first observed height for this distribution
        CONSTRAINT pse_distribution_allocation_pkey PRIMARY KEY (distribution_id, allocation_type)
    );

CREATE INDEX IF NOT EXISTS pse_distribution_allocation_distribution_idx ON pse_distribution_allocation (distribution_id, allocation_type);

CREATE INDEX IF NOT EXISTS pse_distribution_allocation_clearing_account_idx ON pse_distribution_allocation (clearing_account_address);

CREATE INDEX IF NOT EXISTS pse_distribution_allocation_height_idx ON pse_distribution_allocation (start_at_height);

-- Child table with individual recipient transfers
CREATE TABLE
    IF NOT EXISTS pse_transfer (
        distribution_id BIGINT NOT NULL, -- links to pse_distribution_allocation
        allocation_type TEXT NOT NULL, -- links to pse_distribution_allocation
        recipient_address TEXT NOT NULL, -- recipient of transfer
        amount NUMERIC(78, 0) NOT NULL, -- amount allocated to recipient
        score NUMERIC(78, 0) NOT NULL DEFAULT 0, -- present for pse_community, 0 for others
        height BIGINT NOT NULL,
        CONSTRAINT pse_transfer_pkey PRIMARY KEY (
            distribution_id,
            allocation_type,
            recipient_address
        ),
        CONSTRAINT pse_transfer_fk FOREIGN KEY (distribution_id, allocation_type) REFERENCES pse_distribution_allocation (distribution_id, allocation_type) ON DELETE CASCADE
    );

CREATE INDEX IF NOT EXISTS pse_transfer_distribution_idx ON pse_transfer (distribution_id, allocation_type);

CREATE INDEX IF NOT EXISTS pse_transfer_recipient_idx ON pse_transfer (recipient_address);

CREATE INDEX IF NOT EXISTS pse_transfer_height_idx ON pse_transfer (height);
