package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
)

// MustGetRequestIDsForLastBlock reads blocklog from chain state and returns request IDs settled in specific block
// Can only panic on DB error of internal error
func MustGetRequestIDsForLastBlock(chainState state.VirtualState) []coretypes.RequestID {
	if chainState.BlockIndex() == 0 {
		return nil
	}
	partition := subrealm.NewReadOnly(chainState.KVStore(), kv.Key(Interface.Hname().Bytes()))
	a := assert.NewAssert()
	recsBin, exist := getRequestLogRecordsForBlockBin(partition, chainState.BlockIndex(), a)
	if !exist {
		return nil
	}
	ret := make([]coretypes.RequestID, len(recsBin))
	for i, d := range recsBin {
		rec, err := RequestLogRecordFromBytes(d)
		a.RequireNoError(err)
		ret[i] = rec.RequestID
	}
	return ret
}
