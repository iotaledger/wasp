package iotatest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
)

func EnsureCoinCount(t *testing.T, cryptolibSigner iotasigner.Signer, client clients.L1Client, coinCount int) {
	ctx := context.Background()

	getCoinsRes, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{Owner: cryptolibSigner.Address()})
	require.NoError(t, err)

	have := len(getCoinsRes.Address.Coins.Nodes)
	if have >= coinCount {
		return
	}

	existingCoins := iotagraphql.Coins(getCoinsRes.Address.Coins.Nodes)
	totalBalance := existingCoins.TotalBalance().Uint64()
	distributable := totalBalance - iotagraphql.DefaultGasBudget
	splitAmount := distributable / uint64(coinCount)

	txb := iotago.NewProgrammableTransactionBuilder()

	amounts := make([]iotago.Argument, coinCount-1)
	for i := range amounts {
		amounts[i] = txb.MustPure(splitAmount)
	}
	splitCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: amounts,
			},
		},
	)

	splitResults := make([]iotago.Argument, coinCount-1)
	for i := range splitResults {
		splitResults[i] = iotago.Argument{NestedResult: &iotago.NestedResult{Cmd: *splitCmd.Result, Result: uint16(i)}}
	}
	addr := cryptolibSigner.Address()
	txb.TransferArgs(&addr, splitResults)

	gasPayments, err := existingCoins.CoinRefs()
	require.NoError(t, err)

	txData := iotago.NewProgrammable(
		&addr,
		txb.Finish(),
		gasPayments,
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	result, err := client.SignAndExecuteTransaction(ctx, txnBytes, cryptolibSigner)
	require.NoError(t, err)
	require.True(t, result.IsSuccess(), "EnsureCoinCount tx failed: %s", result.ExecuteTransactionBlock.Effects.Errors)
}

func EnsureCoinSplitWithBalance(
	t *testing.T,
	cryptolibSigner iotasigner.Signer,
	client clients.L1Client,
	splitBalance uint64,
) {
	getCoinsRes, err := client.GetCoins(
		context.Background(),
		iotagraphql.GetCoinsRequest{Owner: cryptolibSigner.Address()},
	)
	require.NoError(t, err)

	if len(getCoinsRes.Address.Coins.Nodes) > 1 {
		return
	}

	coins, err := client.GetCoinObjsForTargetAmount(
		context.Background(),
		cryptolibSigner.Address(),
		splitBalance,
		iotagraphql.DefaultGasBudget,
	)
	require.NoError(t, err)

	txb := iotago.NewProgrammableTransactionBuilder()

	splitCmd := txb.Command(
		iotago.Command{
			SplitCoins: &iotago.ProgrammableSplitCoins{
				Coin:    iotago.GetArgumentGasCoin(),
				Amounts: []iotago.Argument{txb.MustPure(splitBalance)},
			},
		},
	)
	addr2 := cryptolibSigner.Address()
	txb.TransferArg(&addr2, splitCmd)

	coinRef, err := coins[0].ObjectRef()
	require.NoError(t, err)

	txData := iotago.NewProgrammable(
		&addr2,
		txb.Finish(),
		[]*iotago.ObjectRef{coinRef},
		iotagraphql.DefaultGasBudget,
		parameterstest.L1Mock.Protocol.ReferenceGasPrice.Uint64(),
	)

	txnBytes, err := bcs.Marshal(&txData)
	require.NoError(t, err)

	result, err := client.SignAndExecuteTransaction(
		context.Background(),
		txnBytes,
		cryptolibSigner,
	)
	require.NoError(t, err)
	require.True(
		t,
		result.IsSuccess(),
		"EnsureCoinSplitWithBalance tx failed: %s",
		result.ExecuteTransactionBlock.Effects.Errors,
	)
}
