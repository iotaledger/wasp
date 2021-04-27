package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(state kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(state, StateVarBlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	return registry.MustLen() - 1
}

// SaveRequestLogRecord appends request record to the record log and creates records for fast lookup
func SaveRequestLogRecord(state kv.KVStore, rec *RequestLogRecord, key RequestLookupKey) error {
	// save lookup record
	lookupTable := collections.NewMap(state, StateVarRequestLookupIndex)
	digest := rec.RequestID.LookupDigest()
	var lst RequestLookupKeyList
	digestExist, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return xerrors.Errorf("SaveRequestLookup: %w", err)
	}
	if !digestExist {
		// new digest, most common
		lst = make(RequestLookupKeyList, 0, 1)
	} else {
		// existing digest (should happen not often)
		bin, err := lookupTable.GetAt(digest[:])
		if err != nil {
			return xerrors.Errorf("SaveRequestLookup: %w", err)
		}
		if lst, err = RequestLookupKeyListFromBytes(bin); err != nil {
			return xerrors.Errorf("SaveRequestLookup: %w", err)
		}
	}
	for i := range lst {
		if lst[i] == key {
			// already in list. Not normal
			return xerrors.New("SaveRequestLookup: inconsistency: duplicate lookup key")
		}
	}
	lst = append(lst, key)
	if err := lookupTable.SetAt(digest[:], lst.Bytes()); err != nil {
		return xerrors.Errorf("SaveRequestLookup: %w", err)
	}
	// save the record. Key is a LookupKey
	if err = collections.NewMap(state, StateVarRequestRecords).SetAt(key.Bytes(), rec.Bytes()); err != nil {
		return xerrors.Errorf("SaveRequestLookup: %w", err)
	}
	return nil
}

func getBlockInfoIntern(ctx coretypes.SandboxView, blockIndex uint32) (*BlockInfo, bool) {
	data, found := getBlockInfoDataIntern(ctx.State(), blockIndex)
	if !found {
		return nil, false
	}
	a := assert.NewAssert(ctx.Log())
	ret, err := BlockInfoFromBytes(blockIndex, data)
	a.RequireNoError(err)
	return ret, true
}

func getBlockInfoDataIntern(state kv.KVStoreReader, blockIndex uint32) ([]byte, bool) {
	data := collections.NewArray32ReadOnly(state, StateVarBlockRegistry).MustGetAt(blockIndex)
	return data, data != nil
}

// RequestIsProcessed does quick lookup to check if it wasn't seen yet
func RequestIsProcessed(state kv.KVStoreReader, reqid *coretypes.RequestID) (bool, error) {
	lookupTable := collections.NewMapReadOnly(state, StateVarRequestLookupIndex)
	digest := reqid.LookupDigest()
	seen, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return false, err
	}
	if !seen {
		return false, nil
	}
	// the lookup record is here, have to check is it is nto a collision of digests
	bin := lookupTable.MustGetAt(digest[:])
	lst, err := RequestLookupKeyListFromBytes(bin)
	if err != nil {
		panic("RequestKnown: data conversion error")
	}
	records := collections.NewMapReadOnly(state, StateVarRequestRecords)
	for i := range lst {
		seen, err := records.HasAt(lst[i].Bytes())
		if err != nil {
			return false, err
		}
		if seen {
			return true, nil
		}
	}
	return false, nil
}

func getRequestRecordDataByRef(state kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(state, StateVarRequestRecords)
	recBin := lookupTable.MustGetAt(lookupKey[:])
	if recBin == nil {
		return nil, false
	}
	return recBin, true
}

func getRequestRecordDataByRequestID(ctx coretypes.SandboxView, reqID coretypes.RequestID) ([]byte, uint32, uint16, bool) {
	lookupDigest := reqID.LookupDigest()
	lookupTable := collections.NewMapReadOnly(ctx.State(), StateVarRequestLookupIndex)
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
		rec, err := RequestLogRecordFromBytes(recBin)
		a.RequireNoError(err)
		if rec.RequestID == reqID {
			return recBin, lookupKeyList[i].BlockIndex(), lookupKeyList[i].RequestIndex(), true
		}
	}
	return nil, 0, 0, false
}
