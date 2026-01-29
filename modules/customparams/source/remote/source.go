package remote

import (
	customparamssource "github.com/forbole/callisto/v4/modules/customparams/source"
	"github.com/forbole/juno/v6/node/remote"

	customparamstypes "github.com/tokenize-x/tx-chain/v6/x/customparams/types"
)

var _ customparamssource.Source = &Source{}

// Source implements customparamssource.Source using a remote node
type Source struct {
	*remote.Source
	customparamsClient customparamstypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, customparamsClient customparamstypes.QueryClient) *Source {
	return &Source{
		Source:             source,
		customparamsClient: customparamsClient,
	}
}

// GetParams implements customparamssource.Source
func (s Source) GetParams(height int64) (customparamstypes.StakingParams, error) {
	res, err := s.customparamsClient.StakingParams(remote.GetHeightRequestContext(s.Ctx, height), &customparamstypes.QueryStakingParamsRequest{})
	if err != nil {
		return customparamstypes.StakingParams{}, err
	}

	return res.Params, nil
}
