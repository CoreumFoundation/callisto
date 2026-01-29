package pse

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	tmctypes "github.com/cometbft/cometbft/rpc/core/types"

	"github.com/tokenize-x/tx-chain/v6/testutil/event"
	psetypes "github.com/tokenize-x/tx-chain/v6/x/pse/types"

	juno "github.com/forbole/juno/v6/types"

	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/utils"
	eventsutil "github.com/forbole/callisto/v4/utils/events"
)

// HandleBlock implements BlockModule
func (m *Module) HandleBlock(
	block *tmctypes.ResultBlock, res *tmctypes.ResultBlockResults, _ []*juno.Transaction, _ *tmctypes.ResultValidators,
) error {
	// Handle PSE allocation distributed events emitted in EndBlock (FinalizeBlock)
	if err := m.handleNonCommunityDistributedEvents(block.Block.Height, res.FinalizeBlockEvents); err != nil {
		return fmt.Errorf("error when handling PSE allocation events: %w", err)
	}

	// Handle PSE community distributions emitted by the PSE module (EventCommunityDistributed)
	if err := m.handleCommunityDistributedEvents(block.Block.Height, block.Block.Time.Unix(), res.FinalizeBlockEvents); err != nil {
		return fmt.Errorf("error when handling PSE community distributed events: %w", err)
	}

	return nil
}

// handleNonCommunityDistributedEvents extracts PSE EventAllocationDistributed events and stores them in the database.
func (m *Module) handleNonCommunityDistributedEvents(height int64, events []abci.Event) error {
	log.Debug().Str("module", "pse").Int64("height", height).
		Msg("handling PSE allocation distributed events")

	// Remove the "mode" attribute added by cosmos-sdk before parsing typed events
	events = eventsutil.RemoveAttributeFromEvent(events, "mode")

	// Parse typed events using FindTypedEvents utility
	typedEvents, err := event.FindTypedEvents[*psetypes.EventAllocationDistributed](events)
	if err != nil {
		// Only ignore "no events found" error; return parsing/casting errors
		if strings.Contains(err.Error(), "can't find event") {
			return nil
		}
		return fmt.Errorf("failed to parse PSE allocation events at height %d: %w", height, err)
	}

	for _, evt := range typedEvents {
		allocationType := evt.ClearingAccount
		if allocationType == "" {
			return fmt.Errorf("PSE allocation event missing required clearing_account at height %d", height)
		}

		// Resolve clearing account module name to address using the configured bech32 prefix
		clearingAcctAddr, err := utils.GetModuleAccountAddressFromPrefix(allocationType, m.bech32Prefix)
		if err != nil {
			return fmt.Errorf("failed to resolve clearing account address for %s at height %d: %w", allocationType, height, err)
		}

		// scheduled_at is already uint64, convert to int64
		scheduledAt := int64(evt.ScheduledAt)

		// Create transfers from recipients
		transfers := make([]database.PSETransfer, 0, len(evt.RecipientAddresses))
		for _, addr := range evt.RecipientAddresses {
			transfers = append(transfers, database.PSETransfer{
				RecipientAddress: addr,
				Amount:           evt.AmountPerRecipient.String(),
				Score:            "0", // non-community allocations have no score
				Height:           height,
			})
		}

		if err := m.db.SavePSEDistributionAllocation(
			scheduledAt,
			allocationType,
			clearingAcctAddr,
			evt.TotalAmount.String(),
			evt.CommunityPoolAmount.String(),
			"0", // non-community allocations have no total_score
			scheduledAt,
			height,
			transfers,
			false, // do not recalc total from transfers for non-community
		); err != nil {
			return fmt.Errorf("error while saving PSE event: %w", err)
		}
	}

	return nil
}

// handleCommunityDistributedEvents parses typed EventCommunityDistributed events emitted by the PSE module
// and stores the per-delegator distribution with score metadata.
func (m *Module) handleCommunityDistributedEvents(height int64, blockTime int64, events []abci.Event) error {
	log.Debug().Str("module", "pse").Int64("height", height).
		Msg("handling PSE community distributed events")

	// Remove the "mode" attribute added by cosmos-sdk before parsing typed events
	events = eventsutil.RemoveAttributeFromEvent(events, "mode")

	// Parse typed events using FindTypedEvents utility
	typedEvents, err := event.FindTypedEvents[*psetypes.EventCommunityDistributed](events)
	if err != nil {
		// Only ignore "no events found" error; return parsing/casting errors
		if strings.Contains(err.Error(), "can't find event") {
			return nil
		}
		return fmt.Errorf("failed to parse PSE community events at height %d: %w", height, err)
	}

	for _, evt := range typedEvents {
		recipientAddress := evt.DelegatorAddress
		if recipientAddress == "" {
			continue
		}

		// scheduled_at is already uint64, convert to int64
		scheduledAt := int64(evt.ScheduledAt)

		transfers := []database.PSETransfer{
			{
				RecipientAddress: recipientAddress,
				Amount:           evt.Amount.String(),
				Score:            evt.Score.String(),
				Height:           height,
			},
		}

		// Resolve community clearing account module name to address using the configured bech32 prefix
		communityClearingAcctAddr, err := utils.GetModuleAccountAddressFromPrefix(psetypes.ClearingAccountCommunity, m.bech32Prefix)
		if err != nil {
			return fmt.Errorf("failed to resolve community clearing account address at height %d: %w", height, err)
		}

		if err := m.db.SavePSEDistributionAllocation(
			scheduledAt,
			psetypes.ClearingAccountCommunity,
			communityClearingAcctAddr,
			evt.Amount.String(),
			sdkmath.NewInt(0).String(),
			evt.TotalPseScore.String(),
			scheduledAt,
			height,
			transfers,
			true, // recalc total from transfers to keep latest stats across blocks
		); err != nil {
			return fmt.Errorf("error while saving PSE community distributed event: %w", err)
		}
	}

	return nil
}
