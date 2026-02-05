package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

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
	require.Empty(t, resp.DryRunTransactionBlock.Error)
	require.True(t, resp.DryRunTransactionBlock.Transaction.Effects.IsSuccess())
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
	require.True(t, resp.IsSuccess())
	require.Empty(t, resp.ExecuteTransactionBlock.Errors)
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
	require.NotNil(t, resp.ExecuteTransactionBlock.Effects, "Effects should be present when ShowEffects is true")
	require.NotNil(t, resp.ExecuteTransactionBlock.Effects.ObjectChanges, "ObjectChanges should be present when ShowObjectChanges is true")
	require.True(t, resp.IsSuccess())
	require.Empty(t, resp.ExecuteTransactionBlock.Errors)
}
