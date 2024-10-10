package suiclient_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/suitest"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestDevInspectTransactionBlock(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	sender := suitest.MakeSignerWithFunds(0, suiconn.AlphanetFaucetURL)

	limit := uint(3)
	coinPages, err := client.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: sender.Address(),
			Limit: limit,
		},
	)
	require.NoError(t, err)
	coins := suijsonrpc.Coins(coinPages.Data)

	ptb := sui.NewProgrammableTransactionBuilder()
	ptb.PayAllSui(sender.Address())
	pt := ptb.Finish()
	tx := sui.NewProgrammable(
		sender.Address(),
		pt,
		coins.CoinRefs(),
		suiclient.DefaultGasBudget,
		suiclient.DefaultGasPrice,
	)
	txBytes, err := bcs.Marshal(&tx.V1.Kind)
	require.NoError(t, err)

	resp, err := client.DevInspectTransactionBlock(
		context.Background(),
		suiclient.DevInspectTransactionBlockRequest{
			SenderAddress: sender.Address(),
			TxKindBytes:   txBytes,
		},
	)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
}

func TestDryRunTransaction(t *testing.T) {
	api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	signer := testAddress
	coins, err := api.GetCoins(
		context.Background(), suiclient.GetCoinsRequest{
			Owner: signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := suijsonrpc.PickupCoins(coins, big.NewInt(100), suiclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)
	tx, err := api.PayAllSui(
		context.Background(),
		suiclient.PayAllSuiRequest{
			Signer:     signer,
			Recipient:  signer,
			InputCoins: pickedCoins.CoinIds(),
			GasBudget:  suijsonrpc.NewBigInt(suiclient.DefaultGasBudget),
		},
	)
	require.NoError(t, err)

	resp, err := api.DryRunTransaction(context.Background(), tx.TxBytes)
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}
