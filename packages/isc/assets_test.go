package isc_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/clients/iscmove/isctypes"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

func TestAssetsBagWithBalancesToAssets(t *testing.T) {
	assetsBag := isctypes.AssetsBagWithBalances{
		AssetsBag: isctypes.AssetsBag{
			ID:   *sui.MustAddressFromHex("0x123"),
			Size: 2,
		},
		Balances: isctypes.AssetsBagBalances{
			suijsonrpc.SuiCoinType: &suijsonrpc.Balance{TotalBalance: &suijsonrpc.BigInt{big.NewInt(33)}},
			"0xa1":                 &suijsonrpc.Balance{TotalBalance: &suijsonrpc.BigInt{big.NewInt(11)}},
			"0xa2":                 &suijsonrpc.Balance{TotalBalance: &suijsonrpc.BigInt{big.NewInt(22)}},
		},
	}
	assets := isc.AssetsFromAssetsBag(assetsBag)
	require.Equal(t, assetsBag.Balances[suijsonrpc.SuiCoinType].TotalBalance.Int, assets.BaseTokens)
	require.Equal(t, assetsBag.Balances["0xa1"].TotalBalance.Int, assets.NativeTokens.MustSet()["0xa1"].Amount)
	require.Equal(t, assetsBag.Balances["0xa2"].TotalBalance.Int, assets.NativeTokens.MustSet()["0xa2"].Amount)
}

func TestAssetsSerialization(t *testing.T) {
	maxVal, e := big.NewInt(2), big.NewInt(256)
	maxVal.Exp(maxVal, e, nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	tokens := isc.NativeTokens{
		&isc.NativeToken{
			CoinType: isc.NativeTokenID("0xabc"),
			Amount:   big.NewInt(100),
		},
		&isc.NativeToken{
			CoinType: isc.NativeTokenID("0xdef"),
			Amount:   big.NewInt(200),
		},
		&isc.NativeToken{
			CoinType: isc.NativeTokenID("0xghe"),
			Amount:   util.MaxUint256,
		},
	}

	assets := isc.NewAssets(big.NewInt(1), tokens)
	rwutil.BytesTest(t, assets, isc.AssetsFromBytes)
}

func TestAssetsSpendBudget(t *testing.T) {
	var toSpend *isc.Assets
	var budget *isc.Assets
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{BaseTokens: big.NewInt(1)}
	require.True(t, budget.Spend(toSpend))
	require.False(t, toSpend.Spend(budget))

	budget = &isc.Assets{BaseTokens: big.NewInt(10)}
	require.True(t, budget.Spend(budget))
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{BaseTokens: big.NewInt(2)}
	toSpend = &isc.Assets{BaseTokens: big.NewInt(1)}
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(&isc.Assets{
		BaseTokens:   big.NewInt(1),
		NativeTokens: []*isc.NativeToken{},
	}))

	budget = &isc.Assets{BaseTokens: big.NewInt(1)}
	toSpend = &isc.Assets{BaseTokens: big.NewInt(2)}
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(&isc.Assets{
		BaseTokens:   big.NewInt(1),
		NativeTokens: []*isc.NativeToken{},
	}))

	nativeTokenID1 := tpkg.RandNativeToken().ID
	nativeTokenID2 := tpkg.RandNativeToken().ID

	budget = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(5)},
		},
	}
	toSpend = budget.Clone()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(5)},
		},
	}
	cloneBudget := budget.Clone()
	toSpend = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(10)},
		},
	}
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(5)},
			{CoinType: isc.NativeTokenID(nativeTokenID2.String()), Amount: big.NewInt(1)},
		},
	}
	toSpend = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(5)},
		},
	}
	expected := &isc.Assets{
		BaseTokens: big.NewInt(0),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID2.String()), Amount: big.NewInt(1)},
		},
	}
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(expected))

	budget = &isc.Assets{
		BaseTokens: big.NewInt(10),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID2.String()), Amount: big.NewInt(1)},
		},
	}
	toSpend = &isc.Assets{
		BaseTokens: big.NewInt(1),
		NativeTokens: isc.NativeTokens{
			{CoinType: isc.NativeTokenID(nativeTokenID1.String()), Amount: big.NewInt(5)},
		},
	}

	require.False(t, budget.Spend(toSpend))
}
