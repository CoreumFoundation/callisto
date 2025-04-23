package dex

import (
	parsecmdtypes "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/spf13/cobra"
)

// NewDexCmd returns the Cobra command allowing to fix various things related to the x/dex module
func NewDexCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dex",
		Short: "Fix things related to the x/dex module",
	}

	cmd.AddCommand(
		paramsCmd(parseConfig),
	)

	return cmd
}
