package chainlog

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

func AppendToChainLog(state kv.KVStore, ts int64, contract coretypes.Hname, recType byte, data []byte) {
	tlog := datatypes.NewMustTimestampedLog(state, ChainLogName(contract, recType))
	tlog.Append(ts, data)
}

func ChainLogName(contract coretypes.Hname, recType byte) kv.Key {
	var buf bytes.Buffer
	buf.Write(contract.Bytes())
	buf.WriteByte(recType)
	return kv.Key(buf.Bytes())
}
