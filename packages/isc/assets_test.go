package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotago/iotatest"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

func TestAssetsNativeToken(t *testing.T) {
	a := isc.NewAssets(1234)
	a.AddCoin(coin.MustTypeFromString("0x3::testiota::TESTIOTA"), 4321)
	require.EqualValues(t, 1, a.Coins.NativeTokens().Size())
}

func TestAssetsBagWithBalancesToAssets(t *testing.T) {
	assetsBag := iscmove.AssetsBagWithBalances{
		AssetsBag: iscmove.AssetsBag{
			ID:   *iotago.MustAddressFromHex("0x123"),
			Size: 2,
		},
		Assets: *iscmove.NewAssets(33).
			SetCoin(iotajsonrpc.MustCoinTypeFromString("0xa1::a::A"), 11).
			SetCoin(iotajsonrpc.MustCoinTypeFromString("0xa2::b::B"), 22).
			AddObject(iotago.Address{1, 2, 3}, iotago.MustTypeFromString("0xa1::c::C")),
	}
	assets, err := isc.AssetsFromAssetsBagWithBalances(&assetsBag)
	require.NoError(t, err)
	require.Equal(t, assetsBag.Coins.Get(iotajsonrpc.IotaCoinType), iotajsonrpc.CoinValue(assets.BaseTokens()))
	require.Equal(t, assetsBag.Coins.Get(iotajsonrpc.MustCoinTypeFromString("0xa1::a::A")), iotajsonrpc.CoinValue(assets.CoinBalance(coin.MustTypeFromString("0xa1::a::A"))))
	require.Equal(t, assetsBag.Coins.Get(iotajsonrpc.MustCoinTypeFromString("0xa2::b::B")), iotajsonrpc.CoinValue(assets.CoinBalance(coin.MustTypeFromString("0xa2::b::B"))))
	require.Equal(t, assetsBag.Objects.MustGet(iotago.Address{1, 2, 3}), iotago.MustTypeFromString("0xa1::c::C"))
}

func TestAssetsSerialization(t *testing.T) {
	assets := isc.NewEmptyAssets().
		AddBaseTokens(42).
		AddCoin(coin.MustTypeFromString("0xa1::a::A"), 100).
		AddObject(isc.NewIotaObject(iotago.ObjectID{1, 2, 3}, iotago.MustTypeFromString("0xa1::c::C")))
	bcs.TestCodecAndHash(t, assets, "1d7bc26ebfeb")
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

	coinType1 := coin.MustTypeFromString("0xa1::a::A")
	coinType2 := coin.MustTypeFromString("0xa2::b::B")

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
		AddCoin(coin.MustTypeFromString("0xa1::a::A"), 100).
		AddObject(isc.NewIotaObject(*iotatest.RandomAddress(), iotago.MustTypeFromString("0xa1::c::C")))
	bcs.TestCodec(t, assets)

	assets = isc.NewEmptyAssets().
		AddBaseTokens(42).
		AddCoin(coin.MustTypeFromString("0xa1::a::A"), 100).
		AddObject(isc.NewIotaObject(*iotatest.TestAddress, iotago.MustTypeFromString("0xa1::c::C")))
	bcs.TestCodecAndHash(t, assets, "d005fba295b6")
}

func TestCoinBalancesCodec(t *testing.T) {
	coinBalance := isc.NewCoinBalances().
		Set(coin.MustTypeFromString("0xa1::a::A"), 100).
		Set(coin.MustTypeFromString("0xa2::b::B"), 200)
	bcs.TestCodecAndHash(t, coinBalance, "9d070cb05d31")
}
