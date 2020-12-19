package chainlog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

func AppendToChainLog(state kv.KVStore, ts int64, contract coretypes.Hname, data []byte) {
	datatypes.NewMustTimestampedLog(state, kv.Key(contract.Bytes())).Append(ts, data)
}
