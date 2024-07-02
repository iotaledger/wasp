package suijsonrpc

import (
	"github.com/iotaledger/wasp/sui-go/sui"
)

type ObjectOwnerInternal struct {
	AddressOwner *sui.Address `json:"AddressOwner,omitempty"`
	ObjectOwner  *sui.Address `json:"ObjectOwner,omitempty"`
	SingleOwner  *sui.Address `json:"SingleOwner,omitempty"`
	Shared       *struct {
		InitialSharedVersion *sui.SequenceNumber `json:"initial_shared_version"`
	} `json:"Shared,omitempty"`
}

type ObjectOwner struct {
	*ObjectOwnerInternal
	*string
}

type Page[T SuiTransactionBlockResponse | SuiEvent | Coin | *Coin | SuiObjectResponse | DynamicFieldInfo | string | *Checkpoint,
	C sui.TransactionDigest | EventId | sui.ObjectID | BigInt] struct {
	Data []T `json:"data"`
	// 'NextCursor' points to the last item in the page.
	// Reading with next_cursor will start from the next item after next_cursor
	NextCursor  *C   `json:"nextCursor,omitempty"`
	HasNextPage bool `json:"hasNextPage"`
}
