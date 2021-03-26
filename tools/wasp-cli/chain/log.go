package chain

import (
	"github.com/iotaledger/wasp/packages/kv/dict"
	"os"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/eventlog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func logCmd(args []string) {
	if len(args) != 1 {
		log.Fatal("Usage: %s chain log <name>", os.Args[0])
	}
	r, err := SCClient(eventlog.Interface.Hname()).CallView(eventlog.FuncGetRecords,
		dict.Dict{
			eventlog.ParamContractHname: coretypes.Hn(args[0]).Bytes(),
		})
	log.Check(err)

	records := collections.NewArrayReadOnly(r, eventlog.ParamRecords)
	for i := uint16(0); i < records.MustLen(); i++ {
		b := records.MustGetAt(i)
		rec, err := collections.ParseRawLogRecord(b)
		log.Check(err)
		log.Printf("%s %s\n", time.Unix(0, rec.Timestamp), string(rec.Data))
	}
}
