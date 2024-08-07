package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestAssetsBagWithBalancesToAssets(t *testing.T) {
	assetsBag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *sui.MustAddressFromHex("0x123"),
			Size: 2,
		},
		Balances: iscmove.AssetsBagBalances{
			suijsonrpc.SuiCoinType: &suijsonrpc.Balance{TotalBalance: suijsonrpc.NewBigInt(33)},
			"0xa1::a::A":           &suijsonrpc.Balance{TotalBalance: suijsonrpc.NewBigInt(11)},
			"0xa2::b::B":           &suijsonrpc.Balance{TotalBalance: suijsonrpc.NewBigInt(22)},
		},
	}
	assets := isc.AssetsFromAssetsBag(assetsBag)
	require.Equal(t, assetsBag.Balances[suijsonrpc.SuiCoinType].TotalBalance.Int, assets.BaseTokens())
	require.Equal(t, assetsBag.Balances["0xa1::a::A"].TotalBalance.Int, assets.CoinBalance("0xa1::a::A"))
	require.Equal(t, assetsBag.Balances["0xa2::b::B"].TotalBalance.Int, assets.CoinBalance("0xa2::b::B"))
}

func TestAssetsSerialization(t *testing.T) {
	assets := isc.NewEmptyAssets().
		AddBaseTokens(42).
		AddCoin("0xa1::a::A", 100).
		AddObject(sui.ObjectID{})
	rwutil.BytesTest(t, assets, isc.AssetsFromBytes)
}

func TestAssetsSpendBudget(t *testing.T) {
	toSpend := isc.NewEmptyAssets()
	budget := isc.NewEmptyAssets()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssets(1)
	require.True(t, budget.Spend(toSpend))
	require.False(t, toSpend.Spend(budget))

	budget = isc.NewAssets(10)
	require.True(t, budget.Spend(budget))
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssets(2)
	toSpend = isc.NewAssets(1)
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(isc.NewAssets(1)))

	budget = isc.NewAssets(1)
	toSpend = isc.NewAssets(2)
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(isc.NewAssets(1)))

	coinType1 := coin.Type("0xa1::a::A")
	coinType2 := coin.Type("0xa2::b::B")

	budget = isc.NewAssets(1).AddCoin(coinType1, 5)
	toSpend = budget.Clone()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssets(1).AddCoin(coinType1, 5)
	cloneBudget := budget.Clone()
	toSpend = isc.NewAssets(1).AddCoin(coinType1, 10)
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = isc.NewAssets(1).
		AddCoin(coinType1, 5).
		AddCoin(coinType2, 1)
	toSpend = isc.NewAssets(1).
		AddCoin(coinType1, 5)
	expected := isc.NewAssets(0).
		AddCoin(coinType2, 1)
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(expected))

	budget = isc.NewAssets(10).
		AddCoin(coinType2, 1)
	toSpend = isc.NewAssets(1).
		AddCoin(coinType1, 5)
	require.False(t, budget.Spend(toSpend))
}
