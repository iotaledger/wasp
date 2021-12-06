package cryptolib

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/ed25519"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

type PrivateKey = ed25519.PrivateKey
type PublicKey = ed25519.PublicKey

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

func (k *KeyPair) Valid() bool {
	return len(k.PrivateKey) > 0
}

func (k *KeyPair) Verify(message, sig []byte) bool {
	return Verify(k.PublicKey, message, sig)
}

func (k *KeyPair) Sign(message []byte) ([]byte, error) {
	return k.PrivateKey.Sign(nil, message, nil)
}

func NewKeyPairFromSeed(seed Seed) KeyPair {
	var seedByte [SeedSize]byte = seed

	privateKey := ed25519.NewKeyFromSeed(seedByte[:])

	return NewKeyPairFromPrivateKey(privateKey)
}

func NewKeyPairFromPrivateKey(privateKey PrivateKey) KeyPair {
	publicKey := privateKey.Public().(PublicKey)
	keyPair := KeyPair{privateKey, publicKey}

	return keyPair
}

// NewKeyPair creates a new key pair with a randomly generated seed
func NewKeyPair() KeyPair {
	seed := tpkg.RandEd25519Seed()
	key := NewKeyPairFromSeed(SeedFromByteArray(seed[:]))

	return key
}

func Ed25519AddressFromPubKey(key PublicKey) *iotago.Ed25519Address {
	ret := iotago.Ed25519AddressFromPubKey(key)
	return &ret
}

func PublicKeyFromString(s string) (publicKey PublicKey, err error) {
	b, err := base58.Decode(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from base58 string: %w", s, err)
	}
	publicKey, err = PublicKeyFromBytes(b)
	return publicKey, err
}

func PublicKeyFromBytes(publicKeyBytes []byte) (PublicKey, error) {
	if len(publicKeyBytes) < ed25519.PublicKeySize {
		return nil, fmt.Errorf("bytes too short")
	}

	return publicKeyBytes, nil
}

func PrivateKeyFromBytes(privateKeyBytes []byte) (PrivateKey, error) {
	if len(privateKeyBytes) < ed25519.PrivateKeySize {
		return nil, fmt.Errorf("bytes too short")
	}

	return privateKeyBytes, nil
}

func Verify(publicKey PublicKey, message, sig []byte) bool {
	return ed25519.Verify(publicKey, message, sig)
}

func Sign(privateKey PrivateKey, message []byte) []byte {
	return ed25519.Sign(privateKey, message)
}

//
//func AddressFromKeyPair(keyPair KeyPair) iotago.Address {
//	addr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)
//	return &addr
//}
