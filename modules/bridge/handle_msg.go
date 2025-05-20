package bridge

import (
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/callisto/v4/types"
	"github.com/forbole/callisto/v4/utils"
	juno "github.com/forbole/juno/v6/types"

	"github.com/rs/zerolog/log"
)

// TxHandler is an interface that defines the methods for handling transactions
// in the bridge module. It is used to handle different bridges and their specific
// transaction types. The interface is implemented by different handlers for
// different bridges, such as XrplMsgHandler.
type TxHandler interface {
	HandleMsg() error
}

// DbHandler is an interface that defines the methods for handling
// database operations in the bridge module. It is used to interact with
// the database and perform operations such as saving transactions and evidence.
type DbHandler interface {
	GetBridgeTransaction(id string) (types.BridgeTransaction, error)
	SaveBridgeTransaction(tx *types.BridgeTransaction) (int64, error)
	SaveBridgeEvidence(evidence *types.BridgeEvidence) (int64, error)
}

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

	if cosmosMsg == nil {
		return fmt.Errorf("error while unpacking message: %s", msg.GetType())
	}

	var handler TxHandler

	// at the moment there is only one bridge contract which is the xrpl contract
	switch {
	case cosmosMsg.Contract == m.cfg.ContractAddress:
		handler = NewXrplMsgHandler(m.cfg.ContractAddress, tx.Height, msg, tx, m.db, m.Source)
	default:
		return nil
	}

	log.Debug().Str("module", "bridge").Str("hash", tx.TxHash).Uint64("height", tx.Height).Msg(fmt.Sprintf("handling bridge message %s", msg.GetType()))

	err := handler.HandleMsg()
	if err != nil {
		fmt.Printf("Error when handling bridge transaction message, error: %s", err)
	}

	return nil
}
