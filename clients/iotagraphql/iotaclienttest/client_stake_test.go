package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotatest"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestRequestAddDelegation(t *testing.T) {
	if l1starter.Instance().IsLocal() {
		t.Skipf("Skipped test as the configured local node does not support this test case")
	}

	client := l1starter.Instance().L1Client()
	signer := iotatest.MakeSignerWithFunds(0, l1starter.Instance().FaucetURL())

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: signer.Address(),
			Limit: 10,
		},
	)
	require.NoError(t, err)

	amount := uint64(iotago.UnitIota)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int).SetUint64(amount), 0, 0, 0)
	require.NoError(t, err)

	validator, err := GetValidatorAddress(context.Background())
	require.NoError(t, err)

	txBytes, err := iotaclient.BCS_RequestAddStake(
		signer.Address(),
		pickedCoins.CoinRefs(),
		iotajsonrpc.NewBigInt(amount),
		&validator,
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
	if l1starter.Instance().IsLocal() {
		t.Skipf("Skipped test as the configured local node does not support this test case")
	}

	client := l1starter.Instance().L1Client()
	signer, err := GetValidatorAddressWithCoins(context.Background())
	require.NoError(t, err)
	stakes, err := client.GetStakes(context.Background(), &signer)
	require.NoError(t, err)
	require.True(t, len(stakes) > 0)
	require.True(t, len(stakes[0].Stakes) > 0)

	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner: &signer,
			Limit: 10,
		},
	)
	require.NoError(t, err)
	pickedCoins, err := iotajsonrpc.PickupCoins(coins, new(big.Int), iotaclient.DefaultGasBudget, 0, 0)
	require.NoError(t, err)

	detail, err := client.GetObject(
		context.Background(), iotaclient.GetObjectRequest{
			ObjectID: &stakes[0].Stakes[0].Data.StakedIotaId,
		},
	)
	require.NoError(t, err)
	txBytes, err := iotaclient.BCS_RequestWithdrawStake(
		&signer,
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
