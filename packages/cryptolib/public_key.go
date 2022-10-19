package cryptolib

import (
	"crypto/ed25519"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/xerrors"

	iotago "github.com/iotaledger/iota.go/v3"
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

func NewPublicKeyFromHexString(s string) (publicKey *PublicKey, err error) {
	bytes, err := hexutil.Decode(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from hex string: %w", s, err)
	}
	publicKey, err = NewPublicKeyFromBytes(bytes)
	return publicKey, err
}

func NewPublicKeyFromString(s string) (publicKey *PublicKey, err error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return publicKey, xerrors.Errorf("failed to parse public key %s from hex string: %w", s, err)
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

func (pkT *PublicKey) AsKey() PublicKeyKey {
	var result [PublicKeySize]byte
	copy(result[:], pkT.key)
	return result
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
	return hex.EncodeToString(pkT.key)
}
