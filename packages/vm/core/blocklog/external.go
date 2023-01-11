package blocklog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

// GetRequestIDsForLastBlock reads blocklog from chain state and returns request IDs settled in specific block
// Can only panic on DB error of internal error
func GetRequestIDsForBlock(stateReader kv.KVStoreReader, blockIndex uint32) ([]isc.RequestID, error) {
	if blockIndex == 0 {
		return []isc.RequestID{}, nil
	}
	partition := subrealm.NewReadOnly(stateReader, kv.Key(Contract.Hname().Bytes()))

	recsBin, exist, err := getRequestLogRecordsForBlockBin(partition, blockIndex)
	if err != nil {
		return nil, err
	}
	if !exist {
		return []isc.RequestID{}, fmt.Errorf("block index %v does not exist", blockIndex)
	}
	ret := make([]isc.RequestID, len(recsBin))
	for i, d := range recsBin {
		rec, err := RequestReceiptFromBytes(d)
		if err != nil {
			panic(err)
		}
		ret[i] = rec.Request.ID()
	}
	return ret, nil
}

func GetRequestReceipt(stateReader kv.KVStoreReader, requestID isc.RequestID) (*RequestReceipt, error) {
	partition := subrealm.NewReadOnly(stateReader, kv.Key(Contract.Hname().Bytes()))
	return IsRequestProcessedInternal(partition, requestID)
}

// IsRequestProcessed check if requestID is stored in the chain state as processed
func IsRequestProcessed(stateReader kv.KVStoreReader, requestID isc.RequestID) (bool, error) {
	requestReceipt, err := GetRequestReceipt(stateReader, requestID)
	if err != nil {
		return false, fmt.Errorf("cannot get request receipt: %w", err)
	}
	return requestReceipt != nil, nil
}

func MustIsRequestProcessed(stateReader kv.KVStoreReader, reqid isc.RequestID) bool {
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
func GetRequestRecordDataByRequestID(stateReader kv.KVStoreReader, reqID isc.RequestID) (*GetRequestReceiptResult, error) {
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
			return nil, errors.New("inconsistency: request log record wasn't found by exact reference")
		}
		rec, err := RequestReceiptFromBytes(recBin)
		if err != nil {
			return nil, err
		}
		if rec.Request.ID().Equals(reqID) {
			return &GetRequestReceiptResult{
				ReceiptBin:   recBin,
				BlockIndex:   lookupKeyList[i].BlockIndex(),
				RequestIndex: lookupKeyList[i].RequestIndex(),
			}, nil
		}
	}
	return nil, nil
}

func GetEventsByBlockIndex(partition kv.KVStoreReader, blockIndex uint32, totalRequests uint16) ([]string, error) {
	ret := make([]string, 0)
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	for reqIdx := uint16(0); reqIdx < totalRequests; reqIdx++ {
		eventIndex := uint16(0)
		for {
			key := NewEventLookupKey(blockIndex, reqIdx, eventIndex)
			msg, err := events.GetAt(key.Bytes())
			if err != nil {
				return nil, err
			}
			if msg == nil {
				break
			}
			ret = append(ret, string(msg))
			eventIndex++
		}
	}
	return ret, nil
}

func GetBlockInfo(partition kv.KVStoreReader, blockIndex uint32) (*BlockInfo, error) {
	if blockIndex == 0 {
		return nil, nil
	}
	data, err := getBlockInfoBytes(partition, blockIndex)
	if err != nil {
		return nil, err
	}
	ret, err := BlockInfoFromBytes(blockIndex, data)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
