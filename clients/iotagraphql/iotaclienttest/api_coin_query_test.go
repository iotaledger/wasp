package iotaclienttest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
)

func TestGetAllBalances(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	balances, err := client.GetAllBalances(context.Background(), owner)
	require.NoError(t, err)
	require.NotEmpty(t, balances)

	for _, balance := range balances {
		t.Logf(
			"Coin Type: %s, Count: %s, Total Balance: %s",
			balance.CoinType,
			balance.CoinObjectCount,
			balance.TotalBalance.String(),
		)
	}
}

func TestGetAllCoins(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	owner := iotago.MustAddressFromHex(testcommon.TestAddress)
	faucetURL := iotaconn.AlphanetFaucetURL

	require.NoError(t, iotaclient.RequestFundsFromFaucet(ctx, owner, faucetURL))

	// Use longer timeout for slow network
	client := clients.NewGraphQLClientWithTimeout(iotaconn.AlphanetGraphQLEndpointURL, 90*time.Second)

	limit := uint(3)
	respWithLimit, err := client.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respWithLimit.Data)
	require.LessOrEqual(t, len(respWithLimit.Data), int(limit))
	require.NotNil(t, respWithLimit.NextCursor)

	respNoLimit, err := client.GetAllCoins(ctx, iotaclient.GetAllCoinsRequest{
		Owner: owner,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(respNoLimit.Data), len(respWithLimit.Data))
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	require.NoError(t, iotaclient.RequestFundsFromFaucet(ctx, owner, iotaconn.AlphanetFaucetURL))

	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	balance, err := client.GetBalance(ctx, iotaclient.GetBalanceRequest{Owner: owner})
	require.NoError(t, err)
	require.True(t, balance.TotalBalance.Clone().Sign() > 0)
}

func TestGetCoinMetadata(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	metadata, err := client.GetCoinMetadata(context.Background(), iotajsonrpc.IotaCoinType.String())
	require.NoError(t, err)
	require.Equal(t, "IOTA", metadata.Name)
}

func TestGetCoins(t *testing.T) {
	ctx := context.Background()
	owner := iotago.MustAddressFromHex(testcommon.TestAddress)

	require.NoError(t, iotaclient.RequestFundsFromFaucet(ctx, owner, iotaconn.AlphanetFaucetURL))

	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)

	fetchCoinType := iotajsonrpc.IotaCoinType.String()
	limit := uint(5)

	resp, err := client.GetCoins(ctx, iotaclient.GetCoinsRequest{
		Owner:    owner,
		Limit:    limit,
		CoinType: &fetchCoinType,
	})
	require.NoError(t, err)

	coins := resp.Data
	require.NotEmpty(t, coins)

	for _, coin := range coins {
		wrappedCoinType := fmt.Sprintf("0x0000000000000000000000000000000000000000000000000000000000000002::coin::Coin<%s>", fetchCoinType)
		require.Equal(t, wrappedCoinType, coin.CoinType.String())
		require.True(t, coin.Balance.Clone().Sign() > 0)
	}
}

func TestGetTotalSupply(t *testing.T) {
	client := clients.NewGraphQLClient(iotaconn.AlphanetGraphQLEndpointURL)
	supply, err := client.GetTotalSupply(context.Background(), iotajsonrpc.IotaCoinType.String())
	require.NoError(t, err)
	require.Truef(t, supply.Value.Clone().Sign() > 0, "total supply should be greater than zero")
}
