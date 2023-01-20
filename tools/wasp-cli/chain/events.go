package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initEventsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "events <name>",
		Short: "Show events of contract <name>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			r, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.ViewGetEventsForContract.Name, dict.Dict{
				blocklog.ParamContractHname: isc.Hn(args[0]).Bytes(),
			})
			log.Check(err)
			logEvents(r)
		},
	}
}
