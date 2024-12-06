package iotatest

import (
	"math/rand"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
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
				TypeArguments: nil,
				Arguments:     nil,
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
