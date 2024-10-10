package iotaclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotasigner"
	"github.com/iotaledger/wasp/clients/iota-go/iotatest"
)

func TestMintToken(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.AlphanetFaucetURL)

	// module name is 'testcoin'
	tokenPackageID, treasuryCap := deployTestcoin(t, client, signer)
	mintAmount := uint64(1000000)
	txnRes, err := client.MintToken(
		context.Background(),
		signer,
		tokenPackageID,
		"testcoin",
		treasuryCap,
		mintAmount,
		&iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())

	// all the minted tokens were sent to the signer, so we should find a single object contains all the minted token
	coins, err := client.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner:    signer.Address(),
			CoinType: &coinType,
			Limit:    10,
		},
	)
	require.NoError(t, err)
	require.Equal(t, mintAmount, coins.Data[0].Balance.Uint64())
}

func deployTestcoin(t *testing.T, client *iotaclient.Client, signer iotasigner.Signer) (
	*iotago.PackageID,
	*iotago.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		iotaclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       iotajsonrpc.NewBigInt(iotaclient.DefaultGasBudget * 10),
		},
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &iotajsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnResponse.Effects.Data.IsSuccess())

	packageID, err := txnResponse.GetPublishedPackageID()
	require.NoError(t, err)

	treasuryCap, err := txnResponse.GetCreatedObjectInfo("coin", "TreasuryCap")
	require.NoError(t, err)

	return packageID, treasuryCap.ObjectID
}

func TestBatchGetObjectsOwnedByAddress(t *testing.T) {
	api := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)

	options := iotajsonrpc.SuiObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
	}
	coinType := fmt.Sprintf("0x2::coin::Coin<%v>", iotajsonrpc.IotaCoinType)
	filterObject, err := api.BatchGetObjectsOwnedByAddress(context.TODO(), testAddress, &options, coinType)
	require.NoError(t, err)
	t.Log(filterObject)
}
