package cryptolib

import (
	"crypto/ed25519"
	"encoding/hex"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

type PublicKey struct {
	key ed25519.PublicKey
}

type PublicKeyKey [PublicKeySize]byte

const PublicKeySize = ed25519.PublicKeySize

func newPublicKeyFromCrypto(cryptoPublicKey ed25519.PublicKey) *PublicKey {
	return &PublicKey{cryptoPublicKey}
}

func NewEmptyPublicKey() *PublicKey {
	return &PublicKey{
		key: make([]byte, PublicKeySize),
	}
}

func NewPublicKeyFromString(s string) (publicKey *PublicKey, err error) {
	b, err := base58.Decode(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from base58 string: %w", s, err)
	}
	publicKey, err = NewPublicKeyFromBytes(b)
	return publicKey, err
}

func NewPublicKeyFromBytes(publicKeyBytes []byte) (*PublicKey, error) {
	if len(publicKeyBytes) < PublicKeySize {
		return nil, xerrors.Errorf("bytes too short")
	}
	return &PublicKey{publicKeyBytes}, nil
}

func (pkT *PublicKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PublicKey) AsByteArray() [PublicKeySize]byte {
	var result [PublicKeySize]byte
	copy(result[:], pkT.key)
	return result
}

func (pkT *PublicKey) AsKey() PublicKeyKey {
	return PublicKeyKey(pkT.AsByteArray())
}

func (pkT *PublicKey) AsString() string {
	return hex.EncodeToString(pkT.key)
}

func (pkT *PublicKey) AsEd25519Address() *iotago.Ed25519Address {
	ret := iotago.Ed25519AddressFromPubKey(pkT.key)
	return &ret
}

func (pkT *PublicKey) Equals(other *PublicKey) bool {
	if len(pkT.key) != len(other.key) {
		return false
	}
	for i := range pkT.key {
		if pkT.key[i] != other.key[i] {
			return false
		}
	}
	return true
}

func (pkT *PublicKey) Verify(message, sig []byte) bool {
	return ed25519.Verify(pkT.key, message, sig)
}

func (pkT *PublicKey) String() string {
	return pkT.AsString()
}
