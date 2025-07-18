package iotasignertest

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/testutil"
)

var TestSeedEd25519 = testutil.TestBytes(32)

var TestSigner = iotasigner.NewSignerByIndex(
	TestSeedEd25519,
	iotasigner.KeySchemeFlagEd25519,
	123,
)

var TestSignedTransaction = testSignedTransaction(
	iotatest.TestTransactionData,
	TestSigner,
)

func testSignedTransaction(tx *iotago.TransactionData, signer iotasigner.Signer) iotasigner.SignedTransaction {
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	if err != nil {
		panic(err)
	}
	signature, err := signer.SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	if err != nil {
		panic(err)
	}
	return *iotasigner.NewSignedTransaction(tx, signature)
}
