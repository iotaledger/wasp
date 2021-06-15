package chain

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var callViewCmd = &cobra.Command{
	Use:   "call-view <name> <funcname> [params]",
	Short: "Call a contract view function",
	Long:  "Call contract <name>, view function <funcname> with given params.",
	Args:  cobra.MinimumNArgs(2), //nolint:gomnd
	Run: func(cmd *cobra.Command, args []string) {
		r, err := SCClient(coretypes.Hn(args[0])).CallView(args[1], util.EncodeParams(args[2:]))
		log.Check(err)
		util.PrintDictAsJson(r)
	},
}
