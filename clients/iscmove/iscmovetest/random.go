// Package iscmovetest provides testing utilities for ISC move operations.
package iscmovetest

// Everything in this file should be test only

import (
	"math/rand"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iscmove"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/transaction"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/v2/packages/vm/gas"
)

type RandomAnchorOption struct {
	ID               *iotago.ObjectID
	Assets           *iscmove.AssetsBag
	AssetsReferentID *iotago.Address
	StateMetadata    *[]byte
	StateIndex       *uint32
}

func RandomAnchor(opts ...RandomAnchorOption) iscmove.Anchor {
	id := *iotatest.RandomAddress()
	assets := iscmove.AssetsBag{
		ID:   *iotatest.RandomAddress(),
		Size: safecast.MustConvert[uint64](rand.Int63()),
	}
	assetsReferentID := *iotatest.RandomAddress()
	schemaVersion := allmigrations.DefaultScheme.LatestSchemaVersion()
	initParams := isc.NewCallArguments([]byte{1, 2, 3})
	stateMetadata := transaction.NewStateMetadata(
		schemaVersion,
		&state.L1Commitment{}, // FIXME properly set trieRoot, blockHash
		&iotago.ObjectID{},
		gas.DefaultFeePolicy(),
		initParams,
		0,
		"http://url",
	).Bytes()
	stateIndex := safecast.MustConvert[uint32](rand.Int31())
	if len(opts) > 0 {
		if opts[0].ID != nil {
			id = *opts[0].ID
		}
		if opts[0].Assets != nil {
			assets = *opts[0].Assets
		}
		if opts[0].StateMetadata != nil {
			stateMetadata = *opts[0].StateMetadata
		}
		if opts[0].StateIndex != nil {
			stateIndex = *opts[0].StateIndex
		}
		if opts[0].AssetsReferentID != nil {
			assetsReferentID = *opts[0].AssetsReferentID
		}
	}
	return iscmove.Anchor{
		ID: id,
		Assets: iscmove.Referent[iscmove.AssetsBag]{
			ID:    assetsReferentID,
			Value: &assets,
		},
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
