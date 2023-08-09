package cryptolib

import (
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type KeyPair struct {
	privateKey *PrivateKey
	publicKey  *PublicKey
}

// NewKeyPair creates a new key pair with a randomly generated seed
func NewKeyPair() *KeyPair {
	privateKey := NewPrivateKey()
	return KeyPairFromPrivateKey(privateKey)
}

func KeyPairFromSeed(seed Seed) *KeyPair {
	privateKey := PrivateKeyFromSeed(seed)
	return KeyPairFromPrivateKey(privateKey)
}

func KeyPairFromPrivateKey(privateKey *PrivateKey) *KeyPair {
	publicKey := privateKey.Public()
	return &KeyPair{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (k *KeyPair) IsValid() bool {
	return k.privateKey.isValid()
}

func (k *KeyPair) Verify(message, sig []byte) bool {
	return k.publicKey.Verify(message, sig)
}

func (k *KeyPair) AsAddressSigner() iotago.AddressSigner {
	addrKeys := k.privateKey.AddressKeysForEd25519Address(k.publicKey.AsEd25519Address())
	return iotago.NewInMemoryAddressSigner(addrKeys)
}

func (k *KeyPair) GetPrivateKey() *PrivateKey {
	return k.privateKey
}

func (k *KeyPair) GetPublicKey() *PublicKey {
	return k.publicKey
}

func (k *KeyPair) Address() *iotago.Ed25519Address {
	return k.GetPublicKey().AsEd25519Address()
}

func (k *KeyPair) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	k.publicKey = new(PublicKey)
	rr.Read(k.publicKey)
	k.privateKey = new(PrivateKey)
	rr.Read(k.privateKey)
	return rr.Err
}

func (k *KeyPair) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.Write(k.publicKey)
	ww.Write(k.privateKey)
	return ww.Err
}
