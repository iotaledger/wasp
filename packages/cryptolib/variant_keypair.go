package cryptolib

import (
	iotago "github.com/iotaledger/iota.go/v3"
)

// VariantKeyPair originates from cryptolib.KeyPair
type VariantKeyPair interface {
	GetPublicKey() *PublicKey
	Address() *iotago.Ed25519Address
	AsAddressSigner() iotago.AddressSigner
	AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys
	SignBytes(data []byte) []byte
	Sign(addr iotago.Address, msg []byte) (signature iotago.Signature, err error)
}
