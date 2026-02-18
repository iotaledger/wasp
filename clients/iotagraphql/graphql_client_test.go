package iotagraphql_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Khan/genqlient/graphql"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaconn"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGraphQL(t *testing.T) {
	client := iotagraphql.NewGraphQLClient(iotaconn.TestnetGraphQLEndpointURL, iotaconn.TestnetFaucetURL)

	t.Run("Standard API Call", func(t *testing.T) {
		addr, err := iotago.AddressFromHex("0x7a89979774c55814f41fc1e3354e2ba38d3d62096d469d86b3132e947de1e8da")
		require.NoError(t, err)
		resp, err := client.GetAllBalances(context.TODO(), *addr)
		require.NoError(t, err)
		require.NotNil(t, resp)

		fmt.Println("All Balances:", resp)
	})

	t.Run("Custom query", func(t *testing.T) {
		q := `
query GetAllBalances($owner: SuiAddress!, $limit: Int, $cursor: String) {
	address(address: $owner) {
		balances(first: $limit, after: $cursor) {
			pageInfo {
				hasNextPage
				endCursor
			}
			nodes {
				coinType {
					repr
				}
				coinObjectCount
				totalBalance
			}
		}
	}
}`
		b, err := client.Query(context.Background(), q, map[string]interface{}{
			"owner": "0xe25afa59deccfec819aaa67bf14f049982d2a1ca87c49c8614da5ea2dc438f72",
		})
		require.NoError(t, err)
		fmt.Println("raw bytes:", string(b))

		var resp graphql.Response
		err = json.Unmarshal(b, &resp)
		require.NoError(t, err)
		fmt.Println("unmarshalled:", resp)
	})
}

func TestMain(m *testing.M) {
	l1starter.TestMain(m)
}

func TestFaucetReturns5CoinsWithCorrectAmount(t *testing.T) {
	ctx := context.Background()
	client := l1starter.Instance().L1Client()

	keyPair := cryptolib.NewKeyPair()
	addr := keyPair.Address().AsIotaAddress()

	err := client.RequestFundsFromFaucet(ctx, *addr)
	require.NoError(t, err)

	coinsResp, err := client.GetCoins(ctx, iotagraphql.GetCoinsRequest{
		Owner: addr,
		Limit: 10,
	})
	require.NoError(t, err)

	coins := coinsResp.Address.Coins.Nodes
	require.Len(t, coins, 5, "faucet should return exactly 5 coins per request")

	for i, coin := range coins {
		require.Equal(t,
			iotagraphql.SingleCoinFundsFromFaucetAmount,
			coin.Balance(),
			"coin %d should have SingleCoinFundsFromFaucetAmount (%d), got %d",
			i, iotagraphql.SingleCoinFundsFromFaucetAmount, coin.Balance(),
		)
	}

	balance, err := client.GetBalance(ctx, iotagraphql.GetBalanceRequest{Owner: addr})
	require.NoError(t, err)
	require.Equal(t,
		iotagraphql.FundsFromFaucetAmount,
		balance.TotalBalance.Uint64(),
		"total balance should equal FundsFromFaucetAmount (%d), got %d",
		iotagraphql.FundsFromFaucetAmount, balance.TotalBalance.Uint64(),
	)
}
