package chainlog

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

func AppendToChainlog(state kv.KVStore, ts int64, contract coretypes.Hname, recType byte, data []byte) {
	var buf bytes.Buffer
	buf.Write(contract.Bytes())
	buf.WriteByte(recType)
	tlog := datatypes.NewMustTimestampedLog(state, kv.Key(buf.Bytes()))
	tlog.Append(ts, data)
}
