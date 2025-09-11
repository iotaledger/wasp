package blocklog

import (
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
)

func (s *StateReader) GetRequestReceiptsInBlock(blockIndex uint32) (*BlockInfo, []*RequestReceipt, error) {
	blockInfo, ok := s.GetBlockInfo(blockIndex)
	if !ok {
		return nil, nil, fmt.Errorf("block not found: %d", blockIndex)
	}
	recs := make([]*RequestReceipt, blockInfo.TotalRequests)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		recBin, ok := s.getRequestRecordDataByRef(blockIndex, reqIdx)
		if !ok {
			return nil, nil, fmt.Errorf("request not found: %d/%d", blockIndex, reqIdx)
		}
		rec, err := RequestReceiptFromBytes(recBin, blockIndex, reqIdx)
		if err != nil {
			return nil, nil, err
		}
		recs[reqIdx] = rec
	}
	return blockInfo, recs, nil
}

// IsRequestProcessed check if requestID is stored in the chain state as processed
func (s *StateReader) IsRequestProcessed(requestID isc.RequestID) (bool, error) {
	requestReceipt, err := s.GetRequestReceipt(requestID)
	if err != nil {
		return false, fmt.Errorf("cannot get request receipt: %w", err)
	}
	return requestReceipt != nil, nil
}

// GetRequestRecordDataByRequestID tries to obtain the receipt data for a given request
// returns nil if receipt was not found
func (s *StateReader) GetRequestRecordDataByRequestID(reqID isc.RequestID) (*RequestReceipt, error) {
	lookupDigest := reqID.LookupDigest()
	lookupTable := collections.NewMapReadOnly(s.state, prefixRequestLookupIndex)
	lookupKeyListBin := lookupTable.GetAt(lookupDigest[:])
	if lookupKeyListBin == nil {
		return nil, nil
	}
	lookupKeyList, err := RequestLookupKeyListFromBytes(lookupKeyListBin)
	if err != nil {
		return nil, err
	}
	for i := range lookupKeyList {
		recBin, found := s.getRequestRecordDataByRef(lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex())
		if !found {
			return nil, errors.New("inconsistency: request log record wasn't found by exact reference")
		}
		rec, err := RequestReceiptFromBytes(recBin, lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex())
		if err != nil {
			return nil, err
		}
		if rec.Request.ID().Equals(reqID) {
			return rec, nil
		}
	}
	return nil, nil
}

func (s *StateReader) GetBlockInfo(blockIndex uint32) (*BlockInfo, bool) {
	data := s.getBlockInfoBytes(blockIndex)
	if data == nil {
		return nil, false
	}
	ret, err := BlockInfoFromBytes(data)
	if err != nil {
		panic(err)
	}
	return ret, true
}

func (s *StateWriter) Prune(latestBlockIndex uint32, blockKeepAmount int32) {
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
	s.pruneBlock(toDelete)
}
