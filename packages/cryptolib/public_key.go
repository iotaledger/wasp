package cryptolib

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

type PublicKey struct {
	key ed25519.PublicKey
}

type PublicKeyKey [PublicKeySize]byte

const PublicKeySize = ed25519.PublicKeySize

var (
	_ rwutil.IoReader = &PublicKey{}
	_ rwutil.IoWriter = &PublicKey{}
)

func publicKeyFromCrypto(cryptoPublicKey ed25519.PublicKey) *PublicKey {
	return &PublicKey{cryptoPublicKey}
}

func NewEmptyPublicKey() *PublicKey {
	return &PublicKey{
		key: make([]byte, PublicKeySize),
	}
}

func PublicKeyFromString(s string) (publicKey *PublicKey, err error) {
	bytes, err := DecodeHex(s)
	if err != nil {
		return publicKey, fmt.Errorf("failed to parse public key %s from hex string: %w", s, err)
	}
	return PublicKeyFromBytes(bytes)
}

func PublicKeyFromBytes(publicKeyBytes []byte) (*PublicKey, error) {
	if len(publicKeyBytes) < PublicKeySize {
		return nil, errors.New("bytes too short")
	}
	return &PublicKey{publicKeyBytes}, nil
}

func (pkT *PublicKey) Clone() *PublicKey {
	key := make([]byte, len(pkT.key))
	copy(key, pkT.key)
	return &PublicKey{key: key}
}

func (pkT *PublicKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PublicKey) AsKey() PublicKeyKey {
	var result [PublicKeySize]byte
	copy(result[:], pkT.key)
	return result
}

func (pkT *PublicKey) AsAddress() *Address {
	return newAddressFromArray(blake2b.Sum256(pkT.key))
}

func (pkT *PublicKey) AsKyberPoint() (kyber.Point, error) {
	return PointFromBytes(pkT.key, new(edwards25519.Curve))
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
	return EncodeHex(pkT.key)
}

func (pkT *PublicKey) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	pkT.key = make([]byte, PublicKeySize)
	rr.ReadN(pkT.key)
	return rr.Err
}

func (pkT *PublicKey) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	if len(pkT.key) != PublicKeySize {
		panic("unexpected public key size for write")
	}
	ww.WriteN(pkT.key)
	return ww.Err
}

func (pkT *PublicKey) Bytes() []byte {
	return rwutil.WriteToBytes(pkT)
}
