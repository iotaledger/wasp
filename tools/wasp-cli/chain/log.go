package chain

import (
	"github.com/spf13/cobra"
)

var logCmd = &cobra.Command{
	Use:   "log <name>",
	Short: "Show log of contract <name>",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// TODO refactor to use blocklog
		// r, err := SCClient(eventlog.Contract.Hname()).CallView(eventlog.FuncGetRecords.Name,
		// 	dict.Dict{
		// 		eventlog.ParamContractHname: iscp.Hn(args[0]).Bytes(),
		// 	})
		// log.Check(err)

		// records := collections.NewArray16ReadOnly(r, eventlog.ParamRecords)
		// for i := uint16(0); i < records.MustLen(); i++ {
		// 	b := records.MustGetAt(i)
		// 	rec, err := collections.ParseRawLogRecord(b)
		// 	log.Check(err)
		// 	log.Printf("%s %s\n", time.Unix(0, rec.Timestamp), string(rec.Data))
		// }
	},
}
