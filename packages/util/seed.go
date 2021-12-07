package util

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"math/rand"
)

type Seed *[cryptolib.SeedSize]byte

func NewSeed_(optionalSeedBytes ...[]byte) Seed {
	seedBytes := [cryptolib.SeedSize]byte{}

	if len(optionalSeedBytes) >= 1 {
		if len(optionalSeedBytes[0]) < cryptolib.SeedSize {
			panic("seed is not long enough")
		}
		copy(seedBytes[:], optionalSeedBytes[0])
		return &seedBytes
	}

	_, err := rand.Read(seedBytes[:])
	if err != nil {
		panic(err)
	}
	return &seedBytes
}

// SubSeed generates the n'th sub seed of this Seed which is then used to generate the KeyPair.

func AddreessFromKey(keyPair cryptolib.KeyPair) iotago.Address {
	addr := cryptolib.Ed25519AddressFromPubKey(keyPair.PublicKey)
	return &addr
}
