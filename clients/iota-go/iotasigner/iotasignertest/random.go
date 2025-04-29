package iotasignertest

import (
	"crypto/rand"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
)

func RandomSigner() iotasigner.Signer {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return iotasigner.NewSigner(b, iotasigner.KeySchemeFlagDefault)
}

func RandomSignedTransaction(signers ...iotasigner.Signer) iotasigner.SignedTransaction {
	tx := iotatest.RandomTransactionData()
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	if err != nil {
		panic(err)
	}
	var signer iotasigner.Signer
	if len(signers) == 0 {
		signer = RandomSigner()
	}
	signature, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	if err != nil {
		panic(err)
	}
	return *iotasigner.NewSignedTransaction(tx, signature)
}
