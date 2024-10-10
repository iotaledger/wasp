package cryptolib

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
)

// VariantKeyPair originates from KeyPair
type Signer interface {
	// IsNil is a mandatory nil check. This includes the referenced keypair implementation pointer. `kp == nil` is not enough.
	// IsNil() bool

	Address() *Address
	Sign(msg []byte) (signature *Signature, err error)
	SignTransactionBlock(txnBytes []byte, intent iotasigner.Intent) (*Signature, error)
}

type suiSigner struct {
	s Signer
}

// TODO: remove, when it is not needed
func SignerToSuiSigner(s Signer) iotasigner.Signer {
	return &suiSigner{s}
}

func (is *suiSigner) Address() *iotago.Address {
	return is.s.Address().AsSuiAddress()
}

func (is *suiSigner) Sign(msg []byte) (signature *iotasigner.Signature, err error) {
	b, err := is.s.Sign(msg)
	if err != nil {
		return nil, err
	}
	return b.AsSuiSignature(), err
}

func (is *suiSigner) SignTransactionBlock(txnBytes []byte, intent iotasigner.Intent) (*iotasigner.Signature, error) {
	signature, err := is.s.SignTransactionBlock(txnBytes, intent)
	if err != nil {
		return nil, err
	}
	return signature.AsSuiSignature(), nil
}
