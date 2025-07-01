package bank

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	juno "github.com/forbole/juno/v6/types"

	eventutils "github.com/forbole/callisto/v4/utils/events"
)

func (m *Module) HandleMsg(index int, _ juno.Message, tx *juno.Transaction) error {
	return m.UpdateAccountsBalances(index, tx)
}

func (m *Module) UpdateAccountsBalances(msgIndex int, tx *juno.Transaction) error {
	if len(tx.Events) == 0 {
		return nil
	}

	for _, eventType := range []string{banktypes.EventTypeCoinReceived, banktypes.EventTypeCoinSpent} {
		if err := m.updateBalanceForEventType(msgIndex, tx, eventType); err != nil {
			return err
		}
	}
	return nil
}

func (m *Module) updateBalanceForEventType(msgIndex int, tx *juno.Transaction, eventType string) error {
	accountAttribute := banktypes.AttributeKeySpender
	if eventType == banktypes.EventTypeCoinReceived {
		accountAttribute = banktypes.AttributeKeyReceiver
	}

	block, err := m.db.GetLastBlockHeightAndTimestamp()
	if err != nil {
		return fmt.Errorf("error while getting latest block height: %s", err)
	}

	msgEvents := eventutils.FindEventsByMsgIndex(sdk.StringifyEvents(tx.Events), msgIndex)

	type addressDenom struct {
		address string
		denom   string
	}
	addressDenomSet := make(map[addressDenom]struct{})

	for _, event := range msgEvents {
		if event.Type != eventType {
			continue
		}
		account, err := tx.FindAttributeByKey(event, accountAttribute)
		if err != nil {
			return err
		}

		coinString, err := tx.FindAttributeByKey(event, sdk.AttributeKeyAmount)
		if err != nil {
			return err
		}

		coin, err := sdk.ParseCoinNormalized(coinString)
		if err != nil {
			return err
		}

		// since the main governance token exists in every transaction, we have decided to skip processing
		// that token.
		if coin.Denom == m.baseDenom {
			continue
		}

		addressDenomSet[addressDenom{address: account, denom: coin.Denom}] = struct{}{}
	}

	for ad := range addressDenomSet {
		storedBalance, found, err := m.db.GetAccountDenomBalance(ad.address, ad.denom)
		if err != nil {
			return err
		}
		if found && storedBalance.Height >= block.Height {
			continue
		}

		queriedBalance, err := m.keeper.GetAccountDenomBalance(ad.address, ad.denom, block.Height)
		if err != nil {
			return err
		}
		if queriedBalance == nil {
			return fmt.Errorf("query balance return nil, account: %s, denom:%s", ad.address, ad.denom)
		}

		if err := m.db.SaveAccountDenomBalance(ad.address, *queriedBalance, block.Height); err != nil {
			return err
		}
	}

	return nil
}
