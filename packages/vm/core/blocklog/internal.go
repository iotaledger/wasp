package blocklog

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"

	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(partition kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(partition, prefixBlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	ret := registry.MustLen() - 1
	return ret
}

// UpdateLatestBlockInfo is called before producing the next block to save anchor tx id and commitment data of the previous one
func UpdateLatestBlockInfo(partition kv.KVStore, anchorTxID iotago.TransactionID, l1Commitment *state.L1Commitment) {
	registry := collections.NewArray32(partition, prefixBlockRegistry)
	lastBlockIndex := registry.MustLen() - 1
	blockInfoBuffer := registry.MustGetAt(lastBlockIndex)
	blockInfo, _ := BlockInfoFromBytes(lastBlockIndex, blockInfoBuffer)

	blockInfo.AnchorTransactionID = anchorTxID
	blockInfo.L1Commitment = l1Commitment

	registry.MustSetAt(lastBlockIndex, blockInfo.Bytes())
}

func GetAnchorTransactionIDByBlockIndex(partition kv.KVStore, blockIndex uint32) iotago.TransactionID {
	registry := collections.NewArray32(partition, prefixBlockRegistry)
	blockInfoBuffer := registry.MustGetAt(blockIndex)
	blockInfo, err := BlockInfoFromBytes(blockIndex, blockInfoBuffer)
	if err != nil {
		panic("Failed to parse blockinfo")
	}

	return blockInfo.AnchorTransactionID
}

// SaveControlAddressesIfNecessary saves new information about state address in the blocklog partition
// If state address does not change, it does nothing
func SaveControlAddressesIfNecessary(partition kv.KVStore, stateAddress, governingAddress iotago.Address, blockIndex uint32) {
	registry := collections.NewArray32(partition, prefixControlAddresses)
	l := registry.MustLen()
	if l != 0 {
		addrs, err := ControlAddressesFromBytes(registry.MustGetAt(l - 1))
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
	registry.MustPush(rec.Bytes())
}

// SaveRequestReceipt appends request record to the record log and creates records for fast lookup
func SaveRequestReceipt(partition kv.KVStore, rec *RequestReceipt, key RequestLookupKey) error {
	// save lookup record for fast lookup
	lookupTable := collections.NewMap(partition, prefixRequestLookupIndex)
	digest := rec.Request.ID().LookupDigest()
	var lst RequestLookupKeyList
	digestExists, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	if !digestExists {
		// new digest, most common
		lst = make(RequestLookupKeyList, 0, 1)
	} else {
		// existing digest (should happen not often)
		bin, err := lookupTable.GetAt(digest[:])
		if err != nil {
			return xerrors.Errorf("SaveRequestReceipt: %w", err)
		}
		if lst, err = RequestLookupKeyListFromBytes(bin); err != nil {
			return xerrors.Errorf("SaveRequestReceipt: %w", err)
		}
	}
	for i := range lst {
		if lst[i] == key {
			// already in list. Not normal
			return xerrors.New("SaveRequestReceipt: inconsistency: duplicate lookup key")
		}
	}
	lst = append(lst, key)
	if err := lookupTable.SetAt(digest[:], lst.Bytes()); err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	// save the record. Key is a LookupKey
	data := rec.Bytes()
	if err = collections.NewMap(partition, prefixRequestReceipts).SetAt(key.Bytes(), data); err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	return nil
}

func SaveEvent(partition kv.KVStore, msg string, key EventLookupKey, contract isc.Hname) error {
	text := fmt.Sprintf("%s: %s", contract.String(), msg)
	if err := collections.NewMap(partition, prefixRequestEvents).SetAt(key.Bytes(), []byte(text)); err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	scLut := collections.NewMap(partition, prefixSmartContractEventsLookup)
	entries, err := scLut.GetAt(contract.Bytes())
	if err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	entries = append(entries, key.Bytes()...)
	err = scLut.SetAt(contract.Bytes(), entries)
	if err != nil {
		return xerrors.Errorf("SaveRequestReceipt: %w", err)
	}
	return nil
}

func mustGetLookupKeyListFromReqID(partition kv.KVStoreReader, reqID *isc.RequestID) (RequestLookupKeyList, error) {
	lookupTable := collections.NewMapReadOnly(partition, prefixRequestLookupIndex)
	digest := reqID.LookupDigest()
	seen, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return nil, err
	}
	if !seen {
		return nil, nil
	}
	// the lookup record is here, have to check is it is not a collision of digests
	bin := lookupTable.MustGetAt(digest[:])
	lst, err := RequestLookupKeyListFromBytes(bin)
	if err != nil {
		panic("mustGetLookupKeyListFromReqID: data conversion error")
	}
	return lst, nil
}

// RequestLookupKeyList contains multiple references for record entries with colliding digests, this function returns the correct record for the given requestID
func getCorrectRecordFromLookupKeyList(partition kv.KVStoreReader, keyList RequestLookupKeyList, reqID *isc.RequestID) (*RequestReceipt, error) {
	records := collections.NewMapReadOnly(partition, prefixRequestReceipts)
	for _, lookupKey := range keyList {
		recBytes, err := records.GetAt(lookupKey.Bytes())
		if err != nil {
			return nil, xerrors.Errorf("records.GetAt(lookupKey.Bytes()) returned: %w", err)
		}
		rec, err := RequestReceiptFromBytes(recBytes)
		if err != nil {
			return nil, xerrors.Errorf("RequestReceiptFromBytes returned: %w", err)
		}
		if rec.Request.ID().Equals(*reqID) {
			rec.BlockIndex = lookupKey.BlockIndex()
			rec.RequestIndex = lookupKey.RequestIndex()
			return rec, nil
		}
	}
	return nil, nil
}

// isRequestProcessedInternal does quick lookup to check if it wasn't seen yet
func isRequestProcessedInternal(partition kv.KVStoreReader, reqID *isc.RequestID) (*RequestReceipt, error) {
	lst, err := mustGetLookupKeyListFromReqID(partition, reqID)
	if err != nil {
		return nil, xerrors.Errorf("cannot mustGetLookupKeyListFromReqID: %w", err)
	}
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	if err != nil {
		return nil, xerrors.Errorf("cannot getCorrectRecordFromLookupKeyList: %w", err)
	}
	return record, nil
}

func getRequestEventsInternal(partition kv.KVStoreReader, reqID *isc.RequestID) ([]string, error) {
	lst, err := mustGetLookupKeyListFromReqID(partition, reqID)
	if err != nil {
		return nil, err
	}
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
		msg, err := events.GetAt(key.Bytes())
		if err != nil {
			return nil, err
		}
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
	entries, err := scLut.GetAt(contract.Bytes())
	if err != nil {
		return nil, err
	}
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	keysBuf := bytes.NewBuffer(entries)
	for {
		key, err := EventLookupKeyFromBytes(keysBuf)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, xerrors.Errorf("getSmartContractEventsIntern unable to parse key. %v", err)
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
		event, err := events.GetAt(key.Bytes())
		if err != nil {
			return nil, xerrors.Errorf("getSmartContractEventsIntern unable to get event by key. %v", err)
		}
		ret = append(ret, string(event))
	}
}

func GetBlockEventsInternal(partition kv.KVStoreReader, blockIndex uint32) ([]string, error) {
	blockInfo, err := getRequestLogRecordsForBlock(partition, blockIndex)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	events := collections.NewMapReadOnly(partition, prefixRequestEvents)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
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

func getRequestLogRecordsForBlock(partition kv.KVStoreReader, blockIndex uint32) (*BlockInfo, error) {
	if blockIndex == 0 {
		return nil, nil
	}
	blockInfoBin, found, err := getBlockInfoDataInternal(partition, blockIndex)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	blockInfo, err := BlockInfoFromBytes(blockIndex, blockInfoBin)
	if err != nil {
		return nil, err
	}
	return blockInfo, nil
}

func getRequestLogRecordsForBlockBin(partition kv.KVStoreReader, blockIndex uint32) ([][]byte, bool, error) {
	blockInfo, err := getRequestLogRecordsForBlock(partition, blockIndex)
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

func getBlockInfoDataInternal(partition kv.KVStoreReader, blockIndex uint32) ([]byte, bool, error) {
	data, err := collections.NewArray32ReadOnly(partition, prefixBlockRegistry).GetAt(blockIndex)
	return data, err == nil, err
}

func getBlockInfo(partition kv.KVStoreReader, blockIndex uint32) (*BlockInfo, error) {
	data, ok, err := getBlockInfoDataInternal(partition, blockIndex)
	if err != nil {
		return nil, xerrors.Errorf("getBlockInfo: %w", err)
	}
	if !ok {
		return nil, xerrors.Errorf("getBlockInfo: can't find block record #%d", blockIndex)
	}
	ret, err := BlockInfoFromBytes(blockIndex, data)
	if err != nil {
		return nil, xerrors.Errorf("getBlockInfo: %w", err)
	}
	return ret, nil
}

func mustGetBlockInfo(partition kv.KVStoreReader, blockIndex uint32) *BlockInfo {
	blockInfo, err := getBlockInfo(partition, blockIndex)
	if err != nil {
		panic(xerrors.Errorf("mustGetBlockInfo: %w", err))
	}
	return blockInfo
}

func RequestReceiptKey(rkey RequestLookupKey) []byte {
	return collections.MapElemKey(prefixRequestReceipts, rkey.Bytes())
}

func getRequestRecordDataByRef(partition kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(partition, prefixRequestReceipts)
	recBin := lookupTable.MustGetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
}

func GetOutputID(stateR kv.KVStoreReader, stateIndex uint32, outputIndex uint16) (iotago.OutputID, error) {
	blockInfo, err := getBlockInfo(stateR, stateIndex)
	if err != nil {
		return iotago.OutputID{}, err
	}

	return iotago.OutputIDFromTransactionIDAndIndex(blockInfo.AnchorTransactionID, outputIndex), nil
}

// tries to get block index from ParamBlockIndex, if no parameter is provided, returns the latest block index
func getBlockIndexParams(ctx isc.SandboxView) uint32 {
	ret := ctx.Params().MustGetUint32(ParamBlockIndex, math.MaxUint32)
	if ret != math.MaxUint32 {
		return ret
	}
	registry := collections.NewArray32ReadOnly(ctx.StateR(), prefixBlockRegistry)
	return registry.MustLen() - 1
}
