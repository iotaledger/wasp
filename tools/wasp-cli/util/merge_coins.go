package util

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func MergeAllCoins(ctx context.Context, limit int) {
	client := cliclients.L1Client()
	w := wallet.Load()

	coins, err := client.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{
		Owner: w.Address().AsIotaAddress(),
	})
	log.Check(err)

	baseCoins := lo.Filter(coins.Data, func(item *iotajsonrpc.Coin, index int) bool {
		return coin.BaseTokenType.MatchesStringType(item.CoinType.String())
	})

	// For now a hard coded limit where it would start to make sense to merge the coins again.
	if len(baseCoins) < limit {
		return
	}

	fmt.Println("Doing automatic merge of coin objects..")

	// Merge all coins except the 0th one, to have one as the destination.
	coinsToMerge := make([]*iotago.ObjectRef, len(baseCoins)-1)

	for i := 1; i < len(baseCoins); i++ {
		coinsToMerge[i-1] = baseCoins[i].Ref()
	}

	_, err = client.MergeCoinsAndExecute(ctx, cryptolib.SignerToIotaSigner(w), baseCoins[0].Ref(), coinsToMerge, iotaclient.DefaultGasBudget)
	log.Check(err)
}

func TryMergeAllCoins(ctx context.Context) {
	MergeAllCoins(ctx, 5)
}
