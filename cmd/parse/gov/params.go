package gov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/assetft"
	"github.com/forbole/callisto/v4/modules/assetnft"
	"github.com/forbole/callisto/v4/modules/auth"
	"github.com/forbole/callisto/v4/modules/customparams"
	"github.com/forbole/callisto/v4/modules/dex"
	"github.com/forbole/callisto/v4/modules/distribution"
	"github.com/forbole/callisto/v4/modules/feemodel"
	"github.com/forbole/callisto/v4/modules/gov"
	"github.com/forbole/callisto/v4/modules/group"
	"github.com/forbole/callisto/v4/modules/mint"
	"github.com/forbole/callisto/v4/modules/slashing"
	"github.com/forbole/callisto/v4/modules/staking"
	modulestypes "github.com/forbole/callisto/v4/modules/types"
	"github.com/forbole/callisto/v4/utils"
	parsecmdtypes "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/forbole/juno/v6/modules/messages"
	"github.com/forbole/juno/v6/types/config"
	"github.com/spf13/cobra"
)

func paramsCmd(parseConfig *parsecmdtypes.Config, messageAddressParser messages.MessageAddressesParser) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Get the current parameters of the gov module",
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
			govModule := buildGovModule(sources, messageAddressParser, cdc, db)
			height, err := parseCtx.Node.LatestHeight()
			if err != nil {
				return err
			}

			return govModule.UpdateParams(height)
		},
	}
}

func buildGovModule(
	sources *modulestypes.Sources,
	messageAddressParser messages.MessageAddressesParser,
	cdc codec.Codec,
	db *database.Db,
) *gov.Module {
	// Build expected modules of gov modules for handleParamChangeProposal
	authModule := auth.NewModule(sources.AuthSource, messageAddressParser, cdc, db)
	distrModule := distribution.NewModule(sources.DistrSource, cdc, db)
	mintModule := mint.NewModule(sources.MintSource, cdc, db)
	slashingModule := slashing.NewModule(sources.SlashingSource, cdc, db)
	stakingModule := staking.NewModule(sources.StakingSource, cdc, db)
	feeModelModule := feemodel.NewModule(sources.FeeModelSource, cdc, db)
	customParamsModule := customparams.NewModule(sources.CustomParamsSource, cdc, db)
	assetFTModule := assetft.NewModule(sources.AssetFTSource, cdc, db)
	assetNFTModule := assetnft.NewModule(sources.AssetNFTSource, cdc, db)
	dexModule := dex.NewModule(sources.DEXSource, cdc, db)
	groupModule := group.NewModule(sources.GroupSource, cdc, db)

	// Build the gov module
	govModule := gov.NewModule(
		sources.GovSource,
		authModule,
		distrModule,
		mintModule,
		slashingModule,
		stakingModule,
		feeModelModule,
		customParamsModule,
		assetFTModule,
		assetNFTModule,
		dexModule,
		groupModule,
		cdc,
		db,
	)
	return govModule
}
