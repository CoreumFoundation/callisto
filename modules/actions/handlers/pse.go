package handlers

import (
	"fmt"

	actionstypes "github.com/forbole/callisto/v4/modules/actions/types"
)

// PSEScoreHandler returns the delegator score for a given address at an optional height.
func PSEScoreHandler(ctx *actionstypes.Context, payload *actionstypes.Payload) (interface{}, error) {
	if ctx == nil || ctx.Sources == nil || ctx.Sources.PSESource == nil {
		return nil, fmt.Errorf("pse source not available")
	}

	addr := payload.GetAddress()
	if addr == "" {
		return nil, fmt.Errorf("address missing from payload")
	}

	height, err := ctx.GetHeight(payload)
	if err != nil {
		return nil, err
	}

	score, err := ctx.Sources.PSESource.DelegatorScore(height, addr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"address": addr,
		"height":  height,
		"score":   score,
	}, nil
}

// PSEScheduledDistributionsHandler returns future scheduled distributions.
func PSEScheduledDistributionsHandler(ctx *actionstypes.Context, payload *actionstypes.Payload) (interface{}, error) {
	if ctx == nil || ctx.Sources == nil || ctx.Sources.PSESource == nil {
		return nil, fmt.Errorf("pse source not available")
	}

	height, err := ctx.GetHeight(payload)
	if err != nil {
		return nil, err
	}

	resp, err := ctx.Sources.PSESource.ScheduledDistributions(height)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PSEClearingAccountBalancesHandler returns balances of clearing accounts.
func PSEClearingAccountBalancesHandler(ctx *actionstypes.Context, payload *actionstypes.Payload) (interface{}, error) {
	if ctx == nil || ctx.Sources == nil || ctx.Sources.PSESource == nil {
		return nil, fmt.Errorf("pse source not available")
	}

	height, err := ctx.GetHeight(payload)
	if err != nil {
		return nil, err
	}

	resp, err := ctx.Sources.PSESource.ClearingAccountBalances(height)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PSEParamsHandler returns the PSE module parameters.
func PSEParamsHandler(ctx *actionstypes.Context, payload *actionstypes.Payload) (interface{}, error) {
	if ctx == nil || ctx.Sources == nil || ctx.Sources.PSESource == nil {
		return nil, fmt.Errorf("pse source not available")
	}

	height, err := ctx.GetHeight(payload)
	if err != nil {
		return nil, err
	}

	resp, err := ctx.Sources.PSESource.Params(height)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
