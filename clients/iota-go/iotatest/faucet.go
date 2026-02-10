package iotatest

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
)

func MakeSignerWithFunds(index int, faucetURL string) iotasigner.Signer {
	return MakeSignerWithFundsFromSeed(testkey.NewTestSeedBytes(), index, faucetURL, "")
}

func MakeSignerWithFundsAndWait(index int, faucetURL, apiURL string) iotasigner.Signer {
	return MakeSignerWithFundsFromSeed(testkey.NewTestSeedBytes(), index, faucetURL, apiURL)
}

func MakeSignerWithFundsFromSeed(seed []byte, index int, faucetURL, apiURL string) iotasigner.Signer {
	keySchemeFlag := iotasigner.KeySchemeFlagDefault

	// there are only 256 different signers can be generated
	signer := iotasigner.NewSignerByIndex(seed, keySchemeFlag, index)

	var err error
	if apiURL != "" {
		// Wait for coins to be visible on the ledger
		err = iotagraphql.RequestFundsFromFaucetAndWait(context.Background(), signer.Address(), faucetURL, apiURL)
	} else {
		// Non-blocking: returns immediately after faucet request succeeds
		err = iotagraphql.RequestFundsFromFaucet(context.Background(), signer.Address(), faucetURL)
	}
	if err != nil {
		panic(err)
	}
	return signer
}
