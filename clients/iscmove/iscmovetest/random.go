package iscmovetest

// Everything in this file should be test only

import (
	"math/rand"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/suitest"
	"github.com/iotaledger/wasp/clients/iscmove"
)

type RandomAnchorOption struct {
	ID            *iotago.ObjectID
	Assets        *iscmove.AssetsBag
	StateMetadata *[]byte
	StateIndex    *uint32
}

func RandomAnchor(opt ...RandomAnchorOption) iscmove.Anchor {
	id := *suitest.RandomAddress()
	assets := iscmove.AssetsBag{
		ID:   *suitest.RandomAddress(),
		Size: uint64(rand.Int63()),
	}
	stateMetadata := make([]byte, 128)
	rand.Read(stateMetadata)
	stateIndex := uint32(rand.Int31())
	if opt[0].ID != nil {
		id = *opt[0].ID
	}
	if opt[0].Assets != nil {
		assets = *opt[0].Assets
	}
	if opt[0].StateMetadata != nil {
		stateMetadata = *opt[0].StateMetadata
	}
	if opt[0].StateIndex != nil {
		stateIndex = *opt[0].StateIndex
	}
	return iscmove.Anchor{
		ID:            id,
		Assets:        assets,
		StateMetadata: stateMetadata,
		StateIndex:    stateIndex,
	}
}

func RandomAssetsBag() iscmove.AssetsBag {
	return iscmove.AssetsBag{
		ID:   *suitest.RandomAddress(),
		Size: 0,
	}
}
