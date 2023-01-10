package cryptolib

import (
	"crypto/ed25519"
	"fmt"

	"go.dedis.ch/kyber/v3/sign/eddsa"
	"go.dedis.ch/kyber/v3/util/key"

	iotago "github.com/iotaledger/iota.go/v3"
)

type PrivateKey struct {
	key ed25519.PrivateKey
}

const PrivateKeySize = ed25519.PrivateKeySize

func NewPrivateKey() *PrivateKey {
	return NewPrivateKeyFromSeed(NewSeed())
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

func (pkT *PrivateKey) Clone() *PrivateKey {
	key := make([]byte, len(pkT.key))
	copy(key, pkT.key)
	return &PrivateKey{key: key}
}

func (pkT *PrivateKey) AsBytes() []byte {
	return pkT.key
}

func (pkT *PrivateKey) String() string {
	return iotago.EncodeHex(pkT.key)
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
	return newPublicKeyFromCrypto(pkT.key.Public().(ed25519.PublicKey))
}

func (pkT *PrivateKey) Sign(message []byte) []byte {
	return ed25519.Sign(pkT.key, message)
}

func (pkT *PrivateKey) AddressKeysForEd25519Address(addr *iotago.Ed25519Address) iotago.AddressKeys {
	return iotago.NewAddressKeysForEd25519Address(addr, pkT.key)
}

func (pkT *PrivateKey) AddressKeys(addr iotago.Address) iotago.AddressKeys {
	return iotago.AddressKeys{Address: addr, Keys: pkT.key}
}
