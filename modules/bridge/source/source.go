package source

type Source interface {
	GetSendToXRPLOperationIDs(
		recipient string,
		height uint64,
	) ([]uint32, error)
}
