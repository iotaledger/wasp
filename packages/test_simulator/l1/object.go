package l1

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type SimObject struct {
	ID         iotago.ObjectID
	Version    uint64
	Digest     iotago.Digest
	Owner      SimOwner
	Type       string // Move type, e.g. "0x...::anchor::Anchor"
	Data       []byte // BCS-encoded content
	PreviousTx iotago.TransactionDigest
}

func (o *SimObject) Clone() *SimObject {
	c := *o
	if o.Data != nil {
		c.Data = make([]byte, len(o.Data))
		copy(c.Data, o.Data)
	}
	return &c
}

func (o *SimObject) Ref() *iotago.ObjectRef {
	id := o.ID
	digest := o.Digest
	return &iotago.ObjectRef{
		ObjectID: &id,
		Version:  o.Version,
		Digest:   &digest,
	}
}

type SimOwner struct {
	AddressOwner *iotago.Address
	ObjectOwner  *iotago.Address // child object (dynamic field parent)
	Shared       *SharedInfo
	Immutable    bool
}

type SharedInfo struct {
	InitialSharedVersion uint64
}

type DynamicField struct {
	ParentID   iotago.ObjectID
	Name       DynFieldName
	ValueObjID iotago.ObjectID
}

type DynFieldName struct {
	TypeRepr string          // e.g. "0x1::ascii::String"
	JSON     json.RawMessage // e.g. `"0000...0002::iota::IOTA"`
	BCSKey   []byte          // for exact match lookups
}

type CoinObject struct {
	CoinType string // inner type, e.g. "0000...0002::iota::IOTA"
	Balance  uint64
}
