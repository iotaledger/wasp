package iotatest

import (
	"context"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
)

func MakeSignerWithFunds(index int, faucetURL, apiURL string) iotasigner.Signer {
	return MakeSignerWithFundsFromSeed(testkey.NewTestSeedBytes(), index, faucetURL, apiURL)
}

func MakeSignerWithFundsFromSeed(seed []byte, index int, faucetURL, apiURL string) iotasigner.Signer {
	keySchemeFlag := iotasigner.KeySchemeFlagDefault

	// there are only 256 different signers can be generated
	signer := iotasigner.NewSignerByIndex(seed, keySchemeFlag, index)

	client := iotagraphql.NewGraphQLClient(apiURL, faucetURL)
	if err := client.RequestFundsFromFaucet(context.Background(), *signer.Address()); err != nil {
		panic(err)
	}
	return signer
}
