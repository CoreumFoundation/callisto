package utils

import (
	"context"
	"fmt"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/forbole/juno/v6/node"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
)

// QueryTxs queries all the transactions from the given node corresponding to the given query
func QueryTxs(node node.Node, query string) ([]*coretypes.ResultTx, error) {
	var txs []*coretypes.ResultTx

	page := 1
	perPage := 100
	stop := false

	ctx := context.Background()
	retryDelay := 500 * time.Millisecond
	doTimeout := 60 * time.Second

	for !stop {
		var result *coretypes.ResultTxSearch
		var searchErr error

		// Retry logic for each page
		doCtx, doCtxCancel := context.WithTimeout(ctx, doTimeout)
		err := retry.Do(doCtx, retryDelay, func() error {
			result, searchErr = node.TxSearch(query, &page, &perPage, "")
			if searchErr != nil {
				return fmt.Errorf("error while running tx search: %s", searchErr)
			}
			return nil
		})
		doCtxCancel()

		if err != nil {
			return nil, err
		}

		page++
		txs = append(txs, result.Txs...)
		stop = len(txs) == result.TotalCount
	}

	return txs, nil
}
