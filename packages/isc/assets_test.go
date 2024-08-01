package isc_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iscmove"
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
		AddBaseTokens(big.NewInt(42)).
		AddCoin("0xa1::a::A", big.NewInt(100)).
		AddObject(sui.ObjectID{})
	rwutil.BytesTest(t, assets, isc.AssetsFromBytes)
}

func TestAssetsSpendBudget(t *testing.T) {
	toSpend := isc.NewEmptyAssets()
	budget := isc.NewEmptyAssets()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssetsBaseTokens(1)
	require.True(t, budget.Spend(toSpend))
	require.False(t, toSpend.Spend(budget))

	budget = isc.NewAssetsBaseTokens(10)
	require.True(t, budget.Spend(budget))
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssetsBaseTokens(2)
	toSpend = isc.NewAssetsBaseTokens(1)
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(isc.NewAssetsBaseTokens(1)))

	budget = isc.NewAssetsBaseTokens(1)
	toSpend = isc.NewAssetsBaseTokens(2)
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(isc.NewAssetsBaseTokens(1)))

	coinType1 := isc.CoinType("0xa1::a::A")
	coinType2 := isc.CoinType("0xa2::b::B")

	budget = isc.NewAssetsBaseTokens(1).AddCoin(coinType1, big.NewInt(5))
	toSpend = budget.Clone()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())

	budget = isc.NewAssetsBaseTokens(1).AddCoin(coinType1, big.NewInt(5))
	cloneBudget := budget.Clone()
	toSpend = isc.NewAssetsBaseTokens(1).AddCoin(coinType1, big.NewInt(10))
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = isc.NewAssetsBaseTokens(1).
		AddCoin(coinType1, big.NewInt(5)).
		AddCoin(coinType2, big.NewInt(1))
	toSpend = isc.NewAssetsBaseTokens(1).
		AddCoin(coinType1, big.NewInt(5))
	expected := isc.NewAssetsBaseTokens(0).
		AddCoin(coinType2, big.NewInt(1))
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(expected))

	budget = isc.NewAssetsBaseTokens(10).
		AddCoin(coinType2, big.NewInt(1))
	toSpend = isc.NewAssetsBaseTokens(1).
		AddCoin(coinType1, big.NewInt(5))
	require.False(t, budget.Spend(toSpend))
}
