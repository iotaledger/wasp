package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

func SaveNextBlockInfo(state kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(state, StateVarBlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	return registry.MustLen() - 1
}

func SaveRequestLookup(state kv.KVStore, reqid *coretypes.RequestID, key coretypes.RequestLookupKey) error {
	lookupTable := collections.NewMap(state, StateVarRequestLookupIndex)
	digest := reqid.LookupDigest()
	bin, _ := lookupTable.GetAt(digest[:])
	var lst request.RequestLookupKeyList
	if bin == nil {
		lst = make(request.RequestLookupKeyList, 0)
	} else {
		var err error
		if lst, err = request.RequestLookupKeyListFromBytes(bin); err != nil {
			return xerrors.New("SaveRequestLookup: data conversion error")
		}
	}
	for i := range lst {
		if lst[i] == key {
			// already in list
			return nil
		}
	}
	lst = append(lst, key)
	return lookupTable.SetAt(digest[:], lst.Bytes())
}

func SaveRequestLogRecord(state kv.KVStore, rec *request.RequestLogRecord, key coretypes.RequestLookupKey) {
	_ = collections.NewMap(state, StateVarRequestRecords).SetAt(key.Bytes(), rec.Bytes())
}

// RequestNotSeen does quick lookup to check if it wasn't seen yet
func RequestNotSeen(state kv.KVStore, reqid *coretypes.RequestID) (bool, error) {
	lookupTable := collections.NewMap(state, StateVarRequestLookupIndex)
	digest := reqid.LookupDigest()
	seen, err := lookupTable.HasAt(digest[:])
	if err != nil {
		return false, err
	}
	if !seen {
		return true, nil
	}
	// the lookup record is here, have to check is it is nto a collision of digests
	bin := lookupTable.MustGetAt(digest[:])
	lst, err := request.RequestLookupKeyListFromBytes(bin)
	if err != nil {
		panic("RequestKnown: data conversion error")
	}
	records := collections.NewMap(state, StateVarRequestRecords)
	for i := range lst {
		seen, err := records.HasAt(lst[i].Bytes())
		if err != nil {
			return false, err
		}
		if seen {
			return false, nil
		}
	}
	return true, nil
}
