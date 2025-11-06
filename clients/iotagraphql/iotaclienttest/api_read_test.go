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

func TestGetChainIdentifier(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	chainID, err := client.GetChainIdentifier(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, chainID)
}

func TestGetCheckpoint(t *testing.T) {
	t.Skip("Requires proper checkpoint digest, not just sequence number")
}

func TestGetCheckpoints(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	ctx := context.Background()

	resp, err := client.GetCheckpoints(ctx, iotaclient.GetCheckpointsRequest{
		Limit: lo.ToPtr(uint64(2)),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
}

func TestGetLatestCheckpointSequenceNumber(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	seqNum, err := client.GetLatestCheckpointSequenceNumber(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, seqNum)
}

func TestGetObject(t *testing.T) {
	ctx := context.Background()
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	require.NoError(t, iotaclient.RequestFundsFromFaucet(ctx, owner, iotaconn.AlphanetFaucetURL))

	limit := uint(1)
	coinsResp, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Data)

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
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := uint(1)
	coinsResp, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Data)

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
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	resp, err := client.QueryTransactionBlocks(ctx, iotaclient.QueryTransactionBlocksRequest{
		Limit: lo.ToPtr(uint(3)),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
}

func TestTryGetPastObject(t *testing.T) {
	ctx := context.Background()
	client := clients.NewGraphQLClientWithTimeout(iotaconn.AlphanetGraphQLEndpointURL, 120*time.Second)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := uint(1)
	coinsResp, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Data)

	coin := coinsResp.Data[0]
	version := coin.Version.Uint64()

	resp, err := client.TryGetPastObject(ctx, iotaclient.TryGetPastObjectRequest{
		ObjectID: coin.CoinObjectID,
		Version:  version,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
