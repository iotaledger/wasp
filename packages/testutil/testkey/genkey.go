// Package testkey provides utilities for generating and managing test keys
package testkey

import (
	testcommon "github.com/iotaledger/wasp/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/testutil/testconfig"
)

func GenKeyAddr(seedOpt ...*cryptolib.Seed) (*cryptolib.KeyPair, *cryptolib.Address) {
	var keyPair *cryptolib.KeyPair
	if len(seedOpt) > 0 {
		keyPair = cryptolib.KeyPairFromSeed(*seedOpt[0])
	} else {
		keyPair = cryptolib.NewKeyPair()
	}
	addr := keyPair.GetPublicKey().AsAddress()
	return keyPair, addr
}

func UseRandomSeed() bool {
	const useRandomSeedByDefault = true
	return testconfig.Get("testing", "USE_RANDOM_SEED", useRandomSeedByDefault)
}

func NewTestSeed() cryptolib.Seed {
	if !UseRandomSeed() {
		return cryptolib.Seed(testcommon.TestSeed)
	}

	return cryptolib.NewSeed()
}

func NewTestSeedBytes() []byte {
	seed := NewTestSeed()
	return seed[:]
}
