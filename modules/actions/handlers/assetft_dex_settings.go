package handlers

import (
	"fmt"

	"github.com/forbole/callisto/v4/modules/actions/types"
	"github.com/rs/zerolog/log"
)

func AssetFTDexSettingsHandler(ctx *types.Context, payload *types.Payload) (interface{}, error) {
	log.Debug().Str("denom", payload.GetDenom()).
		Int64("height", payload.Input.Height).
		Msg("executing assetft dex settings action")

	height, err := ctx.GetHeight(payload)
	if err != nil {
		return nil, err
	}

	dexSettings, err := ctx.Sources.AssetFTSource.GetDexSettings(height, payload.GetDenom())
	if err != nil {
		return nil, fmt.Errorf("error while getting DEX settings: %s", err)
	}

	return dexSettings, nil
}
