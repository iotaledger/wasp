package blocklog

import (
	"errors"
	"fmt"
	"math"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(partition kv.KVStore, blockInfo *BlockInfo) {
	registry := collections.NewArray(partition, PrefixBlockRegistry)
	registry.Push(blockInfo.Bytes())
}

// UpdateLatestBlockInfo is called before producing the next block to save anchor tx id and commitment data of the previous one
func UpdateLatestBlockInfo(partition kv.KVStore, anchorTxID iotago.TransactionID, aliasOutput *isc.AliasOutputWithID, l1commitment *state.L1Commitment) {
	updateUnprocessableRequestsOutputID(partition, anchorTxID)
}

// SaveRequestReceipt appends request record to the record log and creates records for fast lookup
func SaveRequestReceipt(partition kv.KVStore, rec *RequestReceipt, key RequestLookupKey) error {
	// save lookup record for fast lookup
	lookupTable := collections.NewMap(partition, prefixRequestLookupIndex)
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
	for i := range lst {
		if lst[i] == key {
			// already in list. Not normal
			return errors.New("SaveRequestReceipt: inconsistency: duplicate lookup key")
		}
	}
	lst = append(lst, key)
	lookupTable.SetAt(digest[:], lst.Bytes())
	// save the record. Key is a LookupKey
	data := rec.Bytes()
	collections.NewMap(partition, prefixRequestReceipts).SetAt(key.Bytes(), data)
	return nil
}

func SaveEvent(partition kv.KVStore, eventKey []byte, event *isc.Event) {
	collections.NewMap(partition, prefixRequestEvents).SetAt(eventKey, event.Bytes())
}

func mustGetLookupKeyListFromReqID(partition kv.KVStoreReader, reqID isc.RequestID) RequestLookupKeyList {
	lookupTable := collections.NewMapReadOnly(partition, prefixRequestLookupIndex)
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
func getCorrectRecordFromLookupKeyList(partition kv.KVStoreReader, keyList RequestLookupKeyList, reqID isc.RequestID) (*RequestReceipt, error) {
	records := collections.NewMapReadOnly(partition, prefixRequestReceipts)
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

func getRequestReceipt(partition kv.KVStoreReader, reqID isc.RequestID) (*RequestReceipt, error) {
	lst := mustGetLookupKeyListFromReqID(partition, reqID)
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	if err != nil {
		return nil, fmt.Errorf("cannot getCorrectRecordFromLookupKeyList: %w", err)
	}
	return record, nil
}

func getRequestEventsInternal(partition kv.KVStoreReader, reqID isc.RequestID) ([][]byte, error) {
	lst := mustGetLookupKeyListFromReqID(partition, reqID)
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	eventIndex := uint16(0)
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	var ret [][]byte
	for {
		key := NewEventLookupKey(record.BlockIndex, record.RequestIndex, eventIndex).Bytes()
		eventData := events.GetAt(key)
		if eventData == nil {
			return ret, nil
		}
		ret = append(ret, eventData)
		eventIndex++
	}
}

func getSmartContractEventsInternal(partition kv.KVStoreReader, contractID isc.Hname, fromBlock, toBlock uint32) [][]byte {
	registry := collections.NewArrayReadOnly(partition, PrefixBlockRegistry)
	latestBlockIndex := registry.Len() - 1
	adjustedToBlock := toBlock

	if adjustedToBlock > latestBlockIndex {
		adjustedToBlock = latestBlockIndex
	}

	filteredEvents := make([][]byte, 0)
	for blockNumber := fromBlock; blockNumber <= adjustedToBlock; blockNumber++ {
		eventBlockKey := collections.MapElemKey(prefixRequestEvents, codec.EncodeUint32(blockNumber))

		partition.Iterate(eventBlockKey, func(_ kv.Key, value []byte) bool {
			parsedContractID, _ := isc.ContractIDFromEventBytes(value)
			if parsedContractID != contractID {
				return true
			}

			filteredEvents = append(filteredEvents, value)
			return true
		})
	}

	return filteredEvents
}

func pruneEventsByBlockIndex(partition kv.KVStore, blockIndex uint32, totalRequests uint16) {
	events := collections.NewMap(partition, prefixRequestEvents)
	for reqIdx := uint16(0); reqIdx < totalRequests; reqIdx++ {
		eventIndex := uint16(0)
		for {
			key := NewEventLookupKey(blockIndex, reqIdx, eventIndex).Bytes()
			eventData := events.GetAt(key)
			if eventData == nil {
				break
			}
			events.DelAt(key)
			eventIndex++
		}
	}
}

func getRequestLogRecordsForBlockBin(partition kv.KVStoreReader, blockIndex uint32) ([][]byte, bool) {
	blockInfo, ok := GetBlockInfo(partition, blockIndex)
	if !ok {
		return nil, false
	}
	ret := make([][]byte, blockInfo.TotalRequests)
	var found bool
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		ret[reqIdx], found = getRequestRecordDataByRef(partition, blockIndex, reqIdx)
		if !found {
			panic("getRequestLogRecordsForBlockBin: inconsistency: request record not found")
		}
	}
	return ret, true
}

func pruneRequestLookupTable(partition kv.KVStore, lookupDigest isc.RequestLookupDigest, blockIndex uint32) error {
	lut := collections.NewMap(partition, prefixRequestLookupIndex)

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

func pruneRequestLogRecordsByBlockIndex(partition kv.KVStore, blockIndex uint32, totalRequests uint16) {
	receiptMap := collections.NewMap(partition, prefixRequestReceipts)

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

		err = pruneRequestLookupTable(partition, receipt.Request.ID().LookupDigest(), blockIndex)
		if err != nil {
			panic(err)
		}

		receiptMap.DelAt(lookupKey[:])
	}
}

func getBlockInfoBytes(partition kv.KVStoreReader, blockIndex uint32) []byte {
	return collections.NewArrayReadOnly(partition, PrefixBlockRegistry).GetAt(blockIndex)
}

func RequestReceiptKey(rkey RequestLookupKey) []byte {
	return []byte(collections.MapElemKey(prefixRequestReceipts, rkey.Bytes()))
}

func getRequestRecordDataByRef(partition kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(partition, prefixRequestReceipts)
	recBin := lookupTable.GetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
}

func GetOutputID(stateR kv.KVStoreReader, stateIndex uint32, outputIndex uint16) (iotago.OutputID, bool) {
	blockInfo, ok := GetBlockInfo(stateR, stateIndex+1)
	if !ok {
		return iotago.OutputID{}, false
	}
	return iotago.OutputIDFromTransactionIDAndIndex(blockInfo.PreviousAliasOutput.TransactionID(), outputIndex), true
}

// tries to get block index from ParamBlockIndex, if no parameter is provided, returns the latest block index
func getBlockIndexParams(ctx isc.SandboxView) uint32 {
	ret := ctx.Params().MustGetUint32(ParamBlockIndex, math.MaxUint32)
	if ret != math.MaxUint32 {
		return ret
	}
	registry := collections.NewArrayReadOnly(ctx.StateR(), PrefixBlockRegistry)
	return registry.Len() - 1
}

func pruneBlock(partition kv.KVStore, blockIndex uint32) {
	blockInfo, ok := GetBlockInfo(partition, blockIndex)
	if !ok {
		// already pruned?
		return
	}
	registry := collections.NewArray(partition, PrefixBlockRegistry)
	registry.PruneAt(blockIndex)
	pruneRequestLogRecordsByBlockIndex(partition, blockIndex, blockInfo.TotalRequests)
	pruneEventsByBlockIndex(partition, blockIndex, blockInfo.TotalRequests)
}

func eventsToDict(events [][]byte) dict.Dict {
	ret := dict.New()
	retEvents := collections.NewArray(ret, ParamEvent)
	for _, event := range events {
		retEvents.Push(event)
	}
	return ret
}
