package cryptolib

import (
	//	cr "crypto"
	crypto "crypto/ed25519"
	"fmt"
	//	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

type (
	PrivateKey = crypto.PrivateKey
	PublicKey  = crypto.PublicKey
)

const (
	PublicKeySize  = crypto.PublicKeySize
	PrivateKeySize = crypto.PrivateKeySize
)

type KeyPair struct {
	PrivateKey PrivateKey
	PublicKey  PublicKey
}

type AddressSigner KeyPair

func NewKeyPairFromSeed(seed Seed) KeyPair {
	privateKey := NewPrivateKeyFromSeed(seed)
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

func (k *KeyPair) Valid() bool {
	return len(k.PrivateKey) > 0
}

func (k *KeyPair) Verify(message, sig []byte) bool {
	return Verify(k.PublicKey, message, sig)
}

func (k *KeyPair) AsAddressSigner() iotago.AddressSigner {
	addrKeys := iotago.NewAddressKeysForEd25519Address(Ed25519AddressFromPubKey(k.PublicKey), k.PrivateKey)
	return iotago.NewInMemoryAddressSigner(addrKeys)
}

//func (k *KeyPair) Sign(message []byte) ([]byte, error) {
//	return k.PrivateKey.Sign(nil, message, nil) // FIXME this is wrong
//}
//
//func (a AddressSigner) Sign(addr iotago.Address, msg []byte) (signature iotago.Signature, err error) {
//	kp := KeyPair(a)
//	if !addr.Equal(Ed25519AddressFromPubKey(kp.PublicKey)) {
//		return nil, fmt.Errorf("can't sign message for given Ed25519 address")
//	}
//	b, err := kp.Sign(msg)
//	ed25519Sig := &iotago.Ed25519Signature{}
//	copy(ed25519Sig.Signature[:], b)
//	copy(ed25519Sig.PublicKey[:], kp.PublicKey)
//	return ed25519Sig, nil
//}

func PrivateKeyFromBytes(privateKeyBytes []byte) (PrivateKey, error) {
	if len(privateKeyBytes) < PrivateKeySize {
		return nil, fmt.Errorf("bytes too short")
	}

	return privateKeyBytes, nil
}

func NewPrivateKeyFromSeed(seed Seed) PrivateKey {
	var seedByte [SeedSize]byte = seed
	return crypto.NewKeyFromSeed(seedByte[:])
}

//TODO
/*func (pkT PrivateKey) asCrypto() crypto.PrivateKey {
	return crypto.PrivateKey(pkT)
}

func (pkT PrivateKey) Sign(rand io.Reader, message []byte, opts cr.SignerOpts) (signature []byte, err error) {
	return pkT.asCrypto().Sign(rand, message, opts)
}*/

func PublicKeyFromString(s string) (publicKey PublicKey, err error) {
	b, err := base58.Decode(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from base58 string: %w", s, err)
	}
	publicKey, err = PublicKeyFromBytes(b)
	return publicKey, err
}

func PublicKeyFromBytes(publicKeyBytes []byte) (PublicKey, error) {
	if len(publicKeyBytes) < PublicKeySize {
		return nil, fmt.Errorf("bytes too short")
	}

	return publicKeyBytes, nil
}

//TODO
/*func (pkT PublicKey) Equal(pk PublicKey) bool {
	return pkT.asCrypto().Equal(pk.asCrypto())
}

func (pkT PublicKey) asCrypto() crypto.PublicKey {
	return crypto.PublicKey(pkT)
}

func (pkT PublicKey) String() string {
	return base58.Encode(pkT)
}*/
