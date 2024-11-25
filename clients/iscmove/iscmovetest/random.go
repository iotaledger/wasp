package iscmovetest

// Everything in this file should be test only

import (
	"math/rand"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

type RandomAnchorOption struct {
	ID            *iotago.ObjectID
	Assets        *iscmove.AssetsBag
	StateMetadata *[]byte
	StateIndex    *uint32
}

func RandomAnchor(opt ...RandomAnchorOption) iscmove.Anchor {
	id := *iotatest.RandomAddress()
	assets := iscmove.AssetsBag{
		ID:   *iotatest.RandomAddress(),
		Size: uint64(rand.Int63()),
	}
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := isc.NewCallArguments([]byte{1, 2, 3})
	stateMetadata := transaction.NewStateMetadata(
		schemaVersion,
		&state.L1Commitment{}, // FIXME properly set trieRoot, blockHash
		iotago.ObjectID{},
		gas.DefaultFeePolicy(),
		initParams,
		"http://url",
	).Bytes()
	stateIndex := uint32(rand.Int31())
	if len(opt) > 0 {
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
		ID:   *iotatest.RandomAddress(),
		Size: 0,
	}
}

func RandomMessage() *iscmove.Message {
	return &iscmove.Message{
		Contract: uint32(isc.Hn("test_isc_contract")),
		Function: uint32(isc.Hn("test_isc_func")),
		Args:     [][]byte{[]byte("one"), []byte("two"), []byte("three")},
	}
}
