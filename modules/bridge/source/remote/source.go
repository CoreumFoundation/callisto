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

// GetOutgoingPendingOperationSequence returns the operation sequence of the outgoing operations
// Implementation is based on the actual bridge relayer logic, for simple replication of the code in the future,
// the same style of the reference code is used.
// https://github.com/tokenize-x/tx-xrpl-bridge/blob/be8b90d4d8cde0eb74c60ea14edfe06397e8c31f/relayer/coreum/contract.go#L1361
func (s Source) GetOutgoingPendingOperationSequence(
	contractAddress string,
	recipient string,
	height uint64,
) (uint32, error) {
	beforeCtx := remote.GetHeightRequestContext(s.Ctx, int64(height-1))
	operationsBefore, err := s.getPendingOperations(beforeCtx, contractAddress)
	if err != nil {
		return 0, err
	}

	afterCtx := remote.GetHeightRequestContext(s.Ctx, int64(height))
	operationsAfter, err := s.getPendingOperations(afterCtx, contractAddress)
	if err != nil {
		return 0, err
	}

	operationsBeforeMap := lo.SliceToMap(operationsBefore, func(operation types.Operation) (uint32, types.Operation) {
		return operation.GetOperationSequence(), operation
	})

	operationSequences := make([]uint32, 0)
	for _, operation := range operationsAfter {
		if _, ok := operationsBeforeMap[operation.GetOperationSequence()]; !ok {
			if operation.OperationType.CoreumToXRPLTransfer == nil {
				continue
			}
			if operation.OperationType.CoreumToXRPLTransfer.Recipient != recipient {
				continue
			}
			operationSequences = append(operationSequences, operation.GetOperationSequence())
		}
	}

	switch len(operationSequences) {
	case 0:
		return 0, fmt.Errorf("no operation ID found for recipient %s", recipient)
	case 1:
		return operationSequences[0], nil
	default:
		return 0, fmt.Errorf("multiple operation IDs found for recipient %s: %v", recipient, operationSequences)
	}
}

// getPendingOperations returns a list of all pending operations.
func (s Source) getPendingOperations(ctx context.Context, contractAddress string) ([]types.Operation, error) {
	operations := make([]types.Operation, 0)
	var startAfterKey *uint32
	for {
		var res types.PendingOperationsResponse
		err := s.query(ctx, contractAddress, map[types.QueryMethod]types.PagingUint32KeyRequest{
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

func (s Source) query(ctx context.Context, contractAddress string, request, response any) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal query request: %w", err)
	}

	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddress,
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
