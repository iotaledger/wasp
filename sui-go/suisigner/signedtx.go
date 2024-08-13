package suisigner

import (
	"github.com/iotaledger/wasp/sui-go/sui"
)

type SignedTransaction struct {
	Data       *sui.TransactionData
	Signatures []*Signature
}

func NewSignedTransaction(unsignedTx *sui.TransactionData, signature *Signature) *SignedTransaction {
	return &SignedTransaction{
		Data:       unsignedTx,
		Signatures: []*Signature{signature},
	}
}
