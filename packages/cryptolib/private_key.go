package cryptolib

import (
	crypto "crypto/ed25519"
	"fmt"
)

type PrivateKey struct {
	key crypto.PrivateKey
}

const PrivateKeySize = crypto.PrivateKeySize

func NewPrivateKeyFromBytes(privateKeyBytes []byte) (PrivateKey, error) {
	if len(privateKeyBytes) < PrivateKeySize {
		return PrivateKey{}, fmt.Errorf("bytes too short")
	}
	return PrivateKey{privateKeyBytes}, nil
}

func NewPrivateKeyFromSeed(seed Seed) PrivateKey {
	var seedByte [SeedSize]byte = seed
	return PrivateKey{crypto.NewKeyFromSeed(seedByte[:])}
}

func (pkT PrivateKey) asCrypto() crypto.PrivateKey {
	return pkT.key
}

func (pkT PrivateKey) Public() PublicKey {
	return newPublicKeyFromCrypto(pkT.key.Public().(crypto.PublicKey))
}
