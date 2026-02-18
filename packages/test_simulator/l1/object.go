package l1

import (
	"encoding/json"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

// SimObject represents an object in the simulated L1 store.
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

// SimOwner describes the ownership of a simulated object.
type SimOwner struct {
	AddressOwner *iotago.Address
	ObjectOwner  *iotago.Address // child object (dynamic field parent)
	Shared       *SharedInfo
	Immutable    bool
}

type SharedInfo struct {
	InitialSharedVersion uint64
}

// DynamicField represents a dynamic field entry linking parent to child object.
type DynamicField struct {
	ParentID   iotago.ObjectID
	Name       DynFieldName
	ValueObjID iotago.ObjectID
}

// DynFieldName identifies a dynamic field key.
type DynFieldName struct {
	TypeRepr string          // e.g. "0x1::ascii::String"
	Json     json.RawMessage // e.g. `"0000...0002::iota::IOTA"`
	BCSKey   []byte          // for exact match lookups
}

// CoinObject is a convenience struct for coin objects in the store.
// The Data field in SimObject contains BCS(Balance{value}).
type CoinObject struct {
	CoinType string // inner type, e.g. "0000...0002::iota::IOTA"
	Balance  uint64
}
