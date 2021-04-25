package blocklog

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func SaveNextBlockInfo(state kv.KVStore, blockInfo *BlockInfo) uint32 {
	registry := collections.NewArray32(state, BlockRegistry)
	registry.MustPush(blockInfo.Bytes())
	return registry.MustLen() - 1
}

func SaveRequestLookup(state kv.KVStore, reqid *coretypes.RequestID, blockIndex uint32, requestIndex uint16) {
	lookupTable := collections.NewMap(state, RequestLookupMap)
	digest := reqid.LookupDigest()
	bin, _ := lookupTable.GetAt(digest[:])
	var lst *RequestLookupList
	if bin == nil {
		lst = NewRequestLookupList()
	} else {
		var err error
		if lst, err = RequestLookupListFromBytes(bin); err != nil {
			panic("SaveRequestLookup: data conversion error")
		}
	}
	l := lst.List()
	for i := range l {
		if l[i].BlockIndex == blockIndex && l[i].RequestIndex == requestIndex {
			return
		}
	}
	lst.Append(RequestBlockReference{BlockIndex: blockIndex, RequestIndex: requestIndex})
	_ = lookupTable.SetAt(digest[:], lst.Bytes())
}

func RequestKnown(state kv.KVStore, reqid *coretypes.RequestID) bool {
	lookupTable := collections.NewMap(state, RequestLookupMap)
	digest := reqid.LookupDigest()
	ok, err := lookupTable.HasAt(digest[:])
	if err != nil || !ok {
		return false
	}
	bin := lookupTable.MustGetAt(digest[:])
	lst, err := RequestLookupListFromBytes(bin)
	if err != nil {
		panic("RequestKnown: data conversion error")
	}
	lst = lst
	// TODO not finished
	return false
}
