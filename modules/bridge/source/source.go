package source

type Source interface {
	GetOutgoingPendingOperationSequence(
		contractAddress string,
		recipient string,
		height uint64,
	) (uint32, error)
}
