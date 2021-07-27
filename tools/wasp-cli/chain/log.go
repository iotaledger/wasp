package chain

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "events <name>",
	Short: "Show events of contract <name>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, err := SCClient(blocklog.Contract.Hname()).CallView(blocklog.FuncGetEventsForContract.Name,
			dict.Dict{
				blocklog.ParamContractHname: iscp.Hn(args[0]).Bytes(),
			})
		log.Check(err)

		records := collections.NewArray16ReadOnly(r, blocklog.ParamEvent)
		for i := uint16(0); i < records.MustLen(); i++ {
			b := records.MustGetAt(i)
			log.Check(err)
			log.Printf("%s\n", string(b))
		}
	},
}
