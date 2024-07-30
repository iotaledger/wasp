package isctypes

import "github.com/iotaledger/wasp/sui-go/sui"

type RequestEvent struct {
	RequestID sui.ObjectID
	Anchor    sui.Address
}
