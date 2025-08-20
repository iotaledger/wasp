package iotasignertest

import (
	"crypto/rand"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/samber/lo"
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
	return testSignedTransaction(iotatest.RandomTransactionData(), lo.FirstOr(signers, RandomSigner()))
}
