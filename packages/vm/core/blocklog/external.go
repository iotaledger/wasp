package blocklog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
)

// GetRequestIDsForLastBlock reads blocklog from chain state and returns request IDs settled in specific block
// Can only panic on DB error of internal error
func GetRequestIDsForBlock(stateReader state.OptimisticStateReader, blockIndex uint32) ([]iscp.RequestID, error) {
	if blockIndex == 0 {
		return []iscp.RequestID{}, nil
	}
	partition := subrealm.NewReadOnly(stateReader.KVStoreReader(), kv.Key(Contract.Hname().Bytes()))

	recsBin, exist, err := getRequestLogRecordsForBlockBin(partition, blockIndex)
	if err != nil {
		return nil, err
	}
	if !exist {
		return []iscp.RequestID{}, fmt.Errorf("block index %v does not exist", blockIndex)
	}
	ret := make([]iscp.RequestID, len(recsBin))
	for i, d := range recsBin {
		rec, err := RequestReceiptFromBytes(d)
		if err != nil {
			panic(err)
		}
		ret[i] = rec.Request.ID()
	}
	return ret, nil
}

// IsRequestProcessed check if reqid is stored in the chain state as processed
func IsRequestProcessed(stateReader kv.KVStoreReader, reqid *iscp.RequestID) (bool, error) {
	partition := subrealm.NewReadOnly(stateReader, kv.Key(Contract.Hname().Bytes()))
	return isRequestProcessedInternal(partition, reqid)
}

func MustIsRequestProcessed(stateReader kv.KVStoreReader, reqid *iscp.RequestID) bool {
	ret, err := IsRequestProcessed(stateReader, reqid)
	if err != nil {
		panic(err)
	}
	return ret
}
