package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	assetftsource "github.com/forbole/callisto/v4/modules/assetft/source"
	remoteassetftsource "github.com/forbole/callisto/v4/modules/assetft/source/remote"
	assetnftsource "github.com/forbole/callisto/v4/modules/assetnft/source"
	remoteassetnftsource "github.com/forbole/callisto/v4/modules/assetnft/source/remote"
	authsource "github.com/forbole/callisto/v4/modules/auth/source"
	remoteauthsource "github.com/forbole/callisto/v4/modules/auth/source/remote"
	banksource "github.com/forbole/callisto/v4/modules/bank/source"
	localbanksource "github.com/forbole/callisto/v4/modules/bank/source/local"
	remotebanksource "github.com/forbole/callisto/v4/modules/bank/source/remote"
	customparamssource "github.com/forbole/callisto/v4/modules/customparams/source"
	remotecustomparamssource "github.com/forbole/callisto/v4/modules/customparams/source/remote"
	dexsource "github.com/forbole/callisto/v4/modules/dex/source"
	remotedexsource "github.com/forbole/callisto/v4/modules/dex/source/remote"
	distrsource "github.com/forbole/callisto/v4/modules/distribution/source"
	localdistrsource "github.com/forbole/callisto/v4/modules/distribution/source/local"
	remotedistrsource "github.com/forbole/callisto/v4/modules/distribution/source/remote"
	feemodelsource "github.com/forbole/callisto/v4/modules/feemodel/source"
	remotefeemodelsource "github.com/forbole/callisto/v4/modules/feemodel/source/remote"
	govsource "github.com/forbole/callisto/v4/modules/gov/source"
	localgovsource "github.com/forbole/callisto/v4/modules/gov/source/local"
	remotegovsource "github.com/forbole/callisto/v4/modules/gov/source/remote"
	mintsource "github.com/forbole/callisto/v4/modules/mint/source"
	localmintsource "github.com/forbole/callisto/v4/modules/mint/source/local"
	remotemintsource "github.com/forbole/callisto/v4/modules/mint/source/remote"
	slashingsource "github.com/forbole/callisto/v4/modules/slashing/source"
	localslashingsource "github.com/forbole/callisto/v4/modules/slashing/source/local"
	remoteslashingsource "github.com/forbole/callisto/v4/modules/slashing/source/remote"
	stakingsource "github.com/forbole/callisto/v4/modules/staking/source"
	localstakingsource "github.com/forbole/callisto/v4/modules/staking/source/local"
	remotestakingsource "github.com/forbole/callisto/v4/modules/staking/source/remote"
	"github.com/forbole/callisto/v4/utils/simapp"
	nodeconfig "github.com/forbole/juno/v6/node/config"
	"github.com/forbole/juno/v6/node/local"
	"github.com/forbole/juno/v6/node/remote"

	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/nft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v5/x/customparams/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
	feemodeltypes "github.com/CoreumFoundation/coreum/v5/x/feemodel/types"
)

type Sources struct {
	AuthSource         authsource.Source
	BankSource         banksource.Source
	DistrSource        distrsource.Source
	GovSource          govsource.Source
	MintSource         mintsource.Source
	SlashingSource     slashingsource.Source
	StakingSource      stakingsource.Source
	FeeModelSource     feemodelsource.Source
	CustomParamsSource customparamssource.Source
	AssetFTSource      assetftsource.Source
	AssetNFTSource     assetnftsource.Source
	DEXSource          dexsource.Source
}

func BuildSources(nodeCfg nodeconfig.Config, cdc codec.Codec) (*Sources, error) {
	switch cfg := nodeCfg.Details.(type) {
	case *remote.Details:
		return buildRemoteSources(cfg, cdc)
	case *local.Details:
		return buildLocalSources(cfg, cdc)

	default:
		return nil, fmt.Errorf("invalid configuration type: %T", cfg)
	}
}

func buildLocalSources(cfg *local.Details, cdc codec.Codec) (*Sources, error) {
	source, err := local.NewSource(cfg.Home, cdc)
	if err != nil {
		return nil, err
	}

	app := simapp.NewSimApp(cdc)

	sources := &Sources{
		BankSource:     localbanksource.NewSource(source, banktypes.QueryServer(app.BankKeeper)),
		DistrSource:    localdistrsource.NewSource(source, distrkeeper.NewQuerier(app.DistrKeeper)),
		GovSource:      localgovsource.NewSource(source, govkeeper.NewQueryServer(&app.GovKeeper)),
		MintSource:     localmintsource.NewSource(source, mintkeeper.NewQueryServerImpl(app.MintKeeper)),
		SlashingSource: localslashingsource.NewSource(source, slashingtypes.QueryServer(app.SlashingKeeper)),
		StakingSource:  localstakingsource.NewSource(source, stakingkeeper.Querier{Keeper: app.StakingKeeper}),
	}

	// Mount and initialize the stores
	err = source.MountKVStores(app, "keys")
	if err != nil {
		return nil, err
	}

	err = source.InitStores()
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func buildRemoteSources(cfg *remote.Details, cdc codec.Codec) (*Sources, error) {
	source, err := remote.NewSource(cfg.GRPC, cdc)
	if err != nil {
		return nil, fmt.Errorf("error while creating remote source: %s", err)
	}

	return &Sources{
		AuthSource:         remoteauthsource.NewSource(source, authtypes.NewQueryClient(source.GrpcConn)),
		BankSource:         remotebanksource.NewSource(source, banktypes.NewQueryClient(source.GrpcConn)),
		DistrSource:        remotedistrsource.NewSource(source, distrtypes.NewQueryClient(source.GrpcConn)),
		GovSource:          remotegovsource.NewSource(source, govtypesv1.NewQueryClient(source.GrpcConn)),
		MintSource:         remotemintsource.NewSource(source, minttypes.NewQueryClient(source.GrpcConn)),
		SlashingSource:     remoteslashingsource.NewSource(source, slashingtypes.NewQueryClient(source.GrpcConn)),
		StakingSource:      remotestakingsource.NewSource(source, stakingtypes.NewQueryClient(source.GrpcConn)),
		FeeModelSource:     remotefeemodelsource.NewSource(source, feemodeltypes.NewQueryClient(source.GrpcConn)),
		CustomParamsSource: remotecustomparamssource.NewSource(source, customparamstypes.NewQueryClient(source.GrpcConn)),
		AssetFTSource:      remoteassetftsource.NewSource(source, assetfttypes.NewQueryClient(source.GrpcConn)),
		AssetNFTSource:     remoteassetnftsource.NewSource(source, assetnfttypes.NewQueryClient(source.GrpcConn)),
		DEXSource:          remotedexsource.NewSource(source, dextypes.NewQueryClient(source.GrpcConn)),
	}, nil
}
