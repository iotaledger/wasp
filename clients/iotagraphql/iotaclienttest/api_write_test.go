package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestDevInspectTransactionBlock(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	limit := int(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotagraphql.Coins(coinPages.Data)

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.PayAllIota(sender.Address())
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		iotagraphql.DefaultGasBudget,
		iotagraphql.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	require.NoError(t, err)

	resp, err := client.DevInspectTransactionBlock(
		context.Background(),
		iotagraphql.DevInspectTransactionBlockRequest{
			SenderAddress: sender.Address(),
			TxKindBytes:   txBytes,
			GasPrice:      iotagraphql.NewBigInt(iotagraphql.DefaultGasPrice),
		},
	)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
}

func TestDryRunTransaction(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := l1starter.ISCPackageOwner.Address()

	coins, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(coins, big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     signer,
			Recipient:  signer,
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := client.DryRunTransaction(context.Background(), iotagraphql.DryRunTransactionRequest{
		TxDataBytes: tx.TxBytes,
	})
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}

func TestExecuteTransactionBlock(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := l1starter.ISCPackageOwner
	coins, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(coins, big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     signer.Address(),
			Recipient:  signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	signature, err := signer.SignTransactionBlock(tx.TxBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)

	resp, err := client.ExecuteTransactionBlock(context.Background(), iotagraphql.ExecuteTransactionBlockRequest{
		Signatures:  []*iotasigner.Signature{signature},
		TxDataBytes: tx.TxBytes,
		Options: &iotagraphql.IotaTransactionBlockResponseOptions{
			ShowEffects: true,
		},
	})
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}

func TestSignAndExecuteTransaction(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := l1starter.ISCPackageOwner

	coins, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(coins, big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     signer.Address(),
			Recipient:  signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	// Test SignAndExecuteTransaction with options requesting effects and object changes
	// This also tests the isResponseComplete logic to ensure proper handling of incomplete responses
	resp, err := client.SignAndExecuteTransaction(context.Background(), &iotagraphql.SignAndExecuteTransactionRequest{
		TxDataBytes: tx.TxBytes,
		Signer:      signer,
		Options: &iotagraphql.IotaTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.Effects, "Effects should be present when ShowEffects is true")
	require.NotNil(t, resp.ObjectChanges, "ObjectChanges should be present when ShowObjectChanges is true")
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}
