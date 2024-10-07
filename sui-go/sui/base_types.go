package sui

import (
	"encoding/binary"
	"fmt"

	"github.com/iotaledger/wasp/sui-go/sui/serialization"
)

type (
	PackageID      = Address
	ObjectID       = Address
	ObjectIDKey    = [AddressLen]byte
	SequenceNumber = uint64
	Identifier     = string
	ObjectType     = string
)

func PackageIDFromHex(str string) (*PackageID, error) {
	return AddressFromHex(str)
}

func MustPackageIDFromHex(str string) *PackageID {
	packageID, err := AddressFromHex(str)
	if err != nil {
		panic(err)
	}
	return packageID
}

func ObjectIDFromArray(objectID [AddressLen]byte) *ObjectID {
	return AddressFromArray(objectID)
}

func ObjectIDFromHex(str string) (*ObjectID, error) {
	return AddressFromHex(str)
}

func MustObjectIDFromHex(str string) *ObjectID {
	objectID, err := AddressFromHex(str)
	if err != nil {
		panic(err)
	}
	return objectID
}

func (id ObjectID) Key() ObjectIDKey {
	var result ObjectIDKey
	copy(result[:], id.Bytes())
	return result
}

type ObjectInfo struct {
	Ref  *ObjectRef
	Type *ResourceType
}

func NewObjectInfo(ref *ObjectRef, objType *ResourceType) *ObjectInfo {
	var info ObjectInfo
	info.Ref = ref
	info.Type = objType
	return &info
}

type ObjectRefKey [AddressLen + 8]byte

// ObjectRef for BCS, need to keep this order
type ObjectRef struct {
	ObjectID *ObjectID      `json:"objectId"`
	Version  SequenceNumber `json:"version"`
	Digest   *ObjectDigest  `json:"digest"`
}

func ObjectRefFromBytes(b []byte) *ObjectRef {
	var ref ObjectRef
	ref.Version = binary.LittleEndian.Uint64(b[:8])
	var arr [AddressLen]byte
	copy(arr[:], b[8:])
	ref.ObjectID = AddressFromArray(arr)
	// TODO: Why Digest is not read here?
	return &ref
}

func (or *ObjectRef) Equals(other *ObjectRef) bool {
	if or == nil {
		return other == nil
	}
	return or.ObjectID.Equals(*other.ObjectID) &&
		EqualSequenceNumbers(or.Version, other.Version) &&
		or.Digest.Equals(*other.Digest)
}

func (or *ObjectRef) String() string {
	return fmt.Sprintf("obj{id=%s, version=%v, digest=%s}", or.ObjectID.String(), or.Version, or.Digest.String())
}

func (or *ObjectRef) Bytes() []byte {
	version := make([]byte, 8)
	binary.LittleEndian.PutUint64(version, or.Version)
	result := or.ObjectID[:]
	result = append(result, version...)
	// TODO maybe we should add digest here too
	return result
}

func (or *ObjectRef) Key() ObjectRefKey {
	var result ObjectRefKey
	copy(result[:], or.Bytes())
	return result
}

type MoveObjectType struct {
	Other     *StructTag
	GasCoin   *serialization.EmptyEnum
	StakedSui *serialization.EmptyEnum
	Coin      *TypeTag
}

func (o MoveObjectType) IsBcsEnum() {}

func EqualSequenceNumbers(sn1, sn2 SequenceNumber) bool {
	return sn1 == sn2
}
