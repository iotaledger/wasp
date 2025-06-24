package iotago

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/serialization"
)

type Data struct {
	Move    *MoveObject
	Package *MovePackage
}

func (d Data) IsBcsEnum() {}

type MoveObject struct {
	Type              MoveObjectType
	HasPublicTransfer bool
	Version           SequenceNumber
	Contents          []uint8
}

type Owner struct {
	AddressOwner *Address `json:"AddressOwner"`
	ObjectOwner  *Address `json:"ObjectOwner"`
	Shared       *struct {
		InitialSharedVersion SequenceNumber `json:"initial_shared_version"`
	} `json:"Shared,omitempty"`
	Immutable *serialization.EmptyEnum `json:"Immutable,omitempty"`
}

func (o Owner) IsBcsEnum() {}

func (o Owner) Tag() string {
	return ""
}

func (o Owner) Content() string {
	return ""
}

type Object struct {
	Data                Data
	Owner               Owner
	PreviousTransaction TransactionDigest
	StorageRebate       uint64
}

// ObjectType is the representation of a Iota object type, e.g. `0x000...0002::iota::IOTA`
// Two instances of ObjectType are equal iif they represent the same type.
type ObjectType struct { // struct to enforce using the constructor functions
	s string
}

func ObjectTypeFromString(s string) (ObjectType, error) {
	rt, err := NewResourceType(s)
	if err != nil {
		return ObjectType{}, fmt.Errorf("invalid ObjectType %q: %w", s, err)
	}
	return ObjectType{s: rt.String()}, nil
}

func MustTypeFromString(s string) ObjectType {
	t, err := ObjectTypeFromString(s)
	if err != nil {
		panic(err)
	}
	return t
}

func (t *ObjectType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.s)
}

func (t *ObjectType) UnmarshalJSON(b []byte) error {
	return json.Unmarshal(b, &t.s)
}

func (t *ObjectType) MarshalBCS(e *bcs.Encoder) error {
	rt := t.ResourceType()
	e.Encode(rt)
	return nil
}

func (t *ObjectType) UnmarshalBCS(d *bcs.Decoder) error {
	rt := bcs.Decode[ResourceType](d)
	if d.Err() != nil {
		return d.Err()
	}
	*t = ObjectType{s: rt.String()}
	return nil
}

// MatchesStringType returns true if the given string represents the same coin
// type, even if abbreviated (e.g. ""0x2::iota::IOTA"")
func (t ObjectType) MatchesStringType(s string) bool {
	rt, err := ObjectTypeFromString(s)
	if err != nil {
		return false
	}
	return rt.String() == t.s
}

func (t ObjectType) String() string {
	return t.s
}

func (t ObjectType) ShortString() string {
	return t.ResourceType().ShortString()
}

func (t ObjectType) ResourceType() *ResourceType {
	return lo.Must(NewResourceType(t.s))
}

// EqualsStr  parses given object type string and checks it represents the same resource type.
func (t ObjectType) EqualsStr(other string) (bool, error) {
	return IsSameResource(t.s, other)
}

func (t ObjectType) TypeTag() TypeTag {
	coinTypeTag, err := TypeTagFromString(t.String())
	if err != nil {
		panic(err)
	}
	return *coinTypeTag
}

func (t ObjectType) Bytes() []byte {
	return bcs.MustMarshal(&t)
}

func ObjectTypeFromBytes(b []byte) (ObjectType, error) {
	var r ObjectType
	r, err := bcs.Unmarshal[ObjectType](b)
	return r, err
}

func CompareTypes(a, b ObjectType) int {
	return strings.Compare(a.s, b.s)
}
