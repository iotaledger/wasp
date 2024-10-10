package suisignertest

import (
	"github.com/iotaledger/wasp/clients/iota-go/sui/suitest"
	"github.com/iotaledger/wasp/clients/iota-go/suisigner"
	"github.com/iotaledger/wasp/packages/util/bcs"
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
