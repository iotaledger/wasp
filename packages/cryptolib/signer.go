package cryptolib

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
