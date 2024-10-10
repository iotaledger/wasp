package iotaclient_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotatest"
)

const (
	ComingChatValidatorAddress = "0x520289e77c838bae8501ae92b151b99a54407288fdd20dee6e5416bfe943eb7a"
)

func TestRequestAddDelegation(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.AlphanetFaucetURL)

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)

	amount := uint64(iotago.UnitSui)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int).SetUint64(amount), 0, 0, 0)
	require.NoError(t, err)

	validatorAddress := ComingChatValidatorAddress
	validator, err := iotago.AddressFromHex(validatorAddress)
	require.NoError(t, err)

	txBytes, err := iotaclient.BCS_RequestAddStake(
		signer.Address(),
		pickedCoins.CoinRefs(),
		iotajsonrpc.NewBigInt(amount),
		validator,
		iotaclient.DefaultGasBudget,
		iotaclient.DefaultGasPrice,
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Equal(t, "", simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
}

func TestRequestWithdrawDelegation(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)

	signer, err := iotago.AddressFromHex("0xd77955e670f42c1bc5e94b9e68e5fe9bdbed9134d784f2a14dfe5fc1b24b5d9f")
	require.NoError(t, err)
	stakes, err := client.GetStakes(context.Background(), signer)
	require.NoError(t, err)
	require.True(t, len(stakes) > 0)
	require.True(t, len(stakes[0].Stakes) > 0)

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int), iotaclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)

	detail, err := client.GetObject(
		context.Background(), iotaclient.GetObjectRequest{
			ObjectID: &stakes[0].Stakes[0].Data.StakedSuiId,
		},
	)
	require.NoError(t, err)
	txBytes, err := iotaclient.BCS_RequestWithdrawStake(
		signer,
		detail.Data.Ref(),
		pickedCoins.CoinRefs(),
		iotaclient.DefaultGasBudget,
		1000,
	)
	require.NoError(t, err)

	simulate, err := client.DryRunTransaction(context.Background(), txBytes)
	require.NoError(t, err)
	require.Equal(t, "", simulate.Effects.Data.V1.Status.Error)
	require.True(t, simulate.Effects.Data.IsSuccess())
}
