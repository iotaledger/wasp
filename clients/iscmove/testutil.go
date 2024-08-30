package iscmove

import "github.com/iotaledger/wasp/sui-go/sui"

func RandomAnchor() Anchor {
	return Anchor{
		ID: sui.RandomObjectID(),
	}
}
