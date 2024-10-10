package suisignertest

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago/suitest"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

// TEST ONLY functions

func RandomSignedTransaction(signer ...iotasigner.Signer) iotasigner.SignedTransaction {
	tx := suitest.RandomTransactionData()
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	if err != nil {
		panic(err)
	}
	signature, err := signer[0].SignTransactionBlock(txBytes, iotasigner.DefaultIntent())
	if err != nil {
		panic(err)
	}
	return *iotasigner.NewSignedTransaction(tx, signature)
}
