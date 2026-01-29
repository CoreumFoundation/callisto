package remote

import (
	"fmt"

	"github.com/forbole/callisto/v4/modules/pse/source"
	"github.com/forbole/juno/v6/node/remote"
	psetypes "github.com/tokenize-x/tx-chain/v6/x/pse/types"
)

type remoteSource struct {
	src         *remote.Source
	queryClient psetypes.QueryClient
}

// NewSource returns a new remote PSE source with a gRPC query client.
func NewSource(src *remote.Source, queryClient psetypes.QueryClient) source.Source {
	return &remoteSource{
		src:         src,
		queryClient: queryClient,
	}
}

// DelegatorScore queries the current score for a delegator via gRPC.
func (r *remoteSource) DelegatorScore(height int64, address string) (string, error) {
	ctx := remote.GetHeightRequestContext(r.src.Ctx, height)
	req := &psetypes.QueryScoreRequest{Address: address}
	res, err := r.queryClient.Score(ctx, req)
	if err != nil {
		return "", fmt.Errorf("error querying delegator score: %w", err)
	}
	return res.Score.String(), nil
}

func (r *remoteSource) ScheduledDistributions(height int64) (interface{}, error) {
	ctx := remote.GetHeightRequestContext(r.src.Ctx, height)
	req := &psetypes.QueryScheduledDistributionsRequest{}
	res, err := r.queryClient.ScheduledDistributions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error querying scheduled distributions: %w", err)
	}
	return res, nil
}

func (r *remoteSource) ClearingAccountBalances(height int64) (interface{}, error) {
	ctx := remote.GetHeightRequestContext(r.src.Ctx, height)
	req := &psetypes.QueryClearingAccountBalancesRequest{}
	res, err := r.queryClient.ClearingAccountBalances(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error querying clearing account balances: %w", err)
	}
	return res, nil
}

func (r *remoteSource) Params(height int64) (interface{}, error) {
	ctx := remote.GetHeightRequestContext(r.src.Ctx, height)
	req := &psetypes.QueryParamsRequest{}
	res, err := r.queryClient.Params(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("error querying PSE params: %w", err)
	}
	return res, nil
}
