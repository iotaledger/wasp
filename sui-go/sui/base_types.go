package sui

import "github.com/iotaledger/wasp/sui-go/sui/serialization"

type (
	PackageID      = Address
	ObjectID       = Address
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

type MoveObjectType struct {
	Other     *StructTag
	GasCoin   *serialization.EmptyEnum
	StakedSui *serialization.EmptyEnum
	Coin      *TypeTag
}

func (o MoveObjectType) IsBcsEnum() {}
