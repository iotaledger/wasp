package blocklog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(partition kv.KVStore, blockInfo *BlockInfo) {
	registry := collections.NewArray32(partition, PrefixBlockRegistry)
	registry.Push(blockInfo.Bytes())
}

// UpdateLatestBlockInfo is called before producing the next block to save anchor tx id and commitment data of the previous one
func UpdateLatestBlockInfo(partition kv.KVStore, anchorTxID iotago.TransactionID, aliasOutput *isc.AliasOutputWithID, l1commitment *state.L1Commitment) {
	registry := collections.NewArray32(partition, PrefixBlockRegistry)
	lastBlockIndex := registry.Len() - 1
	blockInfoBuffer := registry.GetAt(lastBlockIndex)
	blockInfo, err := BlockInfoFromBytes(blockInfoBuffer)
	if err != nil {
		panic(err)
	}
	registry.SetAt(lastBlockIndex, blockInfo.Bytes())
}

// SaveControlAddressesIfNecessary saves new information about state address in the blocklog partition
// If state address does not change, it does nothing
func SaveControlAddressesIfNecessary(partition kv.KVStore, stateAddress, governingAddress iotago.Address, blockIndex uint32) {
	registry := collections.NewArray32(partition, prefixControlAddresses)
	l := registry.Len()
	if l != 0 {
		addrs, err := ControlAddressesFromBytes(registry.GetAt(l - 1))
		if err != nil {
			panic(fmt.Sprintf("SaveControlAddressesIfNecessary: %v", err))
		}
		if addrs.StateAddress.Equal(stateAddress) && addrs.GoverningAddress.Equal(governingAddress) {
			return
		}
	}
	rec := &ControlAddresses{
		StateAddress:     stateAddress,
		GoverningAddress: governingAddress,
		SinceBlockIndex:  blockIndex,
	}
	registry.Push(rec.Bytes())
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

func SaveEvent(partition kv.KVStore, msg string, key EventLookupKey, contract isc.Hname) {
	text := fmt.Sprintf("%s: %s", contract.String(), msg)
	collections.NewMap(partition, prefixRequestEvents).SetAt(key.Bytes(), []byte(text))
	scLut := collections.NewMap(partition, prefixSmartContractEventsLookup)
	entries := scLut.GetAt(contract.Bytes())
	entries = append(entries, key.Bytes()...)
	scLut.SetAt(contract.Bytes(), entries)
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
		rec, err := RequestReceiptFromBytes(recBytes)
		if err != nil {
			return nil, fmt.Errorf("RequestReceiptFromBytes returned: %w", err)
		}
		if rec.Request.ID().Equals(reqID) {
			rec.BlockIndex = lookupKey.BlockIndex()
			rec.RequestIndex = lookupKey.RequestIndex()
			return rec, nil
		}
	}
	return nil, nil
}

// isRequestProcessedInternal does quick lookup to check if it wasn't seen yet
func isRequestProcessedInternal(partition kv.KVStoreReader, reqID isc.RequestID) (*RequestReceipt, error) {
	lst := mustGetLookupKeyListFromReqID(partition, reqID)
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	if err != nil {
		return nil, fmt.Errorf("cannot getCorrectRecordFromLookupKeyList: %w", err)
	}
	return record, nil
}

func getRequestEventsInternal(partition kv.KVStoreReader, reqID isc.RequestID) ([]string, error) {
	lst := mustGetLookupKeyListFromReqID(partition, reqID)
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	ret := []string{}
	eventIndex := uint16(0)
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	for {
		key := NewEventLookupKey(record.BlockIndex, record.RequestIndex, eventIndex)
		msg := events.GetAt(key.Bytes())
		if msg == nil {
			return ret, nil
		}
		ret = append(ret, string(msg))
		eventIndex++
	}
}

func getSmartContractEventsInternal(partition kv.KVStoreReader, contract isc.Hname, fromBlock, toBlock uint32) ([]string, error) {
	scLut := collections.NewMapReadOnly(partition, prefixSmartContractEventsLookup)
	ret := []string{}
	entries := scLut.GetAt(contract.Bytes())
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	keysBuf := bytes.NewBuffer(entries)
	for {
		key, err := EventLookupKeyFromBytes(keysBuf)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("getSmartContractEventsIntern unable to parse key. %w", err)
		}
		if key == nil { // no more events
			return ret, nil
		}
		keyBlockIndex := key.BlockIndex()
		if keyBlockIndex < fromBlock {
			continue
		}
		if keyBlockIndex > toBlock {
			return ret, nil
		}
		event := events.GetAt(key.Bytes())
		ret = append(ret, string(event))
	}
}

func getRequestLogRecordsForBlockBin(partition kv.KVStoreReader, blockIndex uint32) ([][]byte, bool, error) {
	blockInfo, err := GetBlockInfo(partition, blockIndex)
	if err != nil || blockInfo == nil {
		return nil, false, err
	}
	ret := make([][]byte, blockInfo.TotalRequests)
	found := false
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		ret[reqIdx], found = getRequestRecordDataByRef(partition, blockIndex, reqIdx)
		if !found {
			panic("getRequestLogRecordsForBlockBin: inconsistency: request record not found")
		}
	}
	return ret, true, nil
}

func getBlockInfoBytes(partition kv.KVStoreReader, blockIndex uint32) []byte {
	return collections.NewArray32ReadOnly(partition, PrefixBlockRegistry).GetAt(blockIndex)
}

func RequestReceiptKey(rkey RequestLookupKey) []byte {
	return collections.MapElemKey(prefixRequestReceipts, rkey.Bytes())
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

func GetOutputID(stateR kv.KVStoreReader, stateIndex uint32, outputIndex uint16) (iotago.OutputID, error) {
	blockInfo, err := GetBlockInfo(stateR, stateIndex+1)
	if err != nil {
		return iotago.OutputID{}, err
	}
	return iotago.OutputIDFromTransactionIDAndIndex(blockInfo.PreviousAliasOutput.TransactionID(), outputIndex), nil
}

// tries to get block index from ParamBlockIndex, if no parameter is provided, returns the latest block index
func getBlockIndexParams(ctx isc.SandboxView) uint32 {
	ret := ctx.Params().MustGetUint32(ParamBlockIndex, math.MaxUint32)
	if ret != math.MaxUint32 {
		return ret
	}
	registry := collections.NewArray32ReadOnly(ctx.StateR(), PrefixBlockRegistry)
	return registry.Len() - 1
}
