package source

type Source interface {
	GetSendToXRPLOperationIDs(
		contractAddress string,
		recipient string,
		height uint64,
	) ([]uint32, error)
}
