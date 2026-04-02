package iotatest

import (
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
)

func MakeSigner(index int) iotasigner.Signer {
	return MakeSignerFromSeed(testkey.NewTestSeedBytes(), index)
}

func MakeSignerFromSeed(seed []byte, index int) iotasigner.Signer {
	keySchemeFlag := iotasigner.KeySchemeFlagDefault
	// there are only 256 different signers can be generated
	signer := iotasigner.NewSignerByIndex(seed, keySchemeFlag, index)
	return signer
}
