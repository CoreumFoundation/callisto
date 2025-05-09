package bridge

import (
	"fmt"
	"strconv"
	"strings"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/forbole/callisto/v4/utils"
	juno "github.com/forbole/juno/v6/types"

	"github.com/rs/zerolog/log"
)

var msgFilter = map[string]bool{
	"/cosmwasm.wasm.v1.MsgExecuteContract": true,
}

// HandleMsgExec implements modules.AuthzMessageModule
func (m *Module) HandleMsgExec(index, _ int, executedMsg juno.Message, tx *juno.Transaction) error {
	return m.HandleMsg(index, executedMsg, tx)
}

// HandleMsg implements modules.MessageModule
func (m *Module) HandleMsg(
	_ int, msg juno.Message, tx *juno.Transaction,
) error {
	if _, ok := msgFilter[msg.GetType()]; !ok {
		return nil
	}

	cosmosMsg := utils.UnpackMessage(m.cdc, msg.GetBytes(), &wasmtypes.MsgExecuteContract{})
	if cosmosMsg.Contract != m.cfg.ContractAddress {
		return nil
	}

	log.Debug().Str("module", "bridge").Str("hash", tx.TxHash).Uint64("height", tx.Height).Msg(fmt.Sprintf("handling bridge message %s", msg.GetType()))

	err := m.addCoreumXrplTwoWayTransfers(tx.Height, msg, tx)
	if err != nil {
		fmt.Printf("Error when adding Coreum to XRPL transfer, error: %s", err)
	}
	return nil
}

// addCoreumXrplTwoWayTransfers adds the coreum to xrpl and xrpl to coreum transfer to the database
func (m *Module) addCoreumXrplTwoWayTransfers(height uint64, _ juno.Message, tx *juno.Transaction) error {
	events := juno.FindEventsByType(tx.Events, "wasm")
	for _, event := range events {
		action, err := juno.FindAttributeByKey(event, "action")
		if err != nil {
			return fmt.Errorf("error while getting action attribute: %s", err)
		}

		switch action.Value {
		case "send_to_xrpl":
			if err := m.handleSendToXrpl(event, height, tx); err != nil {
				return err
			}
		case "save_evidence":
			if err := m.handleSaveEvidence(event, height, tx); err != nil {
				return err
			}
		}
	}
	return nil
}

// handleSendToXrpl handles the send_to_xrpl event
// and saves the outgoing transfer to the database
func (m *Module) handleSendToXrpl(event abci.Event, height uint64, tx *juno.Transaction) error {
	sender, err := juno.FindAttributeByKey(event, "sender")
	if err != nil {
		return fmt.Errorf("error while getting sender attribute: %s", err)
	}
	recipient, err := juno.FindAttributeByKey(event, "recipient")
	if err != nil {
		return fmt.Errorf("error while getting recipient attribute: %s", err)
	}
	coin, err := juno.FindAttributeByKey(event, "coin")
	if err != nil {
		return fmt.Errorf("error while getting coin attribute: %s", err)
	}

	operationIds, err := m.Source.GetSendToXRPLOperationIDs(m.cfg.ContractAddress, recipient.Value, height)
	if err != nil {
		return fmt.Errorf("error while getting operation id: %s", err)
	}

	pendingTx, err := m.db.GetOutgoingPendingTransaction(operationIds)
	if err != nil && !strings.Contains(err.Error(), "sql: no rows in result set") {
		return fmt.Errorf("error while getting pending transaction: %s", err)
	}
	if pendingTx != nil {
		return fmt.Errorf("pending transaction already exists for operation id %v", operationIds)
	}

	return m.db.SaveOutgoingTransfer(types.NewOutgoingPendingBridgeTransaction(
		tx.TxHash,
		height,
		types.CounterpartyXRPL,
		sender.Value,
		recipient.Value,
		coin.Value,
		types.BridgeTxDirOutgoing,
		operationIds,
	))
}

// handleSaveEvidence handles the save_evidence event
// and saves the evidence to the database
func (m *Module) handleSaveEvidence(event abci.Event, height uint64, tx *juno.Transaction) error {
	operationType, err := juno.FindAttributeByKey(event, "operation_type")
	if err != nil && err.Error() != "no attribute with key operation_type found inside event with type wasm" {
		return fmt.Errorf("error while getting operation type attribute: %s", err)
	}

	if operationType.Value != "" && operationType.Value != "coreum_to_xrpl_transfer" {
		return nil
	}

	relayerAcc, err := juno.FindAttributeByKey(event, "sender")
	if err != nil {
		return fmt.Errorf("error while getting sender attribute: %s", err)
	}

	thresholdReached, err := juno.FindAttributeByKey(event, "threshold_reached")
	if err != nil {
		return fmt.Errorf("error while getting threshold reached attribute: %s", err)
	}
	isThresholdReached, err := strconv.ParseBool(thresholdReached.Value)
	if err != nil {
		return fmt.Errorf("error while parsing threshold reached value: %s", err)
	}

	evidence := types.NewBridgeEvidence(
		height,
		tx.TxHash,
		relayerAcc.Value,
		isThresholdReached,
	)

	if operationType.Value == "coreum_to_xrpl_transfer" {
		return m.handleCoreumToXrplEvidence(event, evidence, isThresholdReached)
	}
	return m.handleXrplToCoreumEvidence(event, evidence, isThresholdReached, height, tx)
}

// handleCoreumToXrplEvidence handles the coreum to xrpl evidence
// and saves the evidence to the database
func (m *Module) handleCoreumToXrplEvidence(event abci.Event, evidence types.BridgeEvidence, isThresholdReached bool) error {
	operationId, err := juno.FindAttributeByKey(event, "operation_id")
	if err != nil {
		return fmt.Errorf("error while getting operation id attribute: %s", err)
	}

	operationIdInt, err := strconv.ParseUint(operationId.Value, 10, 32)
	if err != nil {
		return fmt.Errorf("error while parsing operation id: %s", err)
	}

	if isThresholdReached {
		transactionResult, err := juno.FindAttributeByKey(event, "transaction_result")
		if err != nil {
			return fmt.Errorf("error while getting transaction result attribute: %s", err)
		}
		xrplTxHash, err := juno.FindAttributeByKey(event, "tx_hash")
		if err != nil {
			return fmt.Errorf("error while getting tx hash attribute: %s", err)
		}

		return m.db.SaveOutgoingFinalEvidence(
			evidence,
			uint32(operationIdInt),
			types.BridgeTxResultToStr[transactionResult.Value],
			xrplTxHash.Value,
		)
	}

	return m.db.SaveOutgoingPendingEvidence(
		evidence,
		uint32(operationIdInt),
	)
}

// handleXrplToCoreumEvidence handles the xrpl to coreum evidence
// and saves the evidence to the database
func (m *Module) handleXrplToCoreumEvidence(event abci.Event, evidence types.BridgeEvidence, isThresholdReached bool, height uint64, tx *juno.Transaction) error {
	xrplHash, err := juno.FindAttributeByKey(event, "hash")
	if err != nil {
		return fmt.Errorf("error while getting hash attribute: %s", err)
	}
	recipient, err := juno.FindAttributeByKey(event, "recipient")
	if err != nil {
		return fmt.Errorf("error while getting recipient attribute: %s", err)
	}

	if isThresholdReached {
		return m.db.SaveIncomingFinalTxAndEvidence(
			evidence,
			types.CounterpartyXRPL,
			xrplHash.Value,
			types.BridgeTxResultAccepted,
		)
	}

	issuer, err := juno.FindAttributeByKey(event, "issuer")
	if err != nil {
		return fmt.Errorf("error while getting issuer attribute: %s", err)
	}
	currency, err := juno.FindAttributeByKey(event, "currency")
	if err != nil {
		return fmt.Errorf("error while getting currency attribute: %s", err)
	}
	amount, err := juno.FindAttributeByKey(event, "amount")
	if err != nil {
		return fmt.Errorf("error while getting amount attribute: %s", err)
	}

	return m.db.SaveIncomingPendingTxAndEvidence(
		types.NewIncomingPendingBridgeTransaction(
			tx.TxHash,
			height,
			types.CounterpartyXRPL,
			xrplHash.Value,
			issuer.Value,
			recipient.Value,
			strings.Join([]string{amount.Value, currency.Value}, ""),
			types.BridgeTxDirIncoming,
		),
		evidence.Relayer,
	)
}
