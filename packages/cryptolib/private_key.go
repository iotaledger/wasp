package cryptolib

import (
	"crypto"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

const PrivateKeySize = ed25519.PrivateKeySize

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
	return &PrivateKey{ed25519.NewKeyFromSeed(seedByte[:])}
}

func (pkT *PrivateKey) isValid() bool {
	return len(pkT.key) > 0
}

func (pkT *PrivateKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PrivateKey) AsString() string {
	return hex.EncodeToString(pkT.key)
}

func (pkT *PrivateKey) Public() *PublicKey {
	return newPublicKeyFromCrypto(pkT.key.Public().(ed25519.PublicKey))
}

func (pkT *PrivateKey) Sign(rand io.Reader, message []byte, opts crypto.SignerOpts) ([]byte, error) {
	return pkT.key.Sign(rand, message, opts)
}

func (pkT *PrivateKey) AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys {
	return iotago.NewAddressKeysForEd25519Address(addr, pkT.key)
}

func (pkT *PrivateKey) AddressKeys(addr iotago.Address) iotago.AddressKeys {
	return iotago.AddressKeys{Address: addr, Keys: pkT.key}
}
