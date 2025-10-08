package bridge

import (
	parsecmdtypes "github.com/forbole/juno/v6/cmd/parse/types"
	"github.com/spf13/cobra"
)

// NewBridgeCmd returns the Cobra command that allows to fix all the things related to the bridge module
func NewBridgeCmd(parseConfig *parsecmdtypes.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridge",
		Short: "Fix things related to the bridge module",
	}

	cmd.AddCommand(
		txsCmd(parseConfig),
	)

	return cmd
}
