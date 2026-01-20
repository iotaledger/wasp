package iotatest

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
)

func MakeSignerWithFunds(index int, faucetURL string, reader ...iotaclient.CoinReader) iotasigner.Signer {
	if index == 0 {
		index = rand.Intn(256)
	}
	return MakeSignerWithFundsFromSeed(testkey.NewTestSeedBytes(), index, faucetURL, reader...)
}

func MakeSignerWithFundsFromSeed(
	seed []byte,
	index int,
	faucetURL string,
	reader ...iotaclient.CoinReader,
) iotasigner.Signer {
	keySchemeFlag := iotasigner.KeySchemeFlagDefault

	// there are only 256 different signers can be generated
	signer := iotasigner.NewSignerByIndex(seed, keySchemeFlag, index)
	var err error
	for i := 0; i < 15; i++ {
		err = iotaclient.RequestFundsFromFaucet(context.Background(), signer.Address(), faucetURL)
		if err == nil {
			break
		}
		if i < 14 && (strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "Too Many Requests")) {
			delay := time.Duration(float64(time.Second) * (10 * float64(i+1)))
			jitter := time.Duration(rand.Float64() * float64(5*time.Second))
			finalDelay := delay + jitter

			fmt.Printf("faucet rate limited (attempt %d/15), sleeping %v before retrying\n", i+1, finalDelay)
			time.Sleep(finalDelay)
			continue
		}
		panic(err)
	}
	if len(reader) > 0 {
		_, err := iotaclient.WaitForCoins(context.Background(), reader[0], signer.Address(), 1, 30*time.Second)
		if err != nil {
			panic(err)
		}
	}
	return signer
}
