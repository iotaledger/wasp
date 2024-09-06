package cryptolib

import (
	"crypto/ed25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/sui-go/suisigner"
)

type KeyPair struct {
	privateKey *PrivateKey `bcs:""`
	publicKey  *PublicKey  `bcs:""`
}

var (
	_ Signer = &KeyPair{}
)

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

func (k *KeyPair) IsNil() bool {
	return k == nil
}

func (k *KeyPair) IsValid() bool {
	return k.privateKey.isValid()
}

func (k *KeyPair) Verify(message, sig []byte) bool {
	return k.publicKey.Verify(message, sig)
}

func (k *KeyPair) GetPrivateKey() *PrivateKey {
	return k.privateKey
}

func (k *KeyPair) GetPublicKey() *PublicKey {
	return k.publicKey
}

func (k *KeyPair) Address() *Address {
	return k.GetPublicKey().AsAddress()
}

func (k *KeyPair) SignBytes(data []byte) []byte {
	return k.GetPrivateKey().Sign(data)
}

func (k *KeyPair) Sign(payload []byte) (*Signature, error) {
	return NewSignature(k.GetPublicKey(), k.SignBytes(payload)), nil
}

func (k *KeyPair) SignTransactionBlock(txnBytes []byte, intent suisigner.Intent) (*Signature, error) {
	data := suisigner.MessageWithIntent(intent, txnBytes)
	hash := blake2b.Sum256(data)
	sig := ed25519.Sign(k.privateKey.key, hash[:])

	return NewSignature(k.GetPublicKey(), sig), nil
}
