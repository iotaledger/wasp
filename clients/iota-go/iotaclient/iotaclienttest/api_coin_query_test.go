package iotaclienttest

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/clients"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	testcommon "github.com/iotaledger/wasp/v2/clients/iota-go/test_common"
	"github.com/iotaledger/wasp/v2/packages/testutil/l1starter"
)

func TestGetAllBalances(t *testing.T) {
	api := l1starter.Instance().L1Client()
	balances, err := api.GetAllBalances(context.Background(), iotago.MustAddressFromHex(testcommon.TestAddress))
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
		address *iotago.Address
		cursor  *iotago.ObjectID
		limit   uint
	}

	tests := []struct {
		name    string
		a       clients.L1Client
		args    args
		want    *iotajsonrpc.CoinPage
		wantErr bool
	}{
		{
			name: "successful with limit",
			a:    l1starter.Instance().L1Client(),
			args: args{
				ctx:     context.Background(),
				address: iotago.MustAddressFromHex(testcommon.TestAddress),
				cursor:  nil,
				limit:   3,
			},
			wantErr: false,
		},
		{
			name: "successful without limit",
			a:    l1starter.Instance().L1Client(),
			args: args{
				ctx:     context.Background(),
				address: iotago.MustAddressFromHex(testcommon.TestAddress),
				cursor:  nil,
				limit:   0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				err := iotaclient.RequestFundsFromFaucet(tt.args.ctx, tt.args.address, l1starter.Instance().FaucetURL())
				require.NoError(t, err)

				got, err := tt.a.GetAllCoins(
					tt.args.ctx, iotaclient.GetAllCoinsRequest{
						Owner:  tt.args.address,
						Cursor: tt.args.cursor,
						Limit:  tt.args.limit,
					},
				)
				if (err != nil) != tt.wantErr {
					t.Errorf("GetAllCoins() error: %v, wantErr %v", err, tt.wantErr)
					return
				}
				// we have called multiple times RequestFundsFromFaucet() on testnet,
				// so the account have several IOTA objects.
				require.GreaterOrEqual(t, len(got.Data), int(tt.args.limit))
				require.NotNil(t, got.NextCursor)
			},
		)
	}
}

func TestGetBalance(t *testing.T) {
	api := l1starter.Instance().L1Client()
	err := iotaclient.RequestFundsFromFaucet(
		context.Background(),
		iotago.MustAddressFromHex(testcommon.TestAddress),
		l1starter.Instance().FaucetURL(),
	)
	require.NoError(t, err)

	balance, err := api.GetBalance(
		context.Background(),
		iotaclient.GetBalanceRequest{Owner: iotago.MustAddressFromHex(testcommon.TestAddress)},
	)
	require.NoError(t, err)
	t.Logf(
		"Coin Name: %v, Count: %v, Total: %v, Locked: %v",
		balance.CoinType, balance.CoinObjectCount,
		balance.TotalBalance, balance.LockedBalance,
	)
}

func TestGetCoinMetadata(t *testing.T) {
	api := l1starter.Instance().L1Client()
	metadata, err := api.GetCoinMetadata(context.Background(), iotajsonrpc.IotaCoinType.String())
	require.NoError(t, err)

	require.Equal(t, "IOTA", metadata.Name)
}

func TestGetCoins(t *testing.T) {
	api := l1starter.Instance().L1Client()
	address := iotago.MustAddressFromHex(testcommon.TestAddress)

	err := iotaclient.RequestFundsFromFaucet(context.Background(), address, l1starter.Instance().FaucetURL())
	require.NoError(t, err)

	defaultCoinType := iotajsonrpc.IotaCoinType.String()
	coins, err := api.GetCoins(
		context.Background(), iotaclient.GetCoinsRequest{
			Owner:    address,
			CoinType: &defaultCoinType,
			Limit:    3,
		},
	)
	require.NoError(t, err)

	require.Greater(t, len(coins.Data), 0)

	for _, data := range coins.Data {
		require.Equal(t, iotajsonrpc.IotaCoinType, data.CoinType)
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
		api     clients.L1Client
		args    args
		want    uint64
		wantErr bool
	}{
		{
			name: "get Iota supply",
			api:  l1starter.Instance().L1Client(),
			args: args{
				context.Background(),
				iotajsonrpc.IotaCoinType.String(),
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

				require.Truef(t, got.Value.Cmp(big.NewInt(0)) > 0, "IOTA supply should be greater than 0")
			},
		)
	}
}
