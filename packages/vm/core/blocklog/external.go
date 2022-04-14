package blocklog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"golang.org/x/xerrors"
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

type GetRequestReceiptResult struct {
	ReceiptBin   []byte
	BlockIndex   uint32
	RequestIndex uint16
}

// GetRequestRecordDataByRequestID tries to obtain the receipt data for a given request
// returns nil if receipt was not found
func GetRequestRecordDataByRequestID(stateReader kv.KVStoreReader, reqID iscp.RequestID) (*GetRequestReceiptResult, error) {
	lookupDigest := reqID.LookupDigest()
	lookupTable := collections.NewMapReadOnly(stateReader, prefixRequestLookupIndex)
	lookupKeyListBin := lookupTable.MustGetAt(lookupDigest[:])
	if lookupKeyListBin == nil {
		return nil, nil
	}
	lookupKeyList, err := RequestLookupKeyListFromBytes(lookupKeyListBin)
	if err != nil {
		return nil, err
	}
	for i := range lookupKeyList {
		recBin, found := getRequestRecordDataByRef(stateReader, lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex())
		if !found {
			return nil, xerrors.Errorf("inconsistency: request log record wasn't found by exact reference")
		}
		rec, err := RequestReceiptFromBytes(recBin)
		if err != nil {
			return nil, err
		}
		if rec.Request.ID() == reqID {
			return &GetRequestReceiptResult{
				ReceiptBin:   recBin,
				BlockIndex:   lookupKeyList[i].BlockIndex(),
				RequestIndex: lookupKeyList[i].RequestIndex(),
			}, nil
		}
	}
	return nil, nil
}
