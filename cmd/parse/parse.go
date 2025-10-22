package parse

import (
	parseauth "github.com/forbole/callisto/v4/cmd/parse/auth"
	parsebank "github.com/forbole/callisto/v4/cmd/parse/bank"
	parsebridge "github.com/forbole/callisto/v4/cmd/parse/bridge"
	parsedex "github.com/forbole/callisto/v4/cmd/parse/dex"
	parsedistribution "github.com/forbole/callisto/v4/cmd/parse/distribution"
	parsefeegrant "github.com/forbole/callisto/v4/cmd/parse/feegrant"
	parsegov "github.com/forbole/callisto/v4/cmd/parse/gov"
	parsemint "github.com/forbole/callisto/v4/cmd/parse/mint"
	parsepricefeed "github.com/forbole/callisto/v4/cmd/parse/pricefeed"
	parsestaking "github.com/forbole/callisto/v4/cmd/parse/staking"
	parseblocks "github.com/forbole/juno/v6/cmd/parse/blocks"
	parsegenesis "github.com/forbole/juno/v6/cmd/parse/genesis"
	parsetransaction "github.com/forbole/juno/v6/cmd/parse/transactions"
	parse "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/forbole/juno/v6/modules/messages"
	"github.com/spf13/cobra"
)

// NewParseCmd returns the Cobra command allowing to parse some chain data without having to re-sync the whole database
func NewParseCmd(parseCfg *parse.Config, parser messages.MessageAddressesParser) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "parse",
		Short:             "Parse some data without the need to re-syncing the whole database from scratch",
		PersistentPreRunE: runPersistentPreRuns(parse.ReadConfigPreRunE(parseCfg)),
	}

	cmd.AddCommand(
		parseauth.NewAuthCmd(parseCfg),
		parsebank.NewBankCmd(parseCfg),
		parseblocks.NewBlocksCmd(parseCfg),
		parsedistribution.NewDistributionCmd(parseCfg),
		parsefeegrant.NewFeegrantCmd(parseCfg),
		parsegenesis.NewGenesisCmd(parseCfg),
		parsegov.NewGovCmd(parseCfg, parser),
		parsemint.NewMintCmd(parseCfg),
		parsepricefeed.NewPricefeedCmd(parseCfg),
		parsestaking.NewStakingCmd(parseCfg),
		parsetransaction.NewTransactionsCmd(parseCfg),
		parsedex.NewDexCmd(parseCfg),
		parsebridge.NewBridgeCmd(parseCfg),
	)

	return cmd
}

func runPersistentPreRuns(preRun func(_ *cobra.Command, _ []string) error) func(_ *cobra.Command, _ []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if root := cmd.Root(); root != nil {
			if root.PersistentPreRunE != nil {
				err := root.PersistentPreRunE(root, args)
				if err != nil {
					return err
				}
			}
		}

		return preRun(cmd, args)
	}
}
