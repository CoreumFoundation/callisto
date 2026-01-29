package remote

import (
	feemodelsource "github.com/forbole/callisto/v4/modules/feemodel/source"
	"github.com/forbole/juno/v6/node/remote"

	feemodeltypes "github.com/tokenize-x/tx-chain/v6/x/feemodel/types"
)

var _ feemodelsource.Source = &Source{}

// Source implements feemodelsource.Source using a remote node
type Source struct {
	*remote.Source
	feemodelClient feemodeltypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, feemodelClient feemodeltypes.QueryClient) *Source {
	return &Source{
		Source:         source,
		feemodelClient: feemodelClient,
	}
}

// GetParams implements feemodelsource.Source
func (s Source) GetParams(height int64) (feemodeltypes.Params, error) {
	res, err := s.feemodelClient.Params(remote.GetHeightRequestContext(s.Ctx, height), &feemodeltypes.QueryParamsRequest{})
	if err != nil {
		return feemodeltypes.Params{}, err
	}

	return res.Params, nil
}
