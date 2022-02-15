package cryptolib

import (
	cr "crypto"
	crypto "crypto/ed25519"
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

type PrivateKey struct {
	key crypto.PrivateKey
}

const PrivateKeySize = crypto.PrivateKeySize

func NewPrivateKey() *PrivateKey {
	seed := tpkg.RandEd25519Seed()
	return NewPrivateKeyFromSeed(seed)
}

func NewPrivateKeyFromBytes(privateKeyBytes []byte) (*PrivateKey, error) {
	if len(privateKeyBytes) < PrivateKeySize {
		return nil, fmt.Errorf("bytes too short")
	}
	return &PrivateKey{privateKeyBytes}, nil
}

func NewPrivateKeyFromSeed(seed Seed) *PrivateKey {
	var seedByte [SeedSize]byte = seed
	return &PrivateKey{crypto.NewKeyFromSeed(seedByte[:])}
}

func (pkT *PrivateKey) isValid() bool {
	return len(pkT.key) > 0
}

func (pkT *PrivateKey) AsCrypto() crypto.PrivateKey {
	return pkT.key
}

func (pkT *PrivateKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PrivateKey) Public() *PublicKey {
	return newPublicKeyFromCrypto(pkT.key.Public().(crypto.PublicKey))
}

func (pkT *PrivateKey) Sign(rand io.Reader, message []byte, opts cr.SignerOpts) ([]byte, error) {
	return pkT.key.Sign(rand, message, opts)
}

func (pkT *PrivateKey) AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys {
	return iotago.NewAddressKeysForEd25519Address(addr, pkT.key)
}
