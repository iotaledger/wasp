package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestDevInspectTransactionBlock(t *testing.T) {
	client := l1starter.Instance().L1Client()
	sender := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

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
		},
	)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
}

func TestDryRunTransaction(t *testing.T) {
	api := l1starter.Instance().L1Client()
	signer := iotago.MustAddressFromHex(testcommon.TestAddress)
	coins, err := api.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, big.NewInt(100), iotaclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := api.PayAllIota(
		context.Background(),
		iotaclient.PayAllIotaRequest{
			Signer:     signer,
			Recipient:  signer,
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := api.DryRunTransaction(context.Background(), tx.TxBytes)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}
