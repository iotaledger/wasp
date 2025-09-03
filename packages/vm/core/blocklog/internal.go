package blocklog

import (
	"errors"
	"fmt"
	"slices"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
)

func (s *StateWriter) GetBlockRegistry() *collections.Array {
	return collections.NewArray(s.state, prefixBlockRegistry)
}

func (s *StateReader) GetBlockRegistry() *collections.ArrayReadOnly {
	return collections.NewArrayReadOnly(s.state, prefixBlockRegistry)
}

func (s *StateReader) IterateBlockRegistryPrefix(f func(blockInfo *BlockInfo)) {
	s.state.Iterate(collections.ArrayElemPrefix(prefixBlockRegistry), func(key kv.Key, value []byte) bool {
		f(lo.Must(BlockInfoFromBytes(value)))
		return true
	})
}

// SaveNextBlockInfo appends block info and returns its index
func (s *StateWriter) SaveNextBlockInfo(blockInfo *BlockInfo) {
	registry := collections.NewArray(s.state, prefixBlockRegistry)
	registry.Push(blockInfo.Bytes())
}

// SaveRequestReceipt appends request record to the record log and creates records for fast lookup
func (s *StateWriter) SaveRequestReceipt(rec *RequestReceipt, key RequestLookupKey) error {
	// save lookup record for fast lookup
	lookupTable := collections.NewMap(s.state, prefixRequestLookupIndex)
	digest := rec.Request.ID().LookupDigest()
	var lst RequestLookupKeyList
	digestExists := lookupTable.HasAt(digest[:])
	if !digestExists {
		// new digest, most common
		lst = make(RequestLookupKeyList, 0, 1)
	} else {
		// existing digest (should happen not often)
		bin := lookupTable.GetAt(digest[:])
		var err2 error
		if lst, err2 = RequestLookupKeyListFromBytes(bin); err2 != nil {
			return fmt.Errorf("SaveRequestReceipt: %w", err2)
		}
	}
	if slices.Contains(lst, key) {
		// already in list. Not normal
		return errors.New("SaveRequestReceipt: inconsistency: duplicate lookup key")
	}
	lst = append(lst, key)
	lookupTable.SetAt(digest[:], lst.Bytes())
	// save the record. Key is a LookupKey
	data := rec.Bytes()
	collections.NewMap(s.state, prefixRequestReceipts).SetAt(key.Bytes(), data)
	return nil
}

func (s *StateReader) mustGetLookupKeyListFromReqID(reqID isc.RequestID) RequestLookupKeyList {
	lookupTable := collections.NewMapReadOnly(s.state, prefixRequestLookupIndex)
	digest := reqID.LookupDigest()
	seen := lookupTable.HasAt(digest[:])
	if !seen {
		return nil
	}
	// the lookup record is here, have to check is it is not a collision of digests
	bin := lookupTable.GetAt(digest[:])
	lst, err := RequestLookupKeyListFromBytes(bin)
	if err != nil {
		panic("mustGetLookupKeyListFromReqID: data conversion error")
	}
	return lst
}

// RequestLookupKeyList contains multiple references for record entries with colliding digests, this function returns the correct record for the given requestID
func (s *StateReader) getCorrectRecordFromLookupKeyList(keyList RequestLookupKeyList, reqID isc.RequestID) (*RequestReceipt, error) {
	records := collections.NewMapReadOnly(s.state, prefixRequestReceipts)
	for _, lookupKey := range keyList {
		recBytes := records.GetAt(lookupKey.Bytes())
		rec, err := RequestReceiptFromBytes(recBytes, lookupKey.BlockIndex(), lookupKey.RequestIndex())
		if err != nil {
			return nil, fmt.Errorf("RequestReceiptFromBytes returned: %w", err)
		}
		if rec.Request.ID().Equals(reqID) {
			return rec, nil
		}
	}
	return nil, nil
}

// GetRequestReceipt returns the receipt for the given request, or nil if not found
func (s *StateReader) GetRequestReceipt(reqID isc.RequestID) (*RequestReceipt, error) {
	lst := s.mustGetLookupKeyListFromReqID(reqID)
	record, err := s.getCorrectRecordFromLookupKeyList(lst, reqID)
	if err != nil {
		return nil, fmt.Errorf("cannot getCorrectRecordFromLookupKeyList: %w", err)
	}
	return record, nil
}

func (s *StateReader) getRequestLogRecordsForBlockBin(blockIndex uint32) ([][]byte, bool) {
	blockInfo, ok := s.GetBlockInfo(blockIndex)
	if !ok {
		return nil, false
	}
	ret := make([][]byte, blockInfo.TotalRequests)
	var found bool
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		ret[reqIdx], found = s.getRequestRecordDataByRef(blockIndex, reqIdx)
		if !found {
			panic("getRequestLogRecordsForBlockBin: inconsistency: request record not found")
		}
	}
	return ret, true
}

func (s *StateWriter) pruneRequestLookupTable(lookupDigest isc.RequestLookupDigest, blockIndex uint32) error {
	lut := collections.NewMap(s.state, prefixRequestLookupIndex)

	res := lut.GetAt(lookupDigest[:])
	if len(res) == 0 {
		return nil
	}

	requests, err := RequestLookupKeyListFromBytes(res)
	if err != nil {
		return err
	}

	filteredRequestKeys := make(RequestLookupKeyList, 0)

	for _, requestKey := range requests {
		if requestKey.BlockIndex() != blockIndex {
			filteredRequestKeys = append(filteredRequestKeys, requestKey)
		}
	}

	lut.SetAt(lookupDigest[:], filteredRequestKeys.Bytes())
	return nil
}

func (s *StateWriter) pruneRequestLogRecordsByBlockIndex(blockIndex uint32, totalRequests uint16) {
	receiptMap := collections.NewMap(s.state, prefixRequestReceipts)

	for reqIdx := uint16(0); reqIdx < totalRequests; reqIdx++ {
		lookupKey := NewRequestLookupKey(blockIndex, reqIdx)

		receiptBytes := receiptMap.GetAt(lookupKey[:])
		if len(receiptBytes) == 0 {
			continue
		}

		receipt, err := RequestReceiptFromBytes(receiptBytes, blockIndex, reqIdx)
		if err != nil {
			panic(err)
		}

		err = s.pruneRequestLookupTable(receipt.Request.ID().LookupDigest(), blockIndex)
		if err != nil {
			panic(err)
		}

		receiptMap.DelAt(lookupKey[:])
	}
}

func (s *StateReader) getBlockInfoBytes(blockIndex uint32) []byte {
	return collections.NewArrayReadOnly(s.state, prefixBlockRegistry).GetAt(blockIndex)
}

func RequestReceiptKey(rkey RequestLookupKey) []byte {
	return []byte(collections.MapElemKey(prefixRequestReceipts, rkey.Bytes()))
}

func (s *StateReader) getRequestRecordDataByRef(blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(s.state, prefixRequestReceipts)
	recBin := lookupTable.GetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
}

// func (s *StateReader) GetOutputID(stateIndex uint32, outputIndex uint16) (iotago.OutputID, bool) {
// 	blockInfo, ok := s.GetBlockInfo(stateIndex + 1)
// 	if !ok {
// 		return iotago.OutputID{}, false
// 	}
// 	return iotago.OutputIDFromTransactionIDAndIndex(blockInfo.PreviousAliasOutput.TransactionID(), outputIndex), true
// }

// tries to get block index from ParamBlockIndex, if no parameter is provided, returns the latest block index
func getBlockIndexParams(ctx isc.SandboxView, blockIndexOptional *uint32) uint32 {
	if blockIndexOptional != nil {
		return *blockIndexOptional
	}
	registry := collections.NewArrayReadOnly(ctx.StateR(), prefixBlockRegistry)
	return registry.Len() - 1
}

func (s *StateWriter) pruneBlock(blockIndex uint32) {
	blockInfo, ok := s.GetBlockInfo(blockIndex)
	if !ok {
		// already pruned?
		return
	}
	registry := collections.NewArray(s.state, prefixBlockRegistry)
	registry.PruneAt(blockIndex)
	s.pruneRequestLogRecordsByBlockIndex(blockIndex, blockInfo.TotalRequests)
}
