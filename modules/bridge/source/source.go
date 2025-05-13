package source

type Source interface {
	GetSendToXRPLOperationID(
		contractAddress string,
		recipient string,
		height uint64,
	) (uint32, error)
}
