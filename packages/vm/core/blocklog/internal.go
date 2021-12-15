package blocklog

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/iotaledger/wasp/packages/util"

	iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(partition kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(partition, KeyPrefixBlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	ret := registry.MustLen() - 1
	return ret
}

// SetAnchorTransactionIDOfLatestBlock is called before producing the next block to save anchor tx id of the previous one
func SetAnchorTransactionIDOfLatestBlock(partition kv.KVStore, transactionId iotago.TransactionID) {
	registry := collections.NewArray32(partition, KeyPrefixBlockRegistry)
	lastBlockIndex := registry.MustLen() - 1
	blockInfoBuffer := registry.MustGetAt(lastBlockIndex)
	blockInfo, _ := BlockInfoFromBytes(lastBlockIndex, blockInfoBuffer)

	blockInfo.AnchorTransactionID = transactionId

	registry.MustSetAt(lastBlockIndex, blockInfo.Bytes())
}

func GetAnchorTransactionIDByBlockIndex(partition kv.KVStore, blockIndex uint32) iotago.TransactionID {
	registry := collections.NewArray32(partition, KeyPrefixBlockRegistry)
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
	registry := collections.NewArray32(partition, KeyPrefixControlAddresses)
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

// SaveRequestLogRecord appends request record to the record log and creates records for fast lookup
func SaveRequestLogRecord(partition kv.KVStore, rec *RequestReceipt, key RequestLookupKey) error {
	// save lookup record for fast lookup
	lookupTable := collections.NewMap(partition, KeyPrefixRequestLookupIndex)
	digest := rec.RequestData.ID().LookupDigest()
	var lst RequestLookupKeyList
	digestExists, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	if !digestExists {
		// new digest, most common
		lst = make(RequestLookupKeyList, 0, 1)
	} else {
		// existing digest (should happen not often)
		bin, err := lookupTable.GetAt(digest[:])
		if err != nil {
			return xerrors.Errorf("SaveRequestLogRecord: %w", err)
		}
		if lst, err = RequestLookupKeyListFromBytes(bin); err != nil {
			return xerrors.Errorf("SaveRequestLogRecord: %w", err)
		}
	}
	for i := range lst {
		if lst[i] == key {
			// already in list. Not normal
			return xerrors.New("SaveRequestLogRecord: inconsistency: duplicate lookup key")
		}
	}
	lst = append(lst, key)
	if err := lookupTable.SetAt(digest[:], lst.Bytes()); err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	// save the record. Key is a LookupKey
	data := rec.Bytes()
	if err = collections.NewMap(partition, KeyPrefixRequestReceipts).SetAt(key.Bytes(), data); err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	return nil
}

func SaveEvent(partition kv.KVStore, msg string, key EventLookupKey, contract iscp.Hname) error {
	text := fmt.Sprintf("%s: %s", contract.String(), msg)
	if err := collections.NewMap(partition, KeyPrefixRequestEvents).SetAt(key.Bytes(), []byte(text)); err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	scLut := collections.NewMap(partition, KeyPrefixSmartContractEventsLookup)
	entries, err := scLut.GetAt(contract.Bytes())
	if err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	entries = append(entries, key.Bytes()...)
	err = scLut.SetAt(contract.Bytes(), entries)
	if err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	return nil
}

func mustGetLookupKeyListFromReqID(partition kv.KVStoreReader, reqID *iscp.RequestID) (RequestLookupKeyList, error) {
	lookupTable := collections.NewMapReadOnly(partition, KeyPrefixRequestLookupIndex)
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
func getCorrectRecordFromLookupKeyList(partition kv.KVStoreReader, keyList RequestLookupKeyList, reqID *iscp.RequestID) (*RequestReceipt, error) {
	records := collections.NewMapReadOnly(partition, KeyPrefixRequestReceipts)
	for _, lookupKey := range keyList {
		recBytes, err := records.GetAt(lookupKey.Bytes())
		if err != nil {
			return nil, err
		}
		rec, err := RequestReceiptFromBytes(recBytes)
		if err != nil {
			return nil, err
		}
		if rec.RequestData.ID() == *reqID {
			rec.BlockIndex = lookupKey.BlockIndex()
			rec.RequestIndex = lookupKey.RequestIndex()
			return rec, nil
		}
	}
	return nil, nil
}

// isRequestProcessedInternal does quick lookup to check if it wasn't seen yet
func isRequestProcessedInternal(partition kv.KVStoreReader, reqID *iscp.RequestID) (bool, error) {
	lst, err := mustGetLookupKeyListFromReqID(partition, reqID)
	if err != nil {
		return false, err
	}
	record, err := getCorrectRecordFromLookupKeyList(partition, lst, reqID)
	return record != nil, err
}

func getRequestEventsInternal(partition kv.KVStoreReader, reqID *iscp.RequestID) ([]string, error) {
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
	events := collections.NewMapReadOnly(partition, KeyPrefixRequestEvents)
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

func getSmartContractEventsInternal(partition kv.KVStoreReader, contract iscp.Hname, fromBlock, toBlock uint32) ([]string, error) {
	scLut := collections.NewMapReadOnly(partition, KeyPrefixSmartContractEventsLookup)
	ret := []string{}
	entries, err := scLut.GetAt(contract.Bytes())
	if err != nil {
		return nil, err
	}
	events := collections.NewMapReadOnly(partition, KeyPrefixRequestEvents)
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
	events := collections.NewMapReadOnly(partition, KeyPrefixRequestEvents)
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
	data, err := collections.NewArray32ReadOnly(partition, KeyPrefixBlockRegistry).GetAt(blockIndex)
	return data, err == nil, err
}

func mustGetBlockInfo(partition kv.KVStoreReader, blockIndex uint32) *BlockInfo {
	data, ok, err := getBlockInfoDataInternal(partition, blockIndex)
	if err != nil {
		panic(xerrors.Errorf("mustGetBlockInfo: %w", err))
	}
	if !ok {
		panic(xerrors.Errorf("mustGetBlockInfo: can't find block recird #%d", blockIndex))
	}
	ret, err := BlockInfoFromBytes(blockIndex, data)
	if err != nil {
		panic(xerrors.Errorf("mustGetBlockInfo: %w", err))
	}
	return ret
}

func getRequestRecordDataByRef(partition kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(partition, KeyPrefixRequestReceipts)
	recBin := lookupTable.MustGetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
}

func getRequestRecordDataByRequestID(ctx iscp.SandboxView, reqID iscp.RequestID) ([]byte, uint32, uint16, bool) {
	lookupDigest := reqID.LookupDigest()
	lookupTable := collections.NewMapReadOnly(ctx.State(), KeyPrefixRequestLookupIndex)
	lookupKeyListBin := lookupTable.MustGetAt(lookupDigest[:])
	if lookupKeyListBin == nil {
		return nil, 0, 0, false
	}
	a := assert.NewAssert(ctx.Log())
	lookupKeyList, err := RequestLookupKeyListFromBytes(lookupKeyListBin)
	a.RequireNoError(err)
	for i := range lookupKeyList {
		recBin, found := getRequestRecordDataByRef(ctx.State(), lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex())
		a.Require(found, "inconsistency: request log record wasn't found by exact reference")
		rec, err := RequestReceiptFromBytes(recBin)
		a.RequireNoError(err)
		if rec.RequestData.ID() == reqID {
			return recBin, lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex(), true
		}
	}
	return nil, 0, 0, false
}

func getNativeTokensUTXOLookupMap(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, KeyPrefixNativeTokenOutputLookupMap)
}

func getNativeTokensUTXOLookupMapR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, KeyPrefixNativeTokenOutputLookupMap)
}

// handling internal UTXOs of NativeAssets

func encodeNativeTokenLookupInfo(stateIndex uint32, outputIndex uint16) []byte {
	ret := make([]byte, 6)
	copy(ret[:4], util.Uint32To4Bytes(stateIndex))
	copy(ret[4:], util.Uint16To2Bytes(outputIndex))
	return ret
}

func mustDecodeNativeTokenLookupInfo(data []byte) (uint32, uint16) {
	if len(data) != 6 {
		panic("internal inconsistency")
	}
	return util.MustUint32From4Bytes(data[:4]), util.MustUint16From2Bytes(data[4:])
}

type NativeTokenUTXOUpdateCmd struct {
	Add         bool // otherwise remove
	ID          iotago.NativeTokenID
	OutputIndex uint16
}

func UpdateNativeAssetsUTXOIndices(state kv.KVStore, stateIndex uint32, updates []*NativeTokenUTXOUpdateCmd) {
	mapping := getNativeTokensUTXOLookupMap(state)
	for _, upd := range updates {
		if upd.Add {
			entry := encodeNativeTokenLookupInfo(stateIndex, upd.OutputIndex)
			mapping.MustSetAt(upd.ID[:], entry)
		} else {
			mapping.MustDelAt(upd.ID[:])
		}
	}
}

func GetUTXOIDForAsset(state kv.KVStore, id *iotago.NativeTokenID) *iotago.UTXOInput {
	mapping := getNativeTokensUTXOLookupMapR(state)
	data := mapping.MustGetAt(id[:])
	if data == nil {
		return nil
	}
	stateIndex, outputIndex := mustDecodeNativeTokenLookupInfo(data)
	return &iotago.UTXOInput{
		TransactionID:          mustGetBlockInfo(state, stateIndex).AnchorTransactionID,
		TransactionOutputIndex: outputIndex,
	}
}
