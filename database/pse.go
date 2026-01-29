package database

import (
	"fmt"
)

// PSETransfer represents a single recipient transfer entry.
type PSETransfer struct {
	RecipientAddress string
	Amount           string
	Score            string
	Height           int64
}

// SavePSEDistributionAllocation stores or updates a parent allocation and its transfers.
// Community distributions may span multiple blocks; when recalcTotalFromTransfers is true,
// total_amount is recomputed from child rows to remain idempotent across repeated events
// while start_at_height tracks the earliest observed height.
func (db *Db) SavePSEDistributionAllocation(
	distributionID int64,
	allocationType string,
	clearingAccount string,
	totalAmount string,
	communityPoolAmount string,
	totalScore string,
	scheduledAt int64,
	startAtHeight int64,
	transfers []PSETransfer,
	recalculateTotalFromTransfers bool,
) error {
	tx, err := db.SQL.Begin()
	if err != nil {
		return fmt.Errorf("error while starting db transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	parentStmt := `
INSERT INTO pse_distribution_allocation(
	distribution_id, allocation_type, clearing_account_address, total_amount, community_pool_amount, total_score, scheduled_at, start_at_height
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (distribution_id, allocation_type) DO UPDATE
SET clearing_account_address = excluded.clearing_account_address,
	community_pool_amount = excluded.community_pool_amount,
	total_score = excluded.total_score,
	total_amount = excluded.total_amount,
	scheduled_at = excluded.scheduled_at,
	start_at_height = LEAST(pse_distribution_allocation.start_at_height, excluded.start_at_height)
`

	if _, err := tx.Exec(parentStmt, distributionID, allocationType, clearingAccount, totalAmount, communityPoolAmount, totalScore, scheduledAt, startAtHeight); err != nil {
		return fmt.Errorf("error while upserting pse_distribution_allocation: %w", err)
	}

	childStmt := `
INSERT INTO pse_transfer(distribution_id, allocation_type, recipient_address, amount, score, height)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (distribution_id, allocation_type, recipient_address) DO UPDATE
SET amount = excluded.amount,
	score = excluded.score,
	height = excluded.height
`

	for _, t := range transfers {
		if _, err := tx.Exec(childStmt, distributionID, allocationType, t.RecipientAddress, t.Amount, t.Score, t.Height); err != nil {
			return fmt.Errorf("error while upserting pse_transfer: %w", err)
		}
	}

	if recalculateTotalFromTransfers {
		if _, err := tx.Exec(`
UPDATE pse_distribution_allocation AS p
SET total_amount = sub.total_amount,
    total_score = sub.total_score
FROM (
	SELECT
		distribution_id,
		allocation_type,
		COALESCE(SUM(amount), 0) AS total_amount,
		COALESCE(SUM(score), 0) AS total_score
	FROM pse_transfer
	WHERE distribution_id = $1 AND allocation_type = $2
	GROUP BY distribution_id, allocation_type
) AS sub
WHERE p.distribution_id = sub.distribution_id
  AND p.allocation_type = sub.allocation_type
`, distributionID, allocationType); err != nil {
			return fmt.Errorf("error while recalculating totals: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error while committing transaction: %w", err)
	}

	return nil
}
