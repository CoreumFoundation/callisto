package local

import (
	"context"
	"encoding/json"
	"fmt"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/forbole/juno/v6/node/remote"
	"github.com/samber/lo"

	"github.com/forbole/callisto/v4/modules/bridge/source"
	"github.com/forbole/callisto/v4/types"
)

var _ source.Source = &Source{}

// Source implements types.Source by using a local node
type Source struct {
	*remote.Source
	queryClient wasmtypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, queryClient wasmtypes.QueryClient) *Source {
	return &Source{
		Source:      source,
		queryClient: queryClient,
	}
}

func (s Source) Name() string {
	return "remote"
}

func (s *Source) GetSendToXRPLOperationIDs(
	recipient string,
	height uint64,
) ([]uint32, error) {
	beforeCtx := remote.GetHeightRequestContext(s.Ctx, int64(height-1))
	operationsBefore, err := s.getPendingOperations(beforeCtx)
	if err != nil {
		return nil, err
	}

	afterCtx := remote.GetHeightRequestContext(s.Ctx, int64(height))
	operationsAfter, err := s.getPendingOperations(afterCtx)
	if err != nil {
		return nil, err
	}

	operationsBeforeMap := lo.SliceToMap(operationsBefore, func(operation types.Operation) (uint32, types.Operation) {
		return operation.GetOperationID(), operation
	})

	operationIDs := make([]uint32, 0)
	for _, operation := range operationsAfter {
		if _, ok := operationsBeforeMap[operation.GetOperationID()]; !ok {
			if operation.OperationType.CoreumToXRPLTransfer == nil {
				continue
			}
			if operation.OperationType.CoreumToXRPLTransfer.Recipient != recipient {
				continue
			}
			operationIDs = append(operationIDs, operation.GetOperationID())
		}
	}

	return operationIDs, nil
}

// getPendingOperations returns a list of all pending operations.
func (s Source) getPendingOperations(ctx context.Context) ([]types.Operation, error) {
	operations := make([]types.Operation, 0)
	var startAfterKey *uint32
	for {
		var res types.PendingOperationsResponse
		err := s.query(ctx, map[types.QueryMethod]types.PagingUint32KeyRequest{
			types.QueryMethodPendingOperations: {
				StartAfterKey: startAfterKey,
				Limit:         &types.Limit,
			},
		}, &res)
		if err != nil {
			return nil, err
		}
		if len(res.Operations) == 0 {
			break
		}
		operations = append(operations, res.Operations...)
		startAfterKey = &res.LastKey
	}

	return operations, nil
}

func (s Source) query(ctx context.Context, request, response any) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal query request: %w", err)
	}

	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   types.BridgeContractAddress,
		QueryData: payload,
	}
	resp, err := s.queryClient.SmartContractState(ctx, query)
	if err != nil {
		return fmt.Errorf("query failed, request:%+v: %w", request, err)
	}

	if err := json.Unmarshal(resp.Data, response); err != nil {
		return fmt.Errorf("failed to unmarshal wasm contract response, request:%s, response:%s",
			string(payload),
			string(resp.Data))

	}

	return nil
}
