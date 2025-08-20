package iotasigner

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
)

type SignedTransaction struct {
	Data       *iotago.TransactionData
	Signatures []*Signature
}

func NewSignedTransaction(unsignedTx *iotago.TransactionData, signature *Signature) *SignedTransaction {
	return &SignedTransaction{
		Data:       unsignedTx,
		Signatures: []*Signature{signature},
	}
}

func (st *SignedTransaction) Digest() (*iotago.Digest, error) {
	digest, err := st.Data.Digest()
	if err != nil {
		return nil, err
	}
	return digest, nil
}

// We use it to find the consumed anchor ref from the TX.
func (st *SignedTransaction) FindInputByID(id iotago.ObjectID) *iotago.ObjectRef {
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
