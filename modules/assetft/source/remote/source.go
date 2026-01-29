package remote

import (
	assetftsource "github.com/forbole/callisto/v4/modules/assetft/source"
	"github.com/forbole/juno/v6/node/remote"

	assetfttypes "github.com/tokenize-x/tx-chain/v6/x/asset/ft/types"
)

var _ assetftsource.Source = &Source{}

// Source implements assetftsource.Source using a remote node
type Source struct {
	*remote.Source
	assetftClient assetfttypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, assetftClient assetfttypes.QueryClient) *Source {
	return &Source{
		Source:        source,
		assetftClient: assetftClient,
	}
}

// GetParams implements assetftsource.Source
func (s Source) GetParams(height int64) (assetfttypes.Params, error) {
	res, err := s.assetftClient.Params(remote.GetHeightRequestContext(s.Ctx, height), &assetfttypes.QueryParamsRequest{})
	if err != nil {
		return assetfttypes.Params{}, err
	}

	return res.Params, nil
}

// GetDexSettings implements assetftsource.Source
// GetDexSettings retrieves the DEX settings for the asset FT module at a specific height
func (s Source) GetDexSettings(height int64, denom string) (assetfttypes.DEXSettings, error) {
	res, err := s.assetftClient.DEXSettings(remote.GetHeightRequestContext(s.Ctx, height), &assetfttypes.QueryDEXSettingsRequest{
		Denom: denom,
	})
	if err != nil {
		return assetfttypes.DEXSettings{}, err
	}

	return res.DEXSettings, nil
}
