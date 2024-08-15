package accounts

import (
	"io"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
)

// ObjectRecord represents a L1 generic object owned by the chain (e.g. NFT)
type ObjectRecord struct {
	ID  sui.ObjectID // transient
	BCS []byte
}

func ObjectRecordFromBytes(data []byte, id sui.ObjectID) (*ObjectRecord, error) {
	return rwutil.ReadFromBytes(data, &ObjectRecord{ID: id})
}

func (rec *ObjectRecord) Bytes() []byte {
	return rwutil.WriteToBytes(rec)
}

var emptyObjectID = sui.ObjectID{}

func (rec *ObjectRecord) Read(r io.Reader) error {
	if rec.ID == emptyObjectID {
		panic("unknown ObjectID for ObjectRecord")
	}
	rr := rwutil.NewReader(r)
	rec.BCS = rr.ReadBytes()
	return rr.Err
}

func (rec *ObjectRecord) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteBytes(rec.BCS)
	return ww.Err
}

func (rec *ObjectRecord) CollectionKey() kv.Key {
	// TODO: parse NFT data and determine the NFT's collection
	return noCollection
}
