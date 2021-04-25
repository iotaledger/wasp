package chain

import (
	"time"

	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var logCmd = &cobra.Command{
	Use:   "log <name>",
	Short: "Show log of contract <name>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		r, err := SCClient(eventlog.Interface.Hname()).CallView(eventlog.FuncGetRecords,
			dict.Dict{
				eventlog.ParamContractHname: coretypes.Hn(args[0]).Bytes(),
			})
		log.Check(err)

		records := collections.NewArray16ReadOnly(r, eventlog.ParamRecords)
		for i := uint16(0); i < records.MustLen(); i++ {
			b := records.MustGetAt(i)
			rec, err := collections.ParseRawLogRecord(b)
			log.Check(err)
			log.Printf("%s %s\n", time.Unix(0, rec.Timestamp), string(rec.Data))
		}
	},
}
