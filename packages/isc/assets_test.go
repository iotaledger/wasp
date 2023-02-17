package isc_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/marshalutil"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util"
)

func TestMarshalling(t *testing.T) {
	maxVal, e := big.NewInt(2), big.NewInt(256)
	maxVal.Exp(maxVal, e, nil)
	maxVal.Sub(maxVal, big.NewInt(1))

	tokens := iotago.NativeTokens{
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{1},
			Amount: big.NewInt(100),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{2},
			Amount: big.NewInt(200),
		},
		&iotago.NativeToken{
			ID:     [iotago.NativeTokenIDLength]byte{3},
			Amount: util.MaxUint256,
		},
	}

	assets := isc.NewAssets(1, tokens)
	bytes := assets.Bytes()
	assets2, err := isc.AssetsFromMarshalUtil(marshalutil.New(bytes))
	require.NoError(t, err)
	require.Equal(t, assets.BaseTokens, assets2.BaseTokens)
	require.Equal(t, len(assets.NativeTokens), len(assets2.NativeTokens))
	for i := range tokens {
		require.Equal(t, assets.NativeTokens[i], assets2.NativeTokens[i])
	}
}

func TestAssets_SpendBudget(t *testing.T) {
	var toSpend *isc.Assets
	var budget *isc.Assets
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{BaseTokens: 1}
	require.True(t, budget.Spend(toSpend))
	require.False(t, toSpend.Spend(budget))

	budget = &isc.Assets{BaseTokens: 10}
	require.True(t, budget.Spend(budget))
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{BaseTokens: 2}
	toSpend = &isc.Assets{BaseTokens: 1}
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(&isc.Assets{
		BaseTokens:   1,
		NativeTokens: []*iotago.NativeToken{},
		NFTs:         []iotago.NFTID{},
	}))

	budget = &isc.Assets{BaseTokens: 1}
	toSpend = &isc.Assets{BaseTokens: 2}
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(&isc.Assets{
		BaseTokens:   1,
		NativeTokens: []*iotago.NativeToken{},
		NFTs:         []iotago.NFTID{},
	}))

	nativeTokenID1 := tpkg.RandNativeToken().ID
	nativeTokenID2 := tpkg.RandNativeToken().ID

	budget = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(5)},
		},
	}
	toSpend = budget.Clone()
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.IsEmpty())

	budget = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(5)},
		},
	}
	cloneBudget := budget.Clone()
	toSpend = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(10)},
		},
	}
	require.False(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(cloneBudget))

	budget = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(5)},
			{ID: nativeTokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(5)},
		},
	}
	expected := &isc.Assets{
		BaseTokens: 0,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID2, Amount: big.NewInt(1)},
		},
	}
	require.True(t, budget.Spend(toSpend))
	require.True(t, budget.Equals(expected))

	budget = &isc.Assets{
		BaseTokens: 10,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID2, Amount: big.NewInt(1)},
		},
	}
	toSpend = &isc.Assets{
		BaseTokens: 1,
		NativeTokens: iotago.NativeTokens{
			{ID: nativeTokenID1, Amount: big.NewInt(5)},
		},
	}

	require.False(t, budget.Spend(toSpend))
}

func TestAddNFTs(t *testing.T) {
	nftSet1 := []iotago.NFTID{
		{1},
		{2},
		{3},
	}

	nftSet2 := []iotago.NFTID{
		{3},
		{4},
		{5},
	}
	a := isc.NewAssets(0, nil, nftSet1...)
	b := isc.NewAssets(0, nil, nftSet2...)
	a.Add(b)
	require.Len(t, a.NFTs, 5)
	require.Contains(t, a.NFTs, iotago.NFTID{1})
	require.Contains(t, a.NFTs, iotago.NFTID{2})
	require.Contains(t, a.NFTs, iotago.NFTID{3})
	require.Contains(t, a.NFTs, iotago.NFTID{4})
	require.Contains(t, a.NFTs, iotago.NFTID{5})
}
