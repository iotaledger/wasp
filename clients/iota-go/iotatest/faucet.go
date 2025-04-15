package iotatest

import (
	"context"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
)

func MakeSignerWithFunds(index int, faucetURL string) iotasigner.Signer {
	return MakeSignerWithFundsFromSeed(testkey.NewTestSeedBytes(), index, faucetURL)
}

func MakeSignerWithFundsFromSeed(seed []byte, index int, faucetURL string) iotasigner.Signer {
	keySchemeFlag := iotasigner.KeySchemeFlagDefault

	// there are only 256 different signers can be generated
	signer := iotasigner.NewSignerByIndex(seed, keySchemeFlag, index)
	err := iotaclient.RequestFundsFromFaucet(context.Background(), signer.Address(), faucetURL)
	if err != nil {
		panic(err)
	}
	return signer
}
