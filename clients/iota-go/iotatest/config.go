package iotatest

import "github.com/iotaledger/wasp/packages/testutil/testconfig"

func UseRandomSeed() bool {
	const useRandomSeedByDefault = true
	return testconfig.Get("testing", "USE_RANDOM_SEED", useRandomSeedByDefault)
}
