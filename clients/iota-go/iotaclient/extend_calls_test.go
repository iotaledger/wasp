package iotaclient_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/contracts"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient/iotaclienttest"
	"github.com/iotaledger/wasp/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iota-go/iotatest"
)

func TestMintToken(t *testing.T) {
	client := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)
	signer := iotatest.MakeSignerWithFunds(0, iotaconn.AlphanetFaucetURL)

	tokenPackageID, treasuryCap := iotaclienttest.DeployCoinPackage(
		t,
		client,
		signer,
		contracts.Testcoin(),
	)
	mintAmount := uint64(1000000)
	_ = iotaclienttest.MintCoins(
		t,
		client,
		signer,
		tokenPackageID,
		contracts.TestcoinModuleName,
		contracts.TestcoinTypeTag,
		treasuryCap.ObjectID,
		mintAmount,
	)
	coinType := fmt.Sprintf("%s::%s::%s", tokenPackageID.String(), contracts.TestcoinModuleName, contracts.TestcoinTypeTag)

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

func TestBatchGetObjectsOwnedByAddress(t *testing.T) {
	api := iotaclient.NewHTTP(iotaconn.AlphanetEndpointURL)

	options := iotajsonrpc.IotaObjectDataOptions{
		ShowType:    true,
		ShowContent: true,
	}
	coinType := fmt.Sprintf("0x2::coin::Coin<%v>", iotajsonrpc.IotaCoinType)
	filterObject, err := api.BatchGetObjectsOwnedByAddress(context.TODO(), testAddress, &options, coinType)
	require.NoError(t, err)
	t.Log(filterObject)
}
