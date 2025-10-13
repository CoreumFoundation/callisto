package utils

import (
	"fmt"
	"strings"
	"time"

	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	"github.com/forbole/juno/v6/node"
	"github.com/rs/zerolog/log"
)

// isRetryableError checks if an error is retryable (timeout, gateway error, etc.)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	retryableErrors := []string{
		"504 Gateway Time-out",
		"503 Service Unavailable",
		"502 Bad Gateway",
		"timeout",
		"connection refused",
		"connection reset",
		"EOF",
		"invalid character '<'", // HTML error pages
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errMsg, retryable) {
			return true
		}
	}

	return false
}

// QueryTxs queries all the transactions from the given node corresponding to the given query
// with retry logic for transient errors like timeouts
func QueryTxs(node node.Node, query string) ([]*coretypes.ResultTx, error) {
	var txs []*coretypes.ResultTx

	page := 1
	perPage := 100
	stop := false

	const maxRetries = 10
	const initialBackoff = 1 * time.Second
	const maxBackoff = 1024 * time.Second

	for !stop {
		var result *coretypes.ResultTxSearch
		var err error

		// Retry logic with exponential backoff
		for attempt := 0; attempt < maxRetries; attempt++ {
			result, err = node.TxSearch(query, &page, &perPage, "")

			if err == nil {
				// Success, break out of retry loop
				break
			}

			// Check if error is retryable
			if !isRetryableError(err) {
				// Non-retryable error, return immediately
				return nil, fmt.Errorf("error while running tx search: %s", err)
			}

			// Don't wait after the last attempt
			if attempt == maxRetries-1 {
				break
			}

			// Calculate backoff duration with exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s
			backoff := initialBackoff * time.Duration(1<<uint(attempt))
			if backoff > maxBackoff {
				backoff = maxBackoff
			}

			log.Warn().
				Int("attempt", attempt+1).
				Int("max_retries", maxRetries).
				Dur("backoff", backoff).
				Err(err).
				Msg("retryable error in tx search, retrying after backoff")

			// Wait before retrying (exponential backoff)
			time.Sleep(backoff)
		}

		// If still error after all retries, return error
		if err != nil {
			return nil, fmt.Errorf("error while running tx search after %d retries: %s", maxRetries, err)
		}

		page++
		txs = append(txs, result.Txs...)
		stop = len(txs) == result.TotalCount
	}

	return txs, nil
}
