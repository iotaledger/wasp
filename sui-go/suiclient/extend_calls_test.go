package suiclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/sui-go/contracts"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
	"github.com/iotaledger/wasp/sui-go/suisigner"
	"github.com/iotaledger/wasp/sui-go/suitest"
)

func TestMintToken(t *testing.T) {
	client := suiclient.NewHTTP(suiconn.TestnetEndpointURL)
	signer := suitest.MakeSignerWithFunds(0, suiconn.TestnetFaucetURL)

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
		&suijsonrpc.SuiTransactionBlockResponseOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	)
	require.NoError(t, err)
	require.True(t, txnRes.Effects.Data.IsSuccess())
	coinType := fmt.Sprintf("%s::testcoin::TESTCOIN", tokenPackageID.String())

	// all the minted tokens were sent to the signer, so we should find a single object contains all the minted token
	coins, err := client.GetCoins(context.Background(), suiclient.GetCoinsRequest{
		Owner:    signer.Address(),
		CoinType: &coinType,
		Limit:    10,
	})
	require.NoError(t, err)
	require.Equal(t, mintAmount, coins.Data[0].Balance.Uint64())
}

func deployTestcoin(t *testing.T, client *suiclient.Client, signer suisigner.Signer) (
	*sui.PackageID,
	*sui.ObjectID,
) {
	testcoinBytecode := contracts.Testcoin()

	txnBytes, err := client.Publish(
		context.Background(),
		suiclient.PublishRequest{
			Sender:          signer.Address(),
			CompiledModules: testcoinBytecode.Modules,
			Dependencies:    testcoinBytecode.Dependencies,
			GasBudget:       suijsonrpc.NewBigInt(suiclient.DefaultGasBudget * 10),
		},
	)
	require.NoError(t, err)
	txnResponse, err := client.SignAndExecuteTransaction(
		context.Background(), signer, txnBytes.TxBytes, &suijsonrpc.SuiTransactionBlockResponseOptions{
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
	api := suiclient.NewHTTP(suiconn.DevnetEndpointURL)

	options := suijsonrpc.SuiObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
	}
	coinType := fmt.Sprintf("0x2::coin::Coin<%v>", suijsonrpc.SuiCoinType)
	filterObject, err := api.BatchGetObjectsOwnedByAddress(context.TODO(), testAddress, &options, coinType)
	require.NoError(t, err)
	t.Log(filterObject)
}
