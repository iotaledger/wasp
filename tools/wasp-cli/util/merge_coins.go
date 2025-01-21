package util

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/samber/lo"
)

func TryMergeAllCoins(ctx context.Context) {
	client := cliclients.L1Client()
	w := wallet.Load()

	coins, err := client.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{
		Owner: w.Address().AsIotaAddress(),
	})
	log.Check(err)

	baseCoins := lo.Filter(coins.Data, func(item *iotajsonrpc.Coin, index int) bool {
		if coin.BaseTokenType.MatchesStringType(item.CoinType.String()) {
			return true
		}
		return false
	})

	// For now a hard coded limit where it would start to make sense to merge the coins again.
	if len(baseCoins) < 5 {
		return
	}

	fmt.Println("Doing automatic merge of coin objects..")

	// Merge all coins from the cursor except the first two, to have two coins ready (moving funds and gas)
	coinsToMerge := make([]*iotago.ObjectRef, len(baseCoins)-2)

	for i := 2; i < len(baseCoins); i++ {
		coinsToMerge[i-2] = baseCoins[i].Ref()
	}

	_, err = client.MergeCoinsAndExecute(ctx, cryptolib.SignerToIotaSigner(w), baseCoins[0].Ref(), coinsToMerge, iotaclient.DefaultGasBudget)
	log.Check(err)
}
