package suisignertest

import (
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui/suitest"
	"github.com/iotaledger/wasp/sui-go/suisigner"
)

// TEST ONLY functions

func RandomSignedTransaction(signer ...suisigner.Signer) suisigner.SignedTransaction {
	tx := suitest.RandomTransactionData()
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	if err != nil {
		panic(err)
	}
	signature, err := signer[0].SignTransactionBlock(txBytes, suisigner.DefaultIntent())
	if err != nil {
		panic(err)
	}
	return *suisigner.NewSignedTransaction(tx, signature)
}
