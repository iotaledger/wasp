package eventlog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func AppendToLog(state kv.KVStore, ts int64, contract coretypes.Hname, data []byte) {
	collections.NewTimestampedLog(state, kv.Key(contract.Bytes())).MustAppend(ts, data)
}
