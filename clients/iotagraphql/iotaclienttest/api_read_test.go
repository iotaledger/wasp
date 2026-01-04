package iotaclienttest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
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
	require.NotEmpty(t, coinsResp.Data)

	coin := coinsResp.Data[0]
	objResp, err := client.GetObject(ctx, iotagraphql.GetObjectRequest{
		ObjectID: coin.CoinObjectID,
		Options: &iotagraphql.IotaObjectDataOptions{
			ShowContent: true,
			ShowType:    true,
		},
	})
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
	require.NotEmpty(t, coinsResp.Data)

	digest := &coinsResp.Data[0].PreviousTransaction
	resp, err := client.GetTransactionBlock(ctx, iotagraphql.GetTransactionBlockRequest{
		Digest: digest,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	fmt.Println("resp: ", resp)
}

func TestQueryTransactionBlocks(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()

	resp, err := client.QueryTransactionBlocks(ctx, iotagraphql.QueryTransactionBlocksRequest{
		Limit: lo.ToPtr(int(3)),
	})
	require.NoError(t, err)
	require.NotEmpty(t, resp.Data)
}

func TestTryGetPastObject(t *testing.T) {
	t.Skip("May fail")
	ctx := context.Background()
	client := iotagraphql.NewGraphQLClientWithTimeout(iotaconn.LocalnetGraphQLEndpointURL, 120*time.Second, nil)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	limit := int(1)
	coinsResp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, coinsResp.Data)

	coin := coinsResp.Data[0]
	version := coin.Version.Uint64()

	resp, err := client.TryGetPastObject(ctx, iotagraphql.TryGetPastObjectRequest{
		ObjectID: coin.CoinObjectID,
		Version:  version,
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
}
