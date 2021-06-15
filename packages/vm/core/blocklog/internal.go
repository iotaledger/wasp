package blocklog

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

// SaveNextBlockInfo appends block info and returns its index
func SaveNextBlockInfo(partition kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(partition, StateVarBlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	ret := registry.MustLen() - 1
	return ret
}

// SaveControlAddressesIfNecessary saves new information about state address in the blocklog partition
// If state address does not change, it does nothing
func SaveControlAddressesIfNecessary(partition kv.KVStore, stateAddress, governingAddress ledgerstate.Address, blockIndex uint32) {
	registry := collections.NewArray32(partition, StateVarControlAddresses)
	l := registry.MustLen()
	if l != 0 {
		addrs, err := ControlAddressesFromBytes(registry.MustGetAt(l - 1))
		if err != nil {
			panic(fmt.Sprintf("SaveControlAddressesIfNecessary: %v", err))
		}
		if addrs.StateAddress.Equals(stateAddress) && addrs.GoverningAddress.Equals(governingAddress) {
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
func SaveRequestLogRecord(partition kv.KVStore, rec *RequestLogRecord, key RequestLookupKey) error {
	// save lookup record for fast lookup
	lookupTable := collections.NewMap(partition, StateVarRequestLookupIndex)
	digest := rec.RequestID.LookupDigest()
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
	if err = collections.NewMap(partition, StateVarRequestRecords).SetAt(key.Bytes(), rec.Bytes()); err != nil {
		return xerrors.Errorf("SaveRequestLogRecord: %w", err)
	}
	return nil
}

// isRequestProcessedIntern does quick lookup to check if it wasn't seen yet
func isRequestProcessedIntern(partition kv.KVStoreReader, reqid *coretypes.RequestID) (bool, error) {
	lookupTable := collections.NewMapReadOnly(partition, StateVarRequestLookupIndex)
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
	records := collections.NewMapReadOnly(partition, StateVarRequestRecords)
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

func getRequestLogRecordsForBlockBin(state kv.KVStoreReader, blockIndex uint32) ([][]byte, bool, error) {
	if blockIndex == 0 {
		return nil, true, nil
	}
	blockInfoBin, found, err := getBlockInfoDataIntern(state, blockIndex)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}
	blockInfo, err := BlockInfoFromBytes(blockIndex, blockInfoBin)
	if err != nil {
		return nil, false, err
	}

	ret := make([][]byte, blockInfo.TotalRequests)
	for reqIdx := uint16(0); reqIdx < blockInfo.TotalRequests; reqIdx++ {
		ret[reqIdx], found = getRequestRecordDataByRef(state, blockIndex, reqIdx)
		if !found {
			panic("getRequestLogRecordsForBlockBin: inconsistency: request record not found")
		}
	}
	return ret, true, nil
}

func getBlockInfoDataIntern(partition kv.KVStoreReader, blockIndex uint32) ([]byte, bool, error) {
	data, err := collections.NewArray32ReadOnly(partition, StateVarBlockRegistry).GetAt(blockIndex)
	return data, err == nil, err
}

func getRequestRecordDataByRef(partition kv.KVStoreReader, blockIndex uint32, requestIndex uint16) ([]byte, bool) {
	lookupKey := NewRequestLookupKey(blockIndex, requestIndex)
	lookupTable := collections.NewMapReadOnly(partition, StateVarRequestRecords)
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
