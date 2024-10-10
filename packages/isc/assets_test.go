package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/suitest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAssetsBagWithBalancesToAssets(t *testing.T) {
	assetsBag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *iotago.MustAddressFromHex("0x123"),
			Size: 2,
		},
		Balances: iscmove.AssetsBagBalances{
			iotajsonrpc.IotaCoinType: &iotajsonrpc.Balance{TotalBalance: 33},
			"0xa1::a::A":             &iotajsonrpc.Balance{TotalBalance: 11},
			"0xa2::b::B":             &iotajsonrpc.Balance{TotalBalance: 22},
		},
	}
	assets := isc.AssetsFromAssetsBagWithBalances(assetsBag)
	require.Equal(t, assetsBag.Balances[iotajsonrpc.IotaCoinType].TotalBalance, uint64(assets.BaseTokens()))
	require.Equal(t, assetsBag.Balances["0xa1::a::A"].TotalBalance, uint64(assets.CoinBalance("0xa1::a::A")))
	require.Equal(t, assetsBag.Balances["0xa2::b::B"].TotalBalance, uint64(assets.CoinBalance("0xa2::b::B")))
}

func TestAssetsSerialization(t *testing.T) {
	assets := isc.NewEmptyAssets().
		AddBaseTokens(42).
		AddCoin("0xa1::a::A", 100).
		AddObject(iotago.ObjectID{})
	bcs.TestCodec(t, assets)
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

func TestAssetsCodec(t *testing.T) {
	assets := isc.NewEmptyAssets().
		AddBaseTokens(42).
		AddCoin("0xa1::a::A", 100).
		AddObject(*suitest.RandomAddress())
	bcs.TestCodec(t, assets)
}

func TestCoinBalancesCodec(t *testing.T) {
	coinBalance := isc.CoinBalances{
		"0xa1::a::A": 100,
		"0xa2::b::B": 200,
	}
	bcs.TestCodec(t, coinBalance)
}
