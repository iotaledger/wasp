package accounts

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// ObjectRecord represents a L1 generic object owned by the chain (e.g. NFT)
type ObjectRecord struct {
	ID  sui.ObjectID `bcs:"-"` // transient
	BCS []byte
}

func ObjectRecordFromBytes(data []byte, id sui.ObjectID) (*ObjectRecord, error) {
	return bcs.UnmarshalOver(data, &ObjectRecord{ID: id})
}

func (rec *ObjectRecord) Bytes() []byte {
	return bcs.MustMarshal(rec)
}

var emptyObjectID = sui.ObjectID{}

func (rec *ObjectRecord) CollectionKey() kv.Key {
	// TODO: parse NFT data and determine the NFT's collection
	return noCollection
}
