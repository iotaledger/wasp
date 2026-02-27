package graphqltypes

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

func (v *RPC_OBJECT_FIELDS) ObjectRef() (*iotago.ObjectRef, error) {
	digest, err := iotago.NewDigest(v.Digest)
	if err != nil {
		return nil, err
	}
	objectID := v.ObjectId
	return &iotago.ObjectRef{
		ObjectID: &objectID,
		Version:  v.Version,
		Digest:   digest,
	}, nil
}

func (v *RPC_OBJECT_FIELDS) ObjectID() iotago.ObjectID {
	return v.ObjectId
}

func (v *RPC_OBJECT_FIELDS) BcsBytes() iotago.Base64Data {
	return v.AsMoveObject.Contents.Bcs
}

// TypeRepr returns the Move type repr string if ShowType or ShowContent was requested.
func (v *RPC_OBJECT_FIELDS) TypeRepr() string {
	if repr := v.AsMoveObjectType.Contents.Type.Repr; repr != "" {
		return repr
	}
	if repr := v.AsMoveObjectContent.Contents.Type.Repr; repr != "" {
		return repr
	}
	return v.AsMoveObject.Contents.Type.Repr
}

// OwnerAddress extracts the owner address from the GraphQL owner union.
// Returns nil if the owner is not an address owner or was not requested.
func (v *RPC_OBJECT_FIELDS) OwnerAddress() *iotago.Address {
	if v.Owner == nil {
		return nil
	}
	o, ok := v.Owner.(*RPC_OBJECT_FIELDSOwnerAddressOwner)
	if !ok {
		return nil
	}
	if o.Owner.AsAddress.Address != (iotago.Address{}) {
		addr := o.Owner.AsAddress.Address
		return &addr
	}
	if o.Owner.AsObject.Address != (iotago.Address{}) {
		addr := o.Owner.AsObject.Address
		return &addr
	}
	return nil
}

// IsDeleted returns true if the object has been wrapped or deleted.
func (v *RPC_OBJECT_FIELDS) IsDeleted() bool {
	return v.Status == ObjectKindWrappedOrDeleted
}

// IsNotFound returns true if the object was not found (zero address).
func (v *RPC_OBJECT_FIELDS) IsNotFound() bool {
	return v.ObjectId == iotago.Address{}
}
