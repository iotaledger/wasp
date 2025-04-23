package util

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/bcs-go"
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

func TryManageCoinsAmount(ctx context.Context) {
	client := cliclients.L1Client()
	w := wallet.Load()

	coinPage, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner: w.Address().AsIotaAddress(),
	})
	log.Check(err)

	coins := iotajsonrpc.Coins(coinPage.Data)
	var mergeCoins []iotago.Argument
	sum := uint64(0)
	ptb := iotago.NewProgrammableTransactionBuilder()

	for i, coin := range coins {
		sum += coin.Balance.Uint64()
		if i == 0 {
			continue
		}
		mergeCoins = append(mergeCoins, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: coin.Ref()}))
	}

	if len(coins) > 1 {
		ptb.Command(iotago.Command{MergeCoins: &iotago.ProgrammableMergeCoins{
			Destination: iotago.GetArgumentGasCoin(),
			Sources:     mergeCoins,
		}})
	}

	argSplitAmount := ptb.MustForceSeparatePure(sum/5 - 100)
	argSplitCoins := ptb.Command(iotago.Command{SplitCoins: &iotago.ProgrammableSplitCoins{
		Coin:    iotago.GetArgumentGasCoin(),
		Amounts: []iotago.Argument{argSplitAmount, argSplitAmount, argSplitAmount, argSplitAmount},
	}})
	ptb.Command(iotago.Command{TransferObjects: &iotago.ProgrammableTransferObjects{
		Objects: []iotago.Argument{
			{NestedResult: &iotago.NestedResult{Cmd: *argSplitCoins.Result, Result: uint16(0)}},
			{NestedResult: &iotago.NestedResult{Cmd: *argSplitCoins.Result, Result: uint16(1)}},
			{NestedResult: &iotago.NestedResult{Cmd: *argSplitCoins.Result, Result: uint16(2)}},
			{NestedResult: &iotago.NestedResult{Cmd: *argSplitCoins.Result, Result: uint16(3)}},
		},
		Address: ptb.MustPure(w.Address().AsIotaAddress()),
	}})
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		w.Address().AsIotaAddress(),
		pt,
		[]*iotago.ObjectRef{coins[0].Ref()},
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	log.Check(err)
	_, err = client.SignAndExecuteTransaction(
		ctx,
		&iotaclient.SignAndExecuteTransactionRequest{
			Signer:      cryptolib.SignerToIotaSigner(w),
			TxDataBytes: txBytes,
			Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
				ShowEffects:       true,
				ShowObjectChanges: true,
			},
		},
	)
	log.Check(err)
}
