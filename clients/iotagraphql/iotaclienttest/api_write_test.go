package iotaclienttest

import (
	"context"
	"math/big"
	"testing"
	"time"

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

// waitForCoins polls GetCoins until at least one coin is returned or timeout occurs.
func waitForCoins(ctx context.Context, client *clients.GraphQLClient, owner *iotago.Address, limit int, timeout time.Duration) (*iotajsonrpc.CoinPage, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		cp, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{Owner: owner, Limit: limit})
		if err == nil && len(cp.Data) > 0 {
			return cp, nil
		}
		if err != nil {
			lastErr = err
		}
		time.Sleep(1 * time.Second)
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return client.GetCoins(ctx, iotaclient.GetCoinsRequest{Owner: owner, Limit: limit})
}

func TestDevInspectTransactionBlock(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	sender := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)

	limit := int(3)
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
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)

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

	resp, err := client.DryRunTransaction(context.Background(), iotaclient.DryRunTransactionRequest{
		TxDataBytes: tx.TxBytes,
	})
	require.NoError(t, err)
	require.True(t, resp.Effects.Data.IsSuccess())
	require.Empty(t, resp.Effects.Data.V1.Status.Error)
}

func TestExecuteTransactionBlock(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	coins, err := waitForCoins(context.Background(), client, signer.Address(), 10, 30*time.Second)
	require.NoError(t, err)
	require.NotEmpty(t, coins.Data, "no coins indexed for %v after faucet", signer.Address().String())
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

func TestSignAndExecuteTransaction(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.TestnetFaucetURL)
	coins, err := waitForCoins(context.Background(), client, signer.Address(), 10, 30*time.Second)
	require.NoError(t, err)
	require.NotEmpty(t, coins.Data, "no coins indexed for %v after faucet", signer.Address().String())
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

	// Test SignAndExecuteTransaction with options requesting effects and object changes
	// This also tests the isResponseComplete logic to ensure proper handling of incomplete responses
	resp, err := client.SignAndExecuteTransaction(context.Background(), &iotaclient.SignAndExecuteTransactionRequest{
		TxDataBytes: tx.TxBytes,
		Signer:      signer,
		Options: &iotajsonrpc.IotaTransactionBlockResponseOptions{
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
