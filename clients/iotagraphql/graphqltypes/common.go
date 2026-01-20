// Package graphqltypes provides types for GraphQL API responses and requests.
// Field names follow JSON serialization conventions for API compatibility.
//
//nolint:stylecheck // Field names use JSON conventions (Id, Url, Json) for serialization
package graphqltypes

import "github.com/iotaledger/wasp/v2/clients/iota-go/iotago"

type ObjectOwnerInternal struct {
	AddressOwner *iotago.Address `json:"AddressOwner,omitempty"`
	ObjectOwner  *iotago.Address `json:"ObjectOwner,omitempty"`
	SingleOwner  *iotago.Address `json:"SingleOwner,omitempty"`
	Shared       *struct {
		InitialSharedVersion *iotago.SequenceNumber `json:"initial_shared_version"`
	} `json:"Shared,omitempty"`
}

func (o ObjectOwnerInternal) IsBcsEnum() {}

type ObjectOwner struct {
	*ObjectOwnerInternal
	*string
}

func (o ObjectOwner) IsBcsEnum() {}

type Page[T IotaTransactionBlockResponse | IotaEvent | Coin | *Coin | IotaObjectResponse | DynamicFieldInfo | string | *Checkpoint,
	C iotago.TransactionDigest | EventID | iotago.ObjectID | BigInt | string] struct {
	Data []T `json:"data"`
	// 'NextCursor' points to the last item in the page.
	// Reading with next_cursor will start from the next item after next_cursor
	NextCursor  *C   `json:"nextCursor,omitempty"`
	HasNextPage bool `json:"hasNextPage"`
}
