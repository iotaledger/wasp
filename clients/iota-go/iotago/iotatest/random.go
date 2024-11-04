package iotatest

import (
	"math/rand"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
)

func RandomObjectRef() *iotago.ObjectRef {
	return &iotago.ObjectRef{
		ObjectID: RandomAddress(),
		Version:  rand.Uint64(),
		Digest:   RandomDigest(),
	}
}

func RandomAddress() *iotago.Address {
	var a iotago.Address
	_, _ = rand.Read(a[:])
	return &a
}

func RandomDigest() *iotago.Digest {
	var b [32]byte
	var d iotago.Digest
	_, _ = rand.Read(b[:])
	d = b[:]
	return &d
}

func RandomTransactionData() *iotago.TransactionData {
	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.Command(
		iotago.Command{
			MoveCall: &iotago.ProgrammableMoveCall{
				Package:       RandomAddress(),
				Module:        "test_module",
				Function:      "test_func",
				TypeArguments: []iotago.TypeTag{},
				Arguments:     []iotago.Argument{},
			},
		},
	)
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		RandomAddress(),
		pt,
		nil,
		10000,
		100,
	)
	return &tx
}

func RandomAnchor() *iscmove.Anchor {
	assetsBag := iscmove.AssetsBag{
		ID:   *RandomAddress(),
		Size: 0,
	}
	return &iscmove.Anchor{
		ID:            *RandomAddress(),
		Assets:        assetsBag,
		StateMetadata: nil,
		StateIndex:    0,
	}
}

func RandomStateAnchor() isc.StateAnchor {
	anchor := RandomAnchor()
	anchorRef := RandomObjectRef()
	return isc.NewStateAnchor(
		&iscmove.RefWithObject[iscmove.Anchor]{
			ObjectRef: *anchorRef,
			Object:    anchor,
		},
		cryptolib.NewRandomAddress(),
		*RandomAddress(),
	)
}
