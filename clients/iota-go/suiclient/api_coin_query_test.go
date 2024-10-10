package suiclient_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/sui"
	"github.com/iotaledger/wasp/clients/iota-go/suiclient"
	"github.com/iotaledger/wasp/clients/iota-go/suiconn"
	"github.com/iotaledger/wasp/clients/iota-go/suijsonrpc"
)

func TestGetAllBalances(t *testing.T) {
	api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	balances, err := api.GetAllBalances(context.TODO(), testAddress)
	require.NoError(t, err)
	for _, balance := range balances {
		t.Logf(
			"Coin Name: %v, Count: %v, Total: %v, Locked: %v",
			balance.CoinType, balance.CoinObjectCount,
			balance.TotalBalance, balance.LockedBalance,
		)
	}
}

func TestGetAllCoins(t *testing.T) {
	type args struct {
		ctx     context.Context
		address *sui.Address
		cursor  *sui.ObjectID
		limit   uint
	}

	tests := []struct {
		name    string
		a       *suiclient.Client
		args    args
		want    *suijsonrpc.CoinPage
		wantErr bool
	}{
		{
			name: "successful with limit",
			a:    suiclient.NewHTTP(suiconn.AlphanetEndpointURL),
			args: args{
				ctx:     context.TODO(),
				address: testAddress,
				cursor:  nil,
				limit:   3,
			},
			wantErr: false,
		},
		{
			name: "successful without limit",
			a:    suiclient.NewHTTP(suiconn.AlphanetEndpointURL),
			args: args{
				ctx:     context.TODO(),
				address: testAddress,
				cursor:  nil,
				limit:   0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.a.GetAllCoins(
					tt.args.ctx, suiclient.GetAllCoinsRequest{
						Owner:  tt.args.address,
						Cursor: tt.args.cursor,
						Limit:  tt.args.limit,
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetAllCoins() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				// we have called multiple times RequestFundsFromFaucet() on testnet, so the account have several SUI objects.
				require.GreaterOrEqual(t, len(got.Data), int(tt.args.limit))
				require.NotNil(t, got.NextCursor)
			},
		)
	}
}

func TestGetBalance(t *testing.T) {
	api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	balance, err := api.GetBalance(context.TODO(), suiclient.GetBalanceRequest{Owner: testAddress})
	require.NoError(t, err)
	t.Logf(
		"Coin Name: %v, Count: %v, Total: %v, Locked: %v",
		balance.CoinType, balance.CoinObjectCount,
		balance.TotalBalance, balance.LockedBalance,
	)
}

func TestGetCoinMetadata(t *testing.T) {
	api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	metadata, err := api.GetCoinMetadata(context.TODO(), suijsonrpc.IotaCoinType)
	require.NoError(t, err)

	testSuiMetadata := &suijsonrpc.SuiCoinMetadata{
		Decimals:    9,
		Description: "The main (gas)token of the IOTA Network.",
		IconUrl:     "https://iota.org/logo.png",
		Id:          sui.MustObjectIDFromHex("0x9a934a2644c4ca2decbe3d126d80720429c5e31896aa756765afa23ae2cb4b99"),
		Name:        "IOTA",
		Symbol:      "IOTA",
	}
	require.Equal(t, testSuiMetadata, metadata)
}

func TestGetCoins(t *testing.T) {
	api := suiclient.NewHTTP(suiconn.AlphanetEndpointURL)
	defaultCoinType := suijsonrpc.IotaCoinType
	coins, err := api.GetCoins(
		context.TODO(), suiclient.GetCoinsRequest{
			Owner:    testAddress,
			CoinType: &defaultCoinType,
			Limit:    3,
		},
	)
	require.NoError(t, err)

	require.Len(t, coins.Data, 3)
	for _, data := range coins.Data {
		require.Equal(t, suijsonrpc.IotaCoinType, data.CoinType)
		require.Greater(t, data.Balance.Int64(), int64(0))
	}
}

func TestGetTotalSupply(t *testing.T) {
	type args struct {
		ctx      context.Context
		coinType string
	}

	tests := []struct {
		name    string
		api     *suiclient.Client
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "get Sui supply",
			api:  suiclient.NewHTTP(suiconn.AlphanetEndpointURL),
			args: args{
				context.TODO(),
				suijsonrpc.IotaCoinType,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got, err := tt.api.GetTotalSupply(tt.args.ctx, tt.args.coinType)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetTotalSupply() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				targetSupply := &suijsonrpc.Supply{Value: suijsonrpc.NewBigInt(suijsonrpc.IotaCoinSupply)}
				require.Equal(t, targetSupply, got)
			},
		)
	}
}
