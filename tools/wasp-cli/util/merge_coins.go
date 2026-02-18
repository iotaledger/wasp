package util

import (
	"context"
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func TryMergeAllCoins(ctx context.Context) error {
	client := cliclients.L1Client()
	w := wallet.Load()

	coins, err := client.GetAllCoins(ctx, iotagraphql.GetAllCoinsRequest{
		Owner: *w.Address().AsIotaAddress(),
	})
	if err != nil {
		return err
	}

	baseCoins := lo.Filter(coins.Address.Coins.Nodes, func(item iotagraphql.Coin, index int) bool {
		return coin.BaseTokenType.MatchesStringType(item.CoinType().String())
	})

	// For now a hard coded limit where it would start to make sense to merge the coins again.
	if len(baseCoins) < 5 {
		return nil
	}

	fmt.Println("Doing automatic merge of coin objects..")

	// Merge all coins from the cursor except the first two, to have two coins ready (moving funds and gas)
	coinsToMerge := make([]*iotago.ObjectRef, len(baseCoins)-2)

	for i := 2; i < len(baseCoins); i++ {
		ref, err := baseCoins[i].ObjectRef()
		if err != nil {
			return err
		}
		coinsToMerge[i-2] = ref
	}

	destRef, err := baseCoins[0].ObjectRef()
	if err != nil {
		return err
	}
	_, err = mergeCoinsAndExecute(ctx, client, cryptolib.SignerToIotaSigner(w), destRef, coinsToMerge, iotagraphql.DefaultGasBudget)
	if err != nil {
		return err
	}
	return nil
}

func TryManageCoinsAmount(ctx context.Context) {
	client := cliclients.L1Client()
	w := wallet.Load()

	coinPage, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: *w.Address().AsIotaAddress(),
	})
	log.Check(err)

	coins := iotagraphql.Coins(coinPage.Address.Coins.Nodes)
	var mergeCoins []iotago.Argument
	sum := uint64(0)
	ptb := iotago.NewProgrammableTransactionBuilder()

	for i := range coins {
		sum += coins[i].Balance()
		if i == 0 {
			continue
		}
		ref, err := coins[i].ObjectRef()
		log.Check(err)
		mergeCoins = append(mergeCoins, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: ref}))
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
	gasRef, err := coins[0].ObjectRef()
	log.Check(err)
	tx := iotago.NewProgrammable(
		w.Address().AsIotaAddress(),
		pt,
		[]*iotago.ObjectRef{gasRef},
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)

	txBytes, err := bcs.Marshal(&tx)
	log.Check(err)
	_, err = client.SignAndExecuteTransaction(
		ctx,
		txBytes,
		cryptolib.SignerToIotaSigner(w),
	)
	log.Check(err)
}

func mergeCoinsAndExecute(
	ctx context.Context,
	client clients.L1Client,
	owner iotasigner.Signer,
	destinationCoin *iotago.ObjectRef,
	sourceCoins []*iotago.ObjectRef,
	gasBudget uint64,
) (*iotagraphql.ExecuteTransactionBlockResponse, error) {
	ptb := iotago.NewProgrammableTransactionBuilder()
	var argCoins []iotago.Argument
	for _, sourceCoin := range sourceCoins {
		argCoins = append(argCoins, ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: sourceCoin}))
	}
	ptb.Command(
		iotago.Command{
			MergeCoins: &iotago.ProgrammableMergeCoins{
				Destination: ptb.MustObj(iotago.ObjectArg{ImmOrOwnedObject: destinationCoin}),
				Sources:     argCoins,
			},
		},
	)
	pt := ptb.Finish()

	gasCoins, err := client.GetCoinObjsForTargetAmount(ctx, *owner.Address(), iotagraphql.DefaultGasPrice, gasBudget)
	if err != nil {
		return nil, fmt.Errorf("failed to find gas payment: %w", err)
	}
	gasCoins, err = iotagraphql.PickupCoinsWithFilter(
		gasCoins,
		gasBudget,
		func(c iotagraphql.Coin) bool {
			addr := c.ObjectID()
			return !pt.IsInInputObjects(&addr)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to find gas payment: %w", err)
	}

	coinRefs, err := gasCoins.CoinRefs()
	if err != nil {
		return nil, fmt.Errorf("failed to get coin refs: %w", err)
	}

	tx := iotago.NewProgrammable(
		owner.Address(),
		pt,
		coinRefs,
		gasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx)
	if err != nil {
		return nil, fmt.Errorf("can't marshal transaction into BCS encoding: %w", err)
	}
	txnResponse, err := client.SignAndExecuteTransaction(
		ctx,
		txBytes,
		owner,
	)
	if err != nil {
		return nil, fmt.Errorf("can't execute the transaction: %w", err)
	}

	return txnResponse, nil
}
