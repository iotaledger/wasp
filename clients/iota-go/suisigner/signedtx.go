package suisigner

import (
	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/packages/hashing"
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

func (st *SignedTransaction) Hash() hashing.HashValue {
	panic("SignedTransaction.Hash not implemented") // TODO: Implement it.
}

// We use it to find the consumed anchor ref from the TX.
func (st *SignedTransaction) FindInputByID(id sui.ObjectID) *sui.ObjectRef {
	if st == nil || st.Data == nil || st.Data.V1 == nil || st.Data.V1.Kind.ProgrammableTransaction == nil {
		return nil
	}
	for _, input := range st.Data.V1.Kind.ProgrammableTransaction.Inputs {
		if input.Object == nil || input.Object.ImmOrOwnedObject == nil {
			continue
		}
		if input.Object.ImmOrOwnedObject.ObjectID.Equals(id) {
			return input.Object.ImmOrOwnedObject
		}
	}
	return nil
}
