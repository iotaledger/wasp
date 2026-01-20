package iotaclienttest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
)

func TestGetObject(t *testing.T) {
	ctx := context.Background()
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := int(1)
	var coinsResp *iotajsonrpc.CoinPage
	var err error
	require.Eventually(t, func() bool {
		coinsResp, err = client.GetCoins(ctx, iotaclient.GetCoinsRequest{
			Owner: owner,
			Limit: limit,
		})
		return err == nil && len(coinsResp.Data) > 0
	}, 120*time.Second, 5*time.Second)

	coin := coinsResp.Data[0]
	objResp, err := client.GetObject(ctx, iotaclient.GetObjectRequest{
		ObjectID: coin.CoinObjectID,
		Options: &iotajsonrpc.IotaObjectDataOptions{
			ShowContent: true,
			ShowType:    true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, objResp)
}

func TestGetTransactionBlock(t *testing.T) {
	ctx := context.Background()
	client := clients.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := int(1)
	var coinsResp *iotajsonrpc.CoinPage
	var err error
	require.Eventually(t, func() bool {
		coinsResp, err = client.GetCoins(ctx, iotaclient.GetCoinsRequest{
			Owner: owner,
			Limit: limit,
		})
		return err == nil && len(coinsResp.Data) > 0
	}, 120*time.Second, 5*time.Second)

	digest := &coinsResp.Data[0].PreviousTransaction
	resp, err := client.GetTransactionBlock(ctx, iotaclient.GetTransactionBlockRequest{
		Digest: digest,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	fmt.Println("resp: ", resp)
}

func TestQueryTransactionBlocks(t *testing.T) {
	ctx := context.Background()
	client := clients.NewGraphQLClientWithTimeout(iotaconn.TestnetGraphQLEndpointURL, 60*time.Second)

	var resp *iotajsonrpc.TransactionBlocksPage
	var err error
	require.Eventually(t, func() bool {
		resp, err = client.QueryTransactionBlocks(ctx, iotaclient.QueryTransactionBlocksRequest{
			Limit: lo.ToPtr(int(3)),
		})
		return err == nil && len(resp.Data) > 0
	}, 3*time.Minute, 5*time.Second)
}

func TestTryGetPastObject(t *testing.T) {
	t.Skip("May fail")
	ctx := context.Background()
	client := clients.NewGraphQLClientWithTimeout(iotaconn.TestnetGraphQLEndpointURL, 120*time.Second)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := int(1)
	var coinsResp *iotajsonrpc.CoinPage
	var err error
	require.Eventually(t, func() bool {
		coinsResp, err = client.GetCoins(ctx, iotaclient.GetCoinsRequest{
			Owner: owner,
			Limit: limit,
		})
		return err == nil && len(coinsResp.Data) > 0
	}, 120*time.Second, 5*time.Second)

	coin := coinsResp.Data[0]
	version := coin.Version.Uint64()

	resp, err := client.TryGetPastObject(ctx, iotaclient.TryGetPastObjectRequest{
		ObjectID: coin.CoinObjectID,
		Version:  version,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
