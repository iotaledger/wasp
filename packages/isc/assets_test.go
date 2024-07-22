package isc_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

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

	assets := isc.NewAssets(1, tokens)
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
