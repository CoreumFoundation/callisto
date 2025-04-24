package dex

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/dex"
	modulestypes "github.com/forbole/callisto/v4/modules/types"
	"github.com/forbole/callisto/v4/utils"
	parsecmdtypes "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/forbole/juno/v6/types/config"
	"github.com/spf13/cobra"
)

func paramsCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Get the current parameters of the dex module",
		RunE: func(cmd *cobra.Command, args []string) error {
			parseCtx, err := parsecmdtypes.GetParserContext(config.Cfg, parseConfig)
			if err != nil {
				return err
			}

			cdc := utils.GetCodec()
			sources, err := modulestypes.BuildSources(config.Cfg.Node, cdc)
			if err != nil {
				return err
			}

			// Get the database
			db := database.Cast(parseCtx.Database)
			dexModule := buildDEXModule(sources, cdc, db)
			height, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return err
			}

			return dexModule.UpdateParams(height)
		},
	}
}

func buildDEXModule(
	sources *modulestypes.Sources,
	cdc codec.Codec,
	db *database.Db,
) *dex.Module {
	// Build the dex module
	dexModule := dex.NewModule(
		sources.DEXSource,
		cdc,
		db,
	)
	return dexModule
}
