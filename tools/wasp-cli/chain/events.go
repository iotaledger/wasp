package chain

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events <name>",
	Short: "Show events of contract <name>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.ViewGetEventsForContract.Name, dict.Dict{
			blocklog.ParamContractHname: iscp.Hn(args[0]).Bytes(),
		})
		log.Check(err)
		logEvents(r)
	},
}
