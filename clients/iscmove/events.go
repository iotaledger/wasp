package iscmove

import "github.com/iotaledger/wasp/sui-go/sui_types"

type RequestEvent struct {
	RequestID sui_types.ObjectID
	Anchor    sui_types.SuiAddress
}
