package chain

import (
	"os"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func logCmd(args []string) {
	if len(args) != 1 {
		log.Fatal("Usage: %s chain log <name>", os.Args[0])
	}
	r, err := SCClient(chainlog.Interface.Hname()).CallView(chainlog.FuncGetLogRecords, codec.MakeDict(map[string]interface{}{
		chainlog.ParamContractHname: codec.EncodeHname(coretypes.Hn(args[0])),
	}))
	log.Check(err)

	records := datatypes.NewMustArray(r, chainlog.ParamRecords)
	for i := uint16(0); i < records.Len(); i++ {
		b := records.GetAt(i)
		rec, err := datatypes.ParseRawLogRecord(b)
		log.Check(err)
		log.Printf("%s %s\n", time.Unix(0, rec.Timestamp), string(rec.Data))
	}
}
