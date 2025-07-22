package cryptolib

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"

	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/util/key"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

const PrivateKeySize = ed25519.PrivateKeySize

var (
	_ rwutil.IoReader = &PrivateKey{}
	_ rwutil.IoWriter = &PrivateKey{}
)

func NewPrivateKey() *PrivateKey {
	return PrivateKeyFromSeed(NewSeed())
}

func PrivateKeyFromBytes(privateKeyBytes []byte) (*PrivateKey, error) {
	if len(privateKeyBytes) < PrivateKeySize {
		return nil, errors.New("bytes too short")
	}
	return &PrivateKey{privateKeyBytes}, nil
}

func PrivateKeyFromSeed(seed Seed) *PrivateKey {
	var seedByte [SeedSize]byte = seed
	return &PrivateKey{ed25519.NewKeyFromSeed(seedByte[:])}
}

func (pkT *PrivateKey) isValid() bool {
	return len(pkT.key) > 0
}

func (pkT *PrivateKey) Clone() *PrivateKey {
	key := make([]byte, len(pkT.key))
	copy(key, pkT.key)
	return &PrivateKey{key: key}
}

func (pkT *PrivateKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PrivateKey) String() string {
	return EncodeHex(pkT.key)
}

func (pkT *PrivateKey) AsStdKey() ed25519.PrivateKey {
	return pkT.key
}

func (pkT *PrivateKey) AsKyberKeyPair() (*key.Pair, error) {
	keyPair := eddsa.EdDSA{}
	if err := keyPair.UnmarshalBinary(pkT.AsBytes()); err != nil {
		return nil, fmt.Errorf("cannot convert node priv key to kyber: %w", err)
	}
	return &key.Pair{Public: keyPair.Public, Private: keyPair.Secret}, nil
}

func (pkT *PrivateKey) Public() *PublicKey {
	return publicKeyFromCrypto(pkT.key.Public().(ed25519.PublicKey))
}

func (pkT *PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(pkT.key, message)
}

func (pkT *PrivateKey) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	pkT.key = make([]byte, PrivateKeySize)
	rr.ReadN(pkT.key)
	return rr.Err
}

func (pkT *PrivateKey) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	if len(pkT.key) != PrivateKeySize {
		panic("unexpected private key size for write")
	}
	ww.WriteN(pkT.key)
	return ww.Err
}
