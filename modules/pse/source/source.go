package source

// Source defines the methods exposed by a PSE source implementation.
// Currently it's a very small interface used by action handlers to query
// delegator scores. Implementations (local/remote) should satisfy this.
type Source interface {
	// DelegatorScore returns the score for the given delegator at the provided height.
	// Implementations may return an error if the query/path is not supported.
	DelegatorScore(height int64, address string) (string, error)

	// ScheduledDistributions returns future distribution schedules.
	ScheduledDistributions(height int64) (interface{}, error)

	// ClearingAccountBalances returns balances of clearing accounts.
	ClearingAccountBalances(height int64) (interface{}, error)

	// Params returns the PSE module parameters.
	Params(height int64) (interface{}, error)
}
