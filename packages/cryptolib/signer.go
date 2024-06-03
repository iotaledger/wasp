package cryptolib

import iotago "github.com/iotaledger/iota.go/v3"

// VariantKeyPair originates from KeyPair
type Signer interface {
	// IsNil is a mandatory nil check. This includes the referenced keypair implementation pointer. `kp == nil` is not enough.
	//IsNil() bool

	//GetPublicKey() *PublicKey
	Address() *Address
	//AsAddressSigner() iotago.AddressSigner
	//AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys
	SignBytes(data []byte) []byte
	Sign(msg []byte) (signature *Signature, err error)
}

type iotagoSigner struct {
	s Signer
}

// TODO: remove, when it is not needed
func SignerToIotago(s Signer) iotago.AddressSigner {
	return &iotagoSigner{s}
}

func (is *iotagoSigner) Sign(addr iotago.Address, msg []byte) (iotago.Signature, error) {
	signature, err := is.s.Sign(msg)
	if err != nil {
		return nil, err
	}
	result := &iotago.Ed25519Signature{}
	copy(result.PublicKey[:], signature.publicKey.key)
	copy(result.Signature[:], signature.signature[:])
	return result, nil
}
