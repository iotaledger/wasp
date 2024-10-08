package suitest

import (
	"math/rand"

	"github.com/iotaledger/wasp/sui-go/sui"
)

func RandomObjectRef() *sui.ObjectRef {
	return &sui.ObjectRef{
		ObjectID: RandomAddress(),
		Version:  rand.Uint64(),
		Digest:   RandomDigest(),
	}
}

func RandomAddress() *sui.Address {
	var a sui.Address
	_, _ = rand.Read(a[:])
	return &a
}

func RandomDigest() *sui.Digest {
	var b [32]byte
	var d sui.Digest
	_, _ = rand.Read(b[:])
	d = b[:]
	return &d
}

func RandomTransactionData() *sui.TransactionData {
	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.Command(sui.Command{MoveCall: &sui.ProgrammableMoveCall{
		Package:       RandomAddress(),
		Module:        "test_module",
		Function:      "test_func",
		TypeArguments: []sui.TypeTag{},
		Arguments:     []sui.Argument{},
	}})
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		RandomAddress(),
		pt,
		nil,
		10000,
		100,
	)
	return &tx
}
