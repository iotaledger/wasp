package cryptolib

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

type KeyPair struct {
	privateKey PrivateKey
	publicKey  PublicKey
}

// NewKeyPair creates a new key pair with a randomly generated seed
func NewKeyPair() KeyPair {
	seed := tpkg.RandEd25519Seed()
	key := NewKeyPairFromSeed(SeedFromByteArray(seed[:]))

	return key
}

func NewKeyPairFromSeed(seed Seed) KeyPair {
	privateKey := NewPrivateKeyFromSeed(seed)
	return NewKeyPairFromPrivateKey(privateKey)
}

func NewKeyPairFromPrivateKey(privateKey PrivateKey) KeyPair {
	publicKey := privateKey.Public()
	return KeyPair{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (k *KeyPair) Valid() bool {
	return len(k.privateKey.asCrypto()) > 0
}

func (k *KeyPair) Verify(message, sig []byte) bool {
	return k.publicKey.Verify(message, sig)
}

func (k *KeyPair) AsAddressSigner() iotago.AddressSigner {
	addrKeys := iotago.NewAddressKeysForEd25519Address(k.publicKey.AsEd25519Address(), k.privateKey.asCrypto())
	return iotago.NewInMemoryAddressSigner(addrKeys)
}

func (k *KeyPair) GetPrivateKey() PrivateKey {
	return k.privateKey
}

func (k *KeyPair) GetPublicKey() PublicKey {
	return k.publicKey
}
