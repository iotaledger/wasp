package graphqltypes

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

type OwnerType string

const (
	OwnerTypeAddress   OwnerType = "address"
	OwnerTypeObject    OwnerType = "object"
	OwnerTypeShared    OwnerType = "shared"
	OwnerTypeImmutable OwnerType = "immutable"
)

type Owner struct {
	Type OwnerType

	// Address is populated for AddressOwner and ObjectOwner types
	Address *iotago.Address

	// InitialSharedVersion is populated for Shared type
	InitialSharedVersion *uint64
}
