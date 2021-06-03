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
func GetRequestIDsForLastBlock(stateReader state.OptimisticStateReader) ([]coretypes.RequestID, error) {
	blockIndex, err := stateReader.BlockIndex()
	if err != nil {
		return nil, err
	}
	if blockIndex == 0 {
		return nil, nil
	}
	partition := subrealm.NewReadOnly(stateReader.KVStoreReader(), kv.Key(Interface.Hname().Bytes()))
	a := assert.NewAssert()
	recsBin, exist, err := getRequestLogRecordsForBlockBin(partition, blockIndex, a)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	ret := make([]coretypes.RequestID, len(recsBin))
	for i, d := range recsBin {
		rec, err := RequestLogRecordFromBytes(d)
		a.RequireNoError(err)
		ret[i] = rec.RequestID
	}
	return ret, nil
}

// IsRequestProcessed check if reqid is stored in the chain state as processed
func IsRequestProcessed(stateReader state.StateReader, reqid *coretypes.RequestID) bool {
	partition := subrealm.NewReadOnly(stateReader.KVStoreReader(), kv.Key(Interface.Hname().Bytes()))
	ret, err := isRequestProcessedIntern(partition, reqid)
	assert.NewAssert().RequireNoError(err)
	return ret
}
