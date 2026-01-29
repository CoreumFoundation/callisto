package remote

import (
	dexsource "github.com/forbole/callisto/v4/modules/dex/source"
	"github.com/forbole/juno/v6/node/remote"

	dextypes "github.com/tokenize-x/tx-chain/v6/x/dex/types"
)

var _ dexsource.Source = &Source{}

// Source implements dexsource.Source using a remote node
type Source struct {
	*remote.Source
	dexClient dextypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, dexClient dextypes.QueryClient) *Source {
	return &Source{
		Source:    source,
		dexClient: dexClient,
	}
}

// GetParams implements dexsource.Source
func (s Source) GetParams(height int64) (dextypes.Params, error) {
	res, err := s.dexClient.Params(remote.GetHeightRequestContext(s.Ctx, height), &dextypes.QueryParamsRequest{})
	if err != nil {
		return dextypes.Params{}, err
	}

	return res.Params, nil
}
