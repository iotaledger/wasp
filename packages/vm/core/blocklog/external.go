package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
)

// GetRequestIDsForLastBlock reads blocklog from chain state and returns request IDs settled in specific block
// Can only panic on DB error of internal error
func GetRequestIDsForLastBlock(stateReader state.StateReader) []coretypes.RequestID {
	blockIndex := stateReader.BlockIndex()
	if blockIndex == 0 {
		return nil
	}
	partition := subrealm.NewReadOnly(stateReader.KVStoreReader(), kv.Key(Interface.Hname().Bytes()))
	a := assert.NewAssert()
	recsBin, exist := getRequestLogRecordsForBlockBin(partition, blockIndex, a)
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

// IsRequestProcessed check if reqid is stored in the chain state as processed
func IsRequestProcessed(stateReader state.StateReader, reqid *coretypes.RequestID) bool {
	partition := subrealm.NewReadOnly(stateReader.KVStoreReader(), kv.Key(Interface.Hname().Bytes()))
	ret, err := isRequestProcessedIntern(partition, reqid)
	assert.NewAssert().RequireNoError(err)
	return ret
}
