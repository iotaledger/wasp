package accounts

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/kv"
)

// ObjectRecord represents a L1 generic object owned by the chain (e.g. NFT)
type ObjectRecord struct {
	ID  iotago.ObjectID // transient
	BCS []byte
}

func ObjectRecordFromBytes(data []byte, id iotago.ObjectID) (*ObjectRecord, error) {
	return bcs.UnmarshalInto(data, &ObjectRecord{ID: id})
}

func (rec *ObjectRecord) Bytes() []byte {
	return bcs.MustMarshal(rec)
}

var emptyObjectID = iotago.ObjectID{}

func (rec *ObjectRecord) CollectionKey() kv.Key {
	// TODO: parse NFT data and determine the NFT's collection
	return NoCollection
}
