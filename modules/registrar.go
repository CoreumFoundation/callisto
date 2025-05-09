package modules

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/forbole/callisto/v4/database"
	"github.com/forbole/callisto/v4/modules/actions"
	"github.com/forbole/callisto/v4/modules/addresses"
	"github.com/forbole/callisto/v4/modules/assetft"
	"github.com/forbole/callisto/v4/modules/assetnft"
	"github.com/forbole/callisto/v4/modules/auth"
	"github.com/forbole/callisto/v4/modules/bank"
	"github.com/forbole/callisto/v4/modules/bridge"
	"github.com/forbole/callisto/v4/modules/consensus"
	"github.com/forbole/callisto/v4/modules/customparams"
	dailyrefetch "github.com/forbole/callisto/v4/modules/daily_refetch"
	"github.com/forbole/callisto/v4/modules/dex"
	"github.com/forbole/callisto/v4/modules/distribution"
	"github.com/forbole/callisto/v4/modules/feegrant"
	"github.com/forbole/callisto/v4/modules/feemodel"
	"github.com/forbole/callisto/v4/modules/gov"
	messagetype "github.com/forbole/callisto/v4/modules/message_type"
	"github.com/forbole/callisto/v4/modules/mint"
	"github.com/forbole/callisto/v4/modules/modules"
	"github.com/forbole/callisto/v4/modules/pricefeed"
	"github.com/forbole/callisto/v4/modules/slashing"
	"github.com/forbole/callisto/v4/modules/staking"
	"github.com/forbole/callisto/v4/modules/types"
	"github.com/forbole/callisto/v4/modules/upgrade"
	"github.com/forbole/callisto/v4/utils"
	jmodules "github.com/forbole/juno/v6/modules"
	"github.com/forbole/juno/v6/modules/messages"
	"github.com/forbole/juno/v6/modules/pruning"
	"github.com/forbole/juno/v6/modules/registrar"
	"github.com/forbole/juno/v6/modules/telemetry"
	juno "github.com/forbole/juno/v6/types"
)

// UniqueAddressesParser returns a wrapper around the given parser that removes all duplicated addresses
func UniqueAddressesParser(parser messages.MessageAddressesParser) messages.MessageAddressesParser {
	return func(tx *juno.Transaction) ([]string, error) {
		addresses, err := parser(tx)
		if err != nil {
			return nil, err
		}

		return utils.RemoveDuplicateValues(addresses), nil
	}
}

// --------------------------------------------------------------------------------------------------------------------

var _ registrar.Registrar = &Registrar{}

// Registrar represents the modules.Registrar that allows to register all modules that are supported by BigDipper
type Registrar struct {
	parser messages.MessageAddressesParser
	cdc    codec.Codec
}

// NewRegistrar allows to build a new Registrar instance
func NewRegistrar(parser messages.MessageAddressesParser, cdc codec.Codec) *Registrar {
	return &Registrar{
		parser: UniqueAddressesParser(parser),
		cdc:    cdc,
	}
}

// BuildModules implements modules.Registrar
func (r *Registrar) BuildModules(ctx registrar.Context) jmodules.Modules {
	db := database.Cast(ctx.Database)

	sources, err := types.BuildSources(ctx.JunoConfig.Node, r.cdc)
	if err != nil {
		panic(err)
	}

	messagetypeModule := messagetype.NewModule(r.parser, r.cdc, db)
	actionsModule := actions.NewModule(ctx.JunoConfig, r.cdc, sources)
	authModule := auth.NewModule(sources.AuthSource, r.parser, r.cdc, db)
	bankModule := bank.NewModule(r.parser, sources.BankSource, r.cdc, db, ctx.JunoConfig.Chain.Bech32Prefix)
	consensusModule := consensus.NewModule(db)
	dailyRefetchModule := dailyrefetch.NewModule(ctx.Proxy, db)
	distrModule := distribution.NewModule(sources.DistrSource, r.cdc, db)
	feegrantModule := feegrant.NewModule(r.cdc, db)
	mintModule := mint.NewModule(sources.MintSource, r.cdc, db)
	slashingModule := slashing.NewModule(sources.SlashingSource, r.cdc, db)
	stakingModule := staking.NewModule(sources.StakingSource, r.cdc, db)
	feeModelModule := feemodel.NewModule(sources.FeeModelSource, r.cdc, db)
	customParamsModule := customparams.NewModule(sources.CustomParamsSource, r.cdc, db)
	assetFTModule := assetft.NewModule(sources.AssetFTSource, r.cdc, db)
	assetNFTModule := assetnft.NewModule(sources.AssetNFTSource, r.cdc, db)
	dexModule := dex.NewModule(sources.DEXSource, r.cdc, db)
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
		r.cdc,
		db,
	)
	upgradeModule := upgrade.NewModule(db, stakingModule)
	bridgeModule := bridge.NewModule(ctx.JunoConfig, sources.BridgeSource, r.cdc, db)

	return []jmodules.Module{
		messages.NewModule(r.parser, ctx.Database),
		telemetry.NewModule(ctx.JunoConfig),
		pruning.NewModule(ctx.JunoConfig, db, ctx.Logger),

		messagetypeModule,
		actionsModule,
		authModule,
		bankModule,
		consensusModule,
		dailyRefetchModule,
		distrModule,
		feegrantModule,
		govModule,
		mintModule,
		modules.NewModule(ctx.JunoConfig.Chain, db),
		pricefeed.NewModule(ctx.JunoConfig, r.cdc, db),
		slashingModule,
		stakingModule,
		upgradeModule,
		feeModelModule,
		customParamsModule,
		assetFTModule,
		assetNFTModule,
		dexModule,
		bridgeModule,
		// This must be the last item.
		addresses.NewModule(r.parser, r.cdc, db),
	}
}
