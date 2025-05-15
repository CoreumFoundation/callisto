package source

type Source interface {
	GetOutgoingPendingOperationID(
		contractAddress string,
		recipient string,
		height uint64,
	) (uint32, error)
}
