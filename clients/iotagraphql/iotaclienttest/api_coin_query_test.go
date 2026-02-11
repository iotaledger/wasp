package iotaclienttest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetAllBalances(t *testing.T) {
	client := l1starter.Instance().L1Client()
	owner := l1starter.ISCPackageOwner.Address()

	balances, err := client.GetAllBalances(context.Background(), owner)
	require.NoError(t, err)
	require.NotEmpty(t, balances)

	for _, balance := range balances {
		t.Logf(
			"Coin  Type: %s, Count: %s, Total Balance: %s",
			balance.CoinType,
			balance.CoinObjectCount,
			balance.TotalBalance.String(),
		)
	}
}

func TestGetAllCoins(t *testing.T) {
	owner := l1starter.ISCPackageOwner.Address()
	// Use longer timeout for slow network
	graphqlURL := l1starter.Instance().APIURL()
	faucetURL := l1starter.Instance().FaucetURL()
	client := iotagraphql.NewGraphQLClientWithTimeout(graphqlURL, faucetURL, 90*time.Second, nil)

	limit := int(3)
	respWithLimit, err := client.GetAllCoins(context.Background(), iotagraphql.GetAllCoinsRequest{
		Owner: owner,
		Limit: limit,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respWithLimit.Data)
	require.LessOrEqual(t, len(respWithLimit.Data), limit)
	require.NotNil(t, respWithLimit.NextCursor)

	respNoLimit, err := client.GetAllCoins(context.Background(), iotagraphql.GetAllCoinsRequest{
		Owner: owner,
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(respNoLimit.Data), len(respWithLimit.Data))
}

func TestGetBalance(t *testing.T) {
	ctx := context.Background()
	owner := l1starter.ISCPackageOwner.Address()

	client := l1starter.Instance().L1Client()

	balance, err := client.GetBalance(ctx, iotagraphql.GetBalanceRequest{Owner: owner})
	require.NoError(t, err)
	require.True(t, balance.TotalBalance.Clone().Sign() > 0)
}

func TestGetCoinMetadata(t *testing.T) {
	client := l1starter.Instance().L1Client()
	metadata, err := client.GetCoinMetadata(context.Background(), iotagraphql.IotaCoinType.String())
	require.NoError(t, err)
	require.Equal(t, "IOTA", metadata.Name)
}

func TestGetCoins(t *testing.T) {
	ctx := context.Background()
	owner := l1starter.ISCPackageOwner.Address()

	client := l1starter.Instance().L1Client()

	fetchCoinType := iotagraphql.IotaCoinType.String()
	limit := int(5)

	resp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
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
