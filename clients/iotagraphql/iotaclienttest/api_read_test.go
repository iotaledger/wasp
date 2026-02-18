package iotaclienttest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetObject(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()
	owner := l1starter.ISCPackageOwner.Address()

	limit := int(1)
	coinsResp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Address.Coins.Nodes)

	coin := coinsResp.Address.Coins.Nodes[0]
	objResp, err := client.GetObject(ctx, coin.ObjectID())
	require.NoError(t, err)
	require.NotNil(t, objResp)
}

func TestGetTransactionBlock(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()
	owner := l1starter.ISCPackageOwner.Address()

	limit := int(1)
	coinsResp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Address.Coins.Nodes)

	// Get the coin's previous transaction via GetObject (GraphQL coins don't carry this directly)
	coinObj, err := client.GetObject(ctx, coinsResp.Address.Coins.Nodes[0].ObjectID())
	require.NoError(t, err)
	digest := *iotago.MustNewDigest(coinObj.Object.PreviousTransactionBlock.Digest)
	resp, err := client.GetTransactionBlock(ctx, digest)
	require.NoError(t, err)
	require.NotNil(t, resp)
	fmt.Println("resp: ", resp)
}

func TestQueryTransactionBlocks(t *testing.T) {
	t.Skip("QueryTransactionBlocks not on IotaClient interface")
}

func TestTryGetPastObject(t *testing.T) {
	t.Skip("May fail")
	ctx := context.Background()
	graphqlURL := l1starter.Instance().APIURL()
	faucetURL := l1starter.Instance().FaucetURL()
	client := iotagraphql.NewGraphQLClientWithTimeout(graphqlURL, faucetURL, 120*time.Second, nil)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := int(1)
	coinsResp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Address.Coins.Nodes)

	coin := coinsResp.Address.Coins.Nodes[0]
	resp, err := client.TryGetPastObject(ctx, coin.ObjectID(), coin.Version)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
