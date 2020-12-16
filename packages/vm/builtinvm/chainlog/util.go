package chainlog

import (
	"bytes"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

func AppendToChainlog(state kv.KVStore, ts int64, contract coretypes.Hname, recType byte, data []byte) {
	tlog, _ := GetOrCreateDataLog(state, contract, recType)
	tlog.Append(ts, data)
}

func GetOrCreateDataLog(state kv.KVStore, contract coretypes.Hname, recType byte) (*datatypes.MustTimestampedLog, bytes.Buffer) {
	var buf bytes.Buffer
	buf.Write(contract.Bytes())
	buf.WriteByte(recType)
	return datatypes.NewMustTimestampedLog(state, kv.Key(buf.Bytes())), buf
}
