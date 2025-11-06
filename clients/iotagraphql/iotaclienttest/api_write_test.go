package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
)

func TestDevInspectTransactionBlock(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	sender := iotatest.MakeSignerWithFunds(0, iotaconn.AlphanetFaucetURL)

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := iotajsonrpc.Coins(coinPages.Data)

	ptb := iotago.NewProgrammableTransactionBuilder()
	ptb.PayAllIota(sender.Address())
	pt := ptb.Finish()
	tx := iotago.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	require.NoError(t, err)

	resp, err := client.DevInspectTransactionBlock(
		context.Background(),
		iotaclient.DevInspectTransactionBlockRequest{
			SenderAddress: sender.Address(),
			TxKindBytes:   txBytes,
			GasPrice:      iotajsonrpc.NewBigInt(iotaclient.DefaultGasPrice),
		},
	)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
}

func TestDryRunTransaction(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	signer := iotago.MustAddressFromHex(testcommon.TestAddress)
	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, big.NewInt(100), iotaclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotaclient.PayAllIotaRequest{
			Signer:     signer,
			Recipient:  signer,
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := client.DryRunTransaction(context.Background(), tx.TxBytes)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}
func TestExecuteTransactionBlock(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.AlphanetFaucetURL)
	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, big.NewInt(100), iotaclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := client.PayAllIota(
		context.Background(),
		iotaclient.PayAllIotaRequest{
			Signer:     signer.Address(),
			Recipient:  signer.Address(),
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	signature, err := signer.SignTransactionBlock(tx.TxBytes, iotasigner.DefaultIntent())
	require.NoError(t, err)

	resp, err := client.ExecuteTransactionBlock(context.Background(), iotaclient.ExecuteTransactionBlockRequest{
		Signatures:  []*iotasigner.Signature{signature},
		TxDataBytes: tx.TxBytes,
	})
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}
