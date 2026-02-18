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
			Owner: *signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(iotagraphql.Coins(coins.Address.Coins.Nodes), big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     *signer,
			Recipient:  *signer,
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := client.DryRunTransaction(context.Background(), tx.TxBytes)
	require.NoError(t, err)
	require.True(t, resp.DryRunTransactionBlock.Transaction.Effects.IsSuccess())
}

func TestExecuteTransactionBlock(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := l1starter.ISCPackageOwner
	coins, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: *signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(iotagraphql.Coins(coins.Address.Coins.Nodes), big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     *signer.Address(),
			Recipient:  *signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	signature, err := signer.SignTransactionBlock(tx.TxBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)

	resp, err := client.ExecuteTransactionBlock(context.Background(), tx.TxBytes, []*iotasigner.Signature{signature})
	require.NoError(t, err)
	require.True(t, resp.IsSuccess())
	require.Empty(t, resp.ExecuteTransactionBlock.Errors)
}

func TestSignAndExecuteTransaction(t *testing.T) {
	client := l1starter.Instance().L1Client()
	signer := l1starter.ISCPackageOwner

	coins, err := client.GetCoins(
		context.Background(), iotagraphql.GetCoinsRequest{
			Owner: *signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotagraphql.PickupCoins(iotagraphql.Coins(coins.Address.Coins.Nodes), big.NewInt(100), iotagraphql.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotagraphql.PayAllIotaRequest{
			Signer:     *signer.Address(),
			Recipient:  *signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotagraphql.NewBigInt(iotagraphql.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := client.SignAndExecuteTransaction(context.Background(), tx.TxBytes, signer)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ExecuteTransactionBlock.Effects, "Effects should be present")
	require.NotNil(t, resp.ExecuteTransactionBlock.Effects.ObjectChanges, "ObjectChanges should be present")
	require.True(t, resp.IsSuccess())
	require.Empty(t, resp.ExecuteTransactionBlock.Errors)
}
