package iotago

import (
	"encoding/binary"
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/serialization"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

type (
	PackageID      = Address
	ObjectID       = Address
	SequenceNumber = uint64
	Identifier     = string
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

// ObjectRef for BCS, need to keep this order
type ObjectRef struct {
	ObjectID *ObjectID      `json:"objectId"`
	Version  SequenceNumber `json:"version"`
	Digest   *ObjectDigest  `json:"digest"`
}

func (o ObjectRef) Hash() hashing.HashValue {
	res, _ := hashing.HashValueFromBytes(o.Bytes())
	return res
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

type MoveObjectType struct {
	Other   *StructTag
	GasCoin *serialization.EmptyEnum
	Iota    *serialization.EmptyEnum
	Coin    *TypeTag
}

func (o MoveObjectType) IsBcsEnum() {}

func EqualSequenceNumbers(sn1, sn2 SequenceNumber) bool {
	return sn1 == sn2
}
