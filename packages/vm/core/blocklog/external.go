package blocklog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

func EventsFromViewResult(viewResult dict.Dict) (ret []*isc.Event, err error) {
	events := collections.NewArrayReadOnly(viewResult, ParamEvent)
	ret = make([]*isc.Event, events.Len())
	for i := range ret {
		ret[i], err = isc.EventFromBytes(events.GetAt(uint32(i)))
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}

func GetRequestsInBlock(partition kv.KVStoreReader, blockIndex uint32) (*BlockInfo, []isc.Request, error) {
	blockInfo, ok := GetBlockInfo(partition, blockIndex)
	if !ok {
		return nil, nil, fmt.Errorf("block not found: %d", blockIndex)
	}
	reqs := make([]isc.Request, blockInfo.TotalRequests)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		recBin, ok := getRequestRecordDataByRef(partition, blockIndex, reqIdx)
		if !ok {
			return nil, nil, fmt.Errorf("request not found: %d/%d", blockIndex, reqIdx)
		}
		rec, err := RequestReceiptFromBytes(recBin, blockIndex, reqIdx)
		if err != nil {
			return nil, nil, err
		}
		reqs[reqIdx] = rec.Request
	}
	return blockInfo, reqs, nil
}

// GetRequestIDsForBlock reads blocklog from chain state and returns request IDs settled in specific block
// Can only panic on DB error of internal error
func GetRequestIDsForBlock(stateReader kv.KVStoreReader, blockIndex uint32) ([]isc.RequestID, error) {
	if blockIndex == 0 {
		return []isc.RequestID{}, nil
	}
	partition := subrealm.NewReadOnly(stateReader, kv.Key(Contract.Hname().Bytes()))

	recsBin, exist := getRequestLogRecordsForBlockBin(partition, blockIndex)
	if !exist {
		return []isc.RequestID{}, fmt.Errorf("block index %v does not exist", blockIndex)
	}
	ret := make([]isc.RequestID, len(recsBin))
	for i, d := range recsBin {
		rec, err := RequestReceiptFromBytes(d, blockIndex, uint16(i))
		if err != nil {
			panic(err)
		}
		ret[i] = rec.Request.ID()
	}
	return ret, nil
}

// GetRequestReceipt returns the receipt for the given request, or nil if not found
func GetRequestReceipt(stateReader kv.KVStoreReader, requestID isc.RequestID) (*RequestReceipt, error) {
	partition := subrealm.NewReadOnly(stateReader, kv.Key(Contract.Hname().Bytes()))
	return getRequestReceipt(partition, requestID)
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
	lookupKeyListBin := lookupTable.GetAt(lookupDigest[:])
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
		rec, err := RequestReceiptFromBytes(recBin, lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex())
		if err != nil {
			return nil, err
		}
		if rec.Request.ID().Equals(reqID) {
			return &GetRequestReceiptResult{
				ReceiptBin:   recBin,
				BlockIndex:   rec.BlockIndex,
				RequestIndex: rec.RequestIndex,
			}, nil
		}
	}
	return nil, nil
}

func GetEventsByBlockIndex(partition kv.KVStoreReader, blockIndex uint32, totalRequests uint16) [][]byte {
	var ret [][]byte
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	for reqIdx := uint16(0); reqIdx < totalRequests; reqIdx++ {
		eventIndex := uint16(0)
		for {
			key := NewEventLookupKey(blockIndex, reqIdx, eventIndex).Bytes()
			eventData := events.GetAt(key)
			if eventData == nil {
				break
			}
			ret = append(ret, eventData)
			eventIndex++
		}
	}
	return ret
}

func GetBlockInfo(partition kv.KVStoreReader, blockIndex uint32) (*BlockInfo, bool) {
	data := getBlockInfoBytes(partition, blockIndex)
	if data == nil {
		return nil, false
	}
	ret, err := BlockInfoFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret, true
}

func Prune(partition kv.KVStore, latestBlockIndex uint32, blockKeepAmount int32) {
	if blockKeepAmount <= 0 {
		// keep all blocks
		return
	}
	if latestBlockIndex < uint32(blockKeepAmount) {
		return
	}
	toDelete := latestBlockIndex - uint32(blockKeepAmount)
	// assume that all blocks prior to `toDelete` have been already deleted, so
	// we only need to delete this one.
	pruneBlock(partition, toDelete)
}

func ReceiptsFromViewCallResult(res dict.Dict) ([]*RequestReceipt, error) {
	receipts := collections.NewArrayReadOnly(res, ParamRequestRecord)
	ret := make([]*RequestReceipt, receipts.Len())
	var err error
	blockIndex, err := codec.DecodeUint32(res.Get(ParamBlockIndex))
	if err != nil {
		return nil, err
	}

	for i := range ret {
		ret[i], err = RequestReceiptFromBytes(receipts.GetAt(uint32(i)), blockIndex, uint16(i))
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
