package iotatest

import (
	"math/rand"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
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
	var d iotago.Digest
	_, _ = rand.Read(d[:])
	return &d
}

func RandomTransactionData() *iotago.TransactionData {
	return testTransactionData(RandomAddress(), RandomAddress())
}
